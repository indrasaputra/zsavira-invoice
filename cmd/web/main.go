package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/SebastiaanKlippert/go-wkhtmltopdf"
	"github.com/labstack/echo/v4"
)

// Item ...
type Item struct {
	Number      int
	Description string
	Quantity    int
	UnitPrice   string
	TotalPrice  string
}

// Recipient ...
type Recipient struct {
	Name       string
	EventDate  string
	EventPlace string
}

// Invoice ...
type Invoice struct {
	Number     string
	Recipient  *Recipient
	Date       string
	Items      []*Item
	GrandTotal string
}

// TemplateRenderer is a custom html/template renderer for Echo framework
type TemplateRenderer struct {
	templates *template.Template
}

// Render renders a template document
func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	// Add global methods if data is a map
	if viewContext, isMap := data.(map[string]interface{}); isMap {
		viewContext["reverse"] = c.Echo().Reverse
	}

	return t.templates.ExecuteTemplate(w, name, data)
}

func main() {
	e := echo.New()
	renderer := &TemplateRenderer{
		templates: template.Must(template.ParseGlob("view/*.html")),
	}
	e.Renderer = renderer

	e.GET("/invoices", func(c echo.Context) error {
		return c.Render(http.StatusOK, "invoice_form.html", nil)
	})
	e.POST("/invoices", createInvoice)

	e.Logger.Fatal(e.Start(":8000"))
}

func createInvoice(c echo.Context) error {
	invDate := c.FormValue("invoice-date")
	if invDate == "" {
		invDate = toDate(time.Now())
	}
	invClient := c.FormValue("invoice-client")
	invEventDate := c.FormValue("invoice-event-date")
	if invEventDate == "" {
		invEventDate = toDate(time.Now())
	}
	invEventPlace := c.FormValue("invoice-event-place")
	invDetails := c.FormValue("invoice-details")

	items, grand := convertDetailsToItemList(invDetails)

	inv := &Invoice{
		Number: generateRandomNumber(),
		Date:   strings.Title(invDate),
		Recipient: &Recipient{
			Name:       strings.Title(invClient),
			EventDate:  strings.Title(invEventDate),
			EventPlace: strings.Title(invEventPlace),
		},
		Items:      items,
		GrandTotal: toCurrency(grand),
	}

	generateHtml(inv)
	createPdf()
	return c.File("view/result.pdf")
}

func generateHtml(inv *Invoice) {
	t, _ := template.ParseFiles("view/invoice.html")
	f, err := os.Create("view/result.html")
	if err != nil {
		log.Println("create file: ", err)
		return
	}
	t.Execute(f, inv)
	f.Close()
}

func createPdf() {
	pdfg, err := wkhtmltopdf.NewPDFGenerator()
	if err != nil {
		log.Fatal(err)
	}

	f, err := os.Open("view/result.html")
	if f != nil {
		defer f.Close()
	}
	if err != nil {
		log.Fatal(err)
	}

	pdfg.AddPage(wkhtmltopdf.NewPageReader(f))

	pdfg.Orientation.Set(wkhtmltopdf.OrientationPortrait)
	pdfg.Dpi.Set(300)

	err = pdfg.Create()
	if err != nil {
		log.Fatal(err)
	}

	err = pdfg.WriteFile("view/result.pdf")
	if err != nil {
		log.Fatal(err)
	}
}

func generateRandomNumber() string {
	m := time.Now().Month()
	f := time.Now().Second() % 10
	s := time.Now().Second() % 10

	if f == 0 && s == 0 {
		s = 1
	}
	return fmt.Sprintf("INV/%d/0%d%d", m, f, s)
}

func convertDetailsToItemList(invDetails string) ([]*Item, int) {
	items := []*Item{}
	grand := 0

	details := strings.Split(invDetails, "\n")
	for i, detail := range details {
		item, total := createItemFromDetail(i+1, detail)
		items = append(items, item)
		grand += total
	}
	return items, grand
}

func createItemFromDetail(number int, detail string) (*Item, int) {
	s := strings.Split(detail, " ")
	n := len(s)

	q, _ := strconv.Atoi(s[n-2])
	p, _ := strconv.Atoi(strings.TrimRight(s[n-1], "\r"))
	t := q * p

	item := &Item{
		Number:      number,
		Description: strings.Join(s[0:n-2], " "),
		Quantity:    q,
		UnitPrice:   toCurrency(p),
		TotalPrice:  toCurrency(t),
	}
	return item, t
}

func toDate(t time.Time) string {
	y, m, d := t.Date()
	return fmt.Sprintf("%d %s %d", d, m, y)
}

func toCurrency(val int) string {
	res := ""
	i := 0
	for val > 0 {
		if i != 0 && i%3 == 0 {
			res = "." + res
		}
		x := val % 10
		val = val / 10

		res = fmt.Sprintf("%d%s", x, res)

		i++
	}
	return res
}
