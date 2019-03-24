package handler

import (
	"math"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/gomodule/redigo/redis"
	"github.com/washingt0/cashbot/database"
	"github.com/washingt0/cashbot/report"
	"go.mongodb.org/mongo-driver/mongo"
)

/*
	STATES
	0 or nil - First contact
	1 - Bot is waiting for the entry data
	2 - Bot is waiting for the tag data
	3 - Bot is waiting for the tag to add to a entry
*/
/*
	REPLY
	0 - None
	1 - Markdown
	2 - File
	3 - Keyboard
*/

func IncomingMessageHandler(m *tgbotapi.Message, mgo *mongo.Client) (tgbotapi.MessageConfig, bool, error) {
	e := database.Entry{}

	e.Owner = m.Chat.UserName
	message, reply, err := parseMessage(m.Text, mgo, &e)

	file := false
	msg := tgbotapi.NewMessage(m.Chat.ID, message)

	if reply == 1 {
		msg.ParseMode = tgbotapi.ModeMarkdown
	} else if reply == 2 {
		file = true
	} else if reply == 3 {
		var t database.Tag
		t.Owner = e.Owner
		if data, err := t.GetAllTags(mgo); err != nil {
			return msg, false, err
		} else {
			msg.ReplyMarkup = makeReplyKeyboard(data)
		}
	} else if reply == -3 {
		msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	}

	return msg, file, err
}

func parseMessage(text string, mgo *mongo.Client, e *database.Entry) (string, int, error) {

	if text == "/start" {
		return "Send the /addentry command to insert a new entry with this format `<VALUE> <DESCRIPTION>` or /help", 0, nil
	} else if text == "/help" {
		return `
/start - Start a conversation with the bot
/addentry - Prepare the bot to receive a entry information
/getreport - Ask the bot to produce a report from your information
/getdayreport - Ask the bot to produce a report with the day information
/getmonthreport - Ask the bot to produce a report with the month information
/getpdfreport - Ask the bot to produce a PDF report from your information
/removelast - Ask the bot to remove the last entry
/addtag - Ask the bot to receive a tag name
/listtag - Ask the bot to list your tags
/clear - Ask the bot to clear all your data
/help - Display this help message
`, 0, nil
		/* === BEGIN REPORTS === */
	} else if text == "/getreport" {
		if data, err := e.GetOwnerEntries(mgo, nil, nil); err != nil {
			return "", 0, err
		} else {
			return makeTable(data), 1, nil
		}
	} else if text == "/getpdfreport" {
		if data, err := e.GetOwnerEntries(mgo, nil, nil); err != nil {
			return "", 0, err
		} else {
			return report.GeneratePDF(data, e.Owner), 2, nil
		}
	} else if text == "/getdayreport" {
		now := time.Now()
		start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		end := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
		if data, err := e.GetOwnerEntries(mgo, &start, &end); err != nil {
			return "", 0, err
		} else {
			return makeTable(data), 1, nil
		}
	} else if text == "/getmonthreport" {
		now := time.Now()
		start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		end := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, now.Location())
		if data, err := e.GetOwnerEntries(mgo, &start, &end); err != nil {
			return "", 0, err
		} else {
			return makeTable(data), 1, nil
		}
		/* === END REPORTS === */
		/* === BEGIN ENTRIES === */
	} else if text == "/addentry" {
		if err := setState(e.Owner, 1); err != nil {
			return "", 0, err
		}
		return "Send the entry information, something like this: `14.00 pot`. If the value was less than zero, I will mark it as a payment", 1, nil
	} else if text == "/removelast" {
		if err := e.DropLastEntry(mgo); err != nil {
			return "", 0, err
		}
		return "Okay, it's gone. ;)", 0, nil
	} else if text == "/clear" {
		if err := e.DropEntries(mgo); err != nil {
			return "", 0, err
		}
		return "Okay, everything is gone. ;)", 0, nil
		/* === END ENTRIES === */
		/* === BEGIN TAGS === */
	} else if text == "/addtag" {
		if err := setState(e.Owner, 2); err != nil {
			return "", 0, err
		}
		return "Send the tag name", 0, nil
	} else if text == "/listtag" {
		t := database.Tag{Owner: e.Owner}
		if data, err := t.GetAllTags(mgo); err != nil {
			return "", 0, err
		} else {
			return makeTagTable(data), 1, nil
		}
	} else if text == "/deletetag" {
		/* === END TAGS === */
	} else {
		state, err := getState(e.Owner)
		if err != nil {
			return "", 0, err
		}
		if state == 1 {
			if err := parseEntryText(text, e); err != nil {
				return "", 0, err
			} else {
				if err := e.AddEntry(mgo); err != nil {
					return "", 0, err
				}
				if err := setState(e.Owner, 3); err != nil {
					return "", 0, err
				}
				return "Ok, I will remember this.", 3, nil
			}
		} else if state == 2 {
			var t database.Tag
			t.Name = text
			t.Owner = e.Owner
			t.CreatedAt = time.Now()
			if err := t.AddTag(mgo); err != nil {
				return "", 0, err
			}
			setState(e.Owner, 0)
			return "Ok, you can use this now.", 0, nil
		} else if state == 3 {
			if text == "Done" {
				if err := setState(e.Owner, 0); err != nil {
					return "", 0, err
				}
				return "Ok, done!", -3, nil
			} else {
				if err := e.AddTagToEntry(mgo, text); err != nil {
					return "", 0, err
				}
				return "Ok, added!", 3, nil
			}
		}
	}

	return "Oops, I did not understand you :'( ", 0, nil
}

func parseEntryText(text string, e *database.Entry) error {
	parts := strings.Split(text, " ")
	if len(parts) >= 1 {
		val, err := strconv.ParseFloat(parts[0], 64)
		if err != nil {
			return err
		}
		if val < 0.0 {
			e.Payment = true
		}
		e.Value = math.Abs(val)

		if len(parts) > 1 {
			e.Description = strings.Join(parts[1:], " ")
		}
		e.CreatedAt = time.Now()
		e.Tags = []string{}
	}
	return nil
}

func setState(user string, state int) error {
	conn, err := database.ConnectRedis()
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.Do("SET", user, state)
	if err != nil {
		return err
	}

	return nil
}

func getState(user string) (int, error) {
	conn, err := database.ConnectRedis()
	if err != nil {
		return -1, err
	}
	defer conn.Close()

	data, err := redis.Int(conn.Do("GET", user))
	if err != nil {
		return -1, err
	}

	return data, nil
}

func makeTable(data []database.Entry) string {
	out := ""
	tin := 0.0
	tou := 0.0
	for _, e := range data {
		out += e.CreatedAt.Format("2006/01/02 15:04")
		out += " | "
		if e.Payment {
			out += "-"
			tou += e.Value
		} else {
			out += "+"
			tin += e.Value
		}
		out += strconv.FormatFloat(e.Value, 'f', 2, 64)
		out += " | "
		out += e.Description
		out += `
`
	}
	out += `-----------------------------------
`
	out += "Total in: " + strconv.FormatFloat(tin, 'f', 2, 64) + `
`
	out += "Total out: " + strconv.FormatFloat(tou, 'f', 2, 64) + `
`
	out += "Balance: " + strconv.FormatFloat(tin-tou, 'f', 2, 64) + `
`
	return out
}

func makeReplyKeyboard(data []database.Tag) tgbotapi.ReplyKeyboardMarkup {
	var keys []tgbotapi.KeyboardButton
	for _, t := range data {
		keys = append(keys, tgbotapi.NewKeyboardButton(t.Name))
	}

	keys = append(keys, tgbotapi.NewKeyboardButton("Done"))

	var keyboard tgbotapi.ReplyKeyboardMarkup

	keyboard.ResizeKeyboard = true
	keyboard.OneTimeKeyboard = true
	keyboard.Selective = true

	for i := 0; i < (len(keys)/3)+1; i++ {
		row := []tgbotapi.KeyboardButton{}
		for j := 0; j < 3; j++ {
			if (i*3)+j >= len(keys) {
				continue
			} else {
				row = append(row, keys[(i*3)+j])
			}
		}
		keyboard.Keyboard = append(keyboard.Keyboard, row)
	}

	return keyboard
}

func makeTagTable(data []database.Tag) string {
	out := `Your Tags
`
	for _, e := range data {
		out += e.CreatedAt.Format("2006/01/02 15:04")
		out += " | "
		out += e.Name
		out += `
`
	}
	return out
}
