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
*/

func IncomingMessageHandler(m *tgbotapi.Message, mgo *mongo.Client) (tgbotapi.MessageConfig, bool, error) {
	e := database.Entry{}

	e.Owner = m.Chat.UserName
	message, md, file, err := parseMessage(m.Text, mgo, &e)

	msg := tgbotapi.NewMessage(m.Chat.ID, message)
	if md {
		msg.ParseMode = tgbotapi.ModeMarkdown
	}

	return msg, file, err
}

func parseMessage(text string, mgo *mongo.Client, e *database.Entry) (string, bool, bool, error) {

	if text == "/start" || text == "/help" {
		return "Send the `/addentry` command to insert a new entry with this format `<VALUE> <DESCRIPTION>`", true, false, nil
	} else if text == "/addentry" {
		if err := setState(e.Owner, 1); err != nil {
			return "", false, false, err
		}
		return "Send the entry information, something like this: `14.00 pot`. If the value was less than zero, I will mark it as a payment", true, false, nil
	} else if text == "/getreport" {
		if data, err := e.GetOwnerEntries(mgo); err != nil {
			return "", false, false, err
		} else {
			return makeTable(data), true, false, nil
		}
	} else if text == "/getpdfreport" {
		if data, err := e.GetOwnerEntries(mgo); err != nil {
			return "", false, false, err
		} else {
			return report.GeneratePDF(data, e.Owner), false, true, nil
		}
	} else if text == "/getmonthreport" {

	} else if text == "/getweekreport" {

	} else {
		state, err := getState(e.Owner)
		if err != nil {
			return "", false, false, err
		}
		if state == 1 {
			if err := parseEntryText(text, e); err != nil {
				return "", false, false, err
			} else {
				if err := e.AddEntry(mgo); err != nil {
					return "", false, false, err
				}
				if err := setState(e.Owner, 0); err != nil {
					return "", false, false, err
				}
				return "Ok, I will remember this.", false, false, nil
			}
		} else if state == 2 {
		}
	}

	return "Oops, I did not understand you :'( ", false, false, nil
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
