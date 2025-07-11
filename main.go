package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

type InvoiceSummary struct {
	SrNo           int
	InvoiceID      int
	CustomerName   string
	MobileNo       string
	CreatedAt      string
	TotalAmountSum float64
}

type PageData struct {
	Summaries  []InvoiceSummary
	GrandTotal float64
}

type TestItem struct {
	TestName         string
	Price            float64
	Discount         float64
	TotalAmount      float64
	TotalAmountWords string
}

type Invoice struct {
	InvoiceNo        int
	InvoiceDatetime  string
	CustomerName     string
	Gender           string
	Mobile           string
	Address          string
	Tests            []TestItem
	TotalAmount      float64
	TotalAmountWords string
}

func main() {
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "host=localhost port=5432 user=postgres password=root dbname=wellness_invoice sslmode=disable"
	}
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	r := gin.Default()

	r.SetFuncMap(template.FuncMap{
		"formatFloat": func(f float64) string {
			return strconv.FormatFloat(f, 'f', 2, 64)
		},
		"add": func(a, b int) int {
			return a + b
		},
	})

	r.LoadHTMLGlob("templates/*")
	r.Static("/static", "./static")

	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusSeeOther, "/login")
	})

	r.GET("/register", func(c *gin.Context) {
		c.HTML(http.StatusOK, "register.html", nil)
	})

	r.POST("/register", func(c *gin.Context) {
		name := c.PostForm("name")
		address := c.PostForm("address")
		mobile := c.PostForm("mobile")
		username := c.PostForm("username")
		password := c.PostForm("password")

		var existing string
		err := db.QueryRow("SELECT username FROM users WHERE username=$1", username).Scan(&existing)
		if err == nil {
			c.HTML(http.StatusBadRequest, "register.html", gin.H{"Error": "Username already exists"})
			return
		}

		_, err = db.Exec("INSERT INTO users (name, address, mobile, username, password) VALUES ($1, $2, $3, $4, $5)",
			name, address, mobile, username, password)
		if err != nil {
			c.HTML(http.StatusInternalServerError, "register.html", gin.H{"Error": "Registration failed"})
			return
		}
		c.Redirect(http.StatusSeeOther, "/login")
	})

	r.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", nil)
	})

	r.POST("/login", func(c *gin.Context) {
		username := c.PostForm("username")
		password := c.PostForm("password")

		var dbUsername, dbPassword string
		err := db.QueryRow("SELECT username, password FROM users WHERE username=$1", username).Scan(&dbUsername, &dbPassword)
		if err != nil || password != dbPassword {
			c.HTML(http.StatusUnauthorized, "login.html", gin.H{"Error": "Invalid username or password"})
			return
		}
		c.HTML(http.StatusOK, "form.html", gin.H{"Username": dbUsername})
	})

	r.POST("/submit", func(c *gin.Context) {
		customerName := c.PostForm("customer_name")
		mobile := c.PostForm("mobile")
		address := c.PostForm("address")
		gender := c.PostForm("gender")

		services := c.PostFormArray("service[]")
		prices := c.PostFormArray("price[]")
		discounts := c.PostFormArray("discount[]")
		totals := c.PostFormArray("total[]")

		var tests []TestItem
		var totalAmount float64

		for i := range services {
			testName := strings.Split(services[i], "|")[0]
			price, _ := strconv.ParseFloat(prices[i], 64)
			discount, _ := strconv.ParseFloat(discounts[i], 64)
			total, _ := strconv.ParseFloat(totals[i], 64)
			totalAmount += total

			tests = append(tests, TestItem{
				TestName:         testName,
				Price:            price,
				Discount:         discount,
				TotalAmount:      total,
				TotalAmountWords: convertToWords(int(total)),
			})
		}

		totalAmountWords := convertToWords(int(totalAmount))

		var invoiceID int
		err := db.QueryRow(`
			INSERT INTO invoices12 (customer_name, mobile, address)
			VALUES ($1, $2, $3)
			RETURNING id
		`, customerName, mobile, address).Scan(&invoiceID)

		if err != nil {
			c.String(http.StatusInternalServerError, "Invoice Insert Error: %v", err)
			return
		}

		for _, test := range tests {
			_, err := db.Exec(`
				INSERT INTO tests12 (invoice_id, test_name, price, discount, total_amount, total_amount_words)
				VALUES ($1, $2, $3, $4, $5, $6)
			`, invoiceID, test.TestName, test.Price, test.Discount, test.TotalAmount, test.TotalAmountWords)

			if err != nil {
				c.String(http.StatusInternalServerError, "Test Insert Error: %v", err)
				return
			}
		}

		invoice := Invoice{
			InvoiceNo:        invoiceID,
			InvoiceDatetime:  time.Now().Format("02-Jan-2006 03:04 PM"),
			CustomerName:     customerName,
			Gender:           gender,
			Mobile:           mobile,
			Address:          address,
			Tests:            tests,
			TotalAmount:      totalAmount,
			TotalAmountWords: totalAmountWords,
		}

		c.HTML(http.StatusOK, "invoice.html", invoice)
	})

	r.GET("/print", invoiceSummaryHandler(db))
	r.GET("/filter", filterInvoicesHandler(db))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}

func invoiceSummaryHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := db.Query(`
			SELECT 
				invoices12.id AS invoice_id,
				invoices12.customer_name,
				invoices12.mobile AS mobile_no,
				invoices12.created_at,
				SUM(tests12.total_amount) AS total_amount_sum
			FROM 
				invoices12
			JOIN 
				tests12 ON invoices12.id = tests12.invoice_id
			GROUP BY 
				invoices12.id, invoices12.customer_name, invoices12.created_at
			ORDER BY 
				invoices12.id
		`)
		if err != nil {
			log.Println("Invoice summary query error:", err)
			c.String(http.StatusInternalServerError, "DB Error")
			return
		}
		defer rows.Close()

		var summaries []InvoiceSummary
		srNo := 1
		for rows.Next() {
			var s InvoiceSummary
			err := rows.Scan(&s.InvoiceID, &s.CustomerName, &s.MobileNo, &s.CreatedAt, &s.TotalAmountSum)
			if err != nil {
				log.Println("Scan error:", err)
				continue
			}
			s.SrNo = srNo
			srNo++
			summaries = append(summaries, s)
		}

		var grandTotal float64
		err = db.QueryRow(`SELECT SUM(total_amount) FROM tests12`).Scan(&grandTotal)
		if err != nil {
			log.Println("Grand total query error:", err)
			grandTotal = 0
		}

		data := PageData{
			Summaries:  summaries,
			GrandTotal: grandTotal,
		}

		c.HTML(http.StatusOK, "print.html", data)
	}
}

func filterInvoicesHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := c.Query("start")
		end := c.Query("end")

		if start == "" || end == "" {
			c.String(http.StatusBadRequest, "Start and end date are required.")
			return
		}

		startTime := start + " 00:00:00"
		endTime := end + " 23:59:59"

		rows, err := db.Query(`
			SELECT 
				invoices12.id AS invoice_id,
				invoices12.customer_name,
				invoices12.mobile AS mobile_no,
				invoices12.created_at,
				SUM(tests12.total_amount) AS total_amount_sum
			FROM 
				invoices12
			JOIN 
				tests12 ON invoices12.id = tests12.invoice_id
			WHERE 
				invoices12.created_at BETWEEN $1 AND $2
			GROUP BY 
				invoices12.id, invoices12.customer_name, invoices12.created_at
			ORDER BY 
				invoices12.id
		`, startTime, endTime)
		if err != nil {
			log.Println("Date filter query error:", err)
			c.String(http.StatusInternalServerError, "DB Error")
			return
		}
		defer rows.Close()

		var summaries []InvoiceSummary
		srNo := 1
		for rows.Next() {
			var s InvoiceSummary
			err := rows.Scan(&s.InvoiceID, &s.CustomerName, &s.MobileNo, &s.CreatedAt, &s.TotalAmountSum)
			if err != nil {
				log.Println("Scan error:", err)
				continue
			}
			s.SrNo = srNo
			srNo++
			summaries = append(summaries, s)
		}

		var grandTotal float64
		err = db.QueryRow(`
			SELECT SUM(tests12.total_amount)
			FROM invoices12
			JOIN tests12 ON invoices12.id = tests12.invoice_id
			WHERE invoices12.created_at BETWEEN $1 AND $2
		`, startTime, endTime).Scan(&grandTotal)
		if err != nil {
			log.Println("Grand total (filtered) query error:", err)
			grandTotal = 0
		}

		data := PageData{
			Summaries:  summaries,
			GrandTotal: grandTotal,
		}

		c.HTML(http.StatusOK, "print.html", data)
	}
}

func convertToWords(num int) string {
	ones := []string{"", "One", "Two", "Three", "Four", "Five", "Six", "Seven", "Eight", "Nine", "Ten",
		"Eleven", "Twelve", "Thirteen", "Fourteen", "Fifteen", "Sixteen", "Seventeen",
		"Eighteen", "Nineteen"}
	tens := []string{"", "", "Twenty", "Thirty", "Forty", "Fifty", "Sixty", "Seventy", "Eighty", "Ninety"}

	if num == 0 {
		return "Zero Rupees Only"
	}
	if num > 99999 {
		return "Amount too large"
	}

	words := ""
	if num >= 1000 {
		words += ones[num/1000] + " Thousand "
		num %= 1000
	}
	if num >= 100 {
		words += ones[num/100] + " Hundred "
		num %= 100
	}
	if num >= 20 {
		words += tens[num/10] + " "
		num %= 10
	}
	if num > 0 {
		words += ones[num]
	}
	return strings.TrimSpace(words) + " Rupees Only"
}
