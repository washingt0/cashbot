package report

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/jung-kurt/gofpdf"
	"github.com/washingt0/cashbot/database"
)

const (
	MARGIN = 10
)

func GeneratePDF(data []database.Entry, user string) string {
	filename := filepath.Join(os.TempDir(), strconv.FormatInt(time.Now().UnixNano(), 10)+"_report_"+user+".pdf")
	var err error

	pdf := gofpdf.New("P", "mm", "A4", "")
	tr := pdf.UnicodeTranslatorFromDescriptor("")
	title := tr("Money Report of " + user)

	pdf.SetHeaderFunc(func() {
		pdf.SetFont("Arial", "B", 14)
		wd := pdf.GetStringWidth(title) + 6
		p, _ := pdf.GetPageSize()
		pdf.SetX(p - wd - MARGIN)
		pdf.SetTextColor(0, 0, 0)
		pdf.CellFormat(wd, 9, title, "0", 1, "C", false, 0, "")
		pdf.Ln(8)
	})

	pdf.SetFooterFunc(func() {
		pdf.SetY(-15)
		pdf.SetX(MARGIN)
		pdf.SetFont("Arial", "I", 8)
		pw, _ := pdf.GetPageSize()
		pw -= 2 * MARGIN
		cellW := pw / 3
		pdf.CellFormat(cellW, 10, title, "", 0, "L", false, 0, "")
		pdf.CellFormat(cellW, 10, tr(fmt.Sprintf("Page %d/{nb}", pdf.PageNo())), "", 0, "C", false, 0, "")
		pdf.CellFormat(cellW, 10, time.Now().Format("2006/01/02 15:04"), "", 0, "R", false, 0, "")
	})

	pdf.AliasNbPages("")

	pdf.AddPage()
	pdf.SetFont("Helvetica", "B", 8)
	pdf.SetHeaderFunc(nil)

	if len(data) == 0 {
		pdf.SetFont("Arial", "B", 16)
		pdf.Text(30, 110, tr("There is no data :'( "))
		pdf.OutputFileAndClose(filename)
		return filename
	}

	pdf.SetFillColor(225, 225, 225)
	pdf.SetTextColor(0, 0, 0)
	pdf.SetX(MARGIN)
	w, _ := pdf.GetPageSize()
	w -= 2 * MARGIN
	for _, i := range []string{"Date", "Value", "Description", "Tags"} {
		pdf.CellFormat(w/4.0, 7, tr(i), "1", 0, "", true, 0, "")
	}
	pdf.Ln(-1)
	pdf.SetTextColor(0, 0, 0)
	pdf.SetFillColor(225, 225, 225)
	fill := true

	pdf.SetFont("Helvetica", "", 8)
	pdf.SetFillColor(255, 255, 255)

	tin := 0.0
	tou := 0.0

	for _, x := range data {
		var out string
		if x.Payment {
			out = "-"
			tou += x.Value
			pdf.SetFillColor(229, 126, 105)
		} else {
			out = "+"
			tin += x.Value
			pdf.SetFillColor(186, 247, 165)
		}
		pdf.SetX(MARGIN)
		pdf.CellFormat(w/4.0, 6, tr(x.CreatedAt.Format("2006/01/02 15:04")), "1", 0, "", true, 0, "")
		pdf.CellFormat(w/4.0, 6, tr(out+strconv.FormatFloat(x.Value, 'f', 2, 64)), "1", 0, "", true, 0, "")
		pdf.CellFormat(w/4.0, 6, tr(x.Description), "1", 0, "", true, 0, "")
		pdf.CellFormat(w/4.0, 6, tr(fmt.Sprintf("%+v", x.Tags)), "1", 0, "", true, 0, "")
		pdf.Ln(-1)
		fill = !fill
	}

	pdf.SetFillColor(225, 225, 225)
	pdf.SetFont("Helvetica", "B", 10)
	pdf.SetTextColor(46, 81, 35)
	pdf.CellFormat(w/3.0, 7, tr("Total in: ")+strconv.FormatFloat(tin, 'f', 2, 64), "1", 0, "R", true, 0, "")
	pdf.SetTextColor(140, 37, 11)
	pdf.CellFormat(w/3.0, 7, tr("Total out: ")+strconv.FormatFloat(tou, 'f', 2, 64), "1", 0, "R", true, 0, "")
	if (tin - tou) >= 0 {
		pdf.SetTextColor(46, 81, 35)
	} else {
		pdf.SetTextColor(140, 37, 11)
	}
	pdf.CellFormat(w/3.0, 7, tr("Balance: ")+strconv.FormatFloat(tin-tou, 'f', 2, 64), "1", 0, "R", true, 0, "")

	err = pdf.OutputFileAndClose(filename)
	if err != nil {
		log.Fatal(err)
	}

	return filename
}
