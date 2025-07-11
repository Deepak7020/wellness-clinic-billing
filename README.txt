📄 PROJECT NAME: Medical Invoice Generation System

🛠️ TECHNOLOGY STACK:
- Backend: GoLang (Gin Framework)
- Database: PostgreSQL
- Frontend: HTML, CSS, JavaScript

📁 HOW TO USE THIS PROJECT:

1. Create a PostgreSQL database named: `wellness_invoice`
2. Create the following two tables:
   - `invoices12`
   - `tests12`
   *(You can refer to the `postgrsql.txt` file for the SQL schema)*
3. Open `main.go` in VS Code or GoLand.
4. Run the project:
   ```bash
   go run main.go
Open your browser and go to: http://localhost:8080

🌟 FEATURES:

Fill patient details and test information through a simple web form.

Automatic calculation of total amount and amount in words.

Beautifully formatted invoice (printable).

View daily reports and summaries from the /print route.

Filter invoice records by date from /filter.

🖥️ BUILD EXE GUI (Optional):

To create a Windows executable (.exe):

Open terminal and run:

bash
Copy
Edit
GOOS=windows GOARCH=amd64 go build -o InvoiceApp.exe
To create an installer setup GUI:

Use Inno Setup or NSIS (Nullsoft Scriptable Install System)

Example NSIS script:

arduino
Copy
Edit
Outfile "InvoiceSetup.exe"
InstallDir $PROGRAMFILES\InvoiceApp
Page directory
Page instfiles
Section
  SetOutPath $INSTDIR
  File "InvoiceApp.exe"
  File /r "templates\*"
  File /r "static\*"
  File "go.mod"
  File "go.sum"
  File "README.txt"
SectionEnd
📞 CONTACT:8208192987
Created by: Deepak Sasane
Technologies: GoLang, PostgreSQL, HTML/CSS/JS

yaml
Copy
Edit

---

Let me know if you want me to:

✅ Create the `.zip` file with all files and this `README.txt`  
✅ Or help you write the SQL schema for `invoices12` and `tests12` tables (if not ready)

Shall I zip it now and give you the download link?







