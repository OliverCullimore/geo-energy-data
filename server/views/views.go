package views

import (
	"github.com/olivercullimore/geo-energy-data/server/models"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

type Page struct {
	Title string
	View  interface{}
}

// Render will accept a ResponseWriter, environment, template and page interface and writes the code
// and rendered template in HTML format to the ResponseWriter.
func Render(w http.ResponseWriter, env *models.Env, view string, code int, p Page) {
	// Check view template exists
	viewTemplateFile := env.TemplateDir + view + ".html"
	info, err := os.Stat(viewTemplateFile)
	if os.IsNotExist(err) || info.IsDir() {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		env.Logger.Println(err)
		return
	}
	// Check view layout template exists
	viewLayoutTemplateFile := env.TemplateDir + "_layout.html"
	info2, err := os.Stat(viewLayoutTemplateFile)
	if os.IsNotExist(err) || info2.IsDir() {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		env.Logger.Println(err)
		return
	}
	// Render
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(code)
	// Add functions and parse templates
	templates, err := template.New("").Funcs(template.FuncMap{
		"IncludeFile": includeFile,
		"CurrentDate": currentDate,
	}).ParseFiles(viewTemplateFile, viewLayoutTemplateFile)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		env.Logger.Println(err)
		return
	}
	// Execute template
	err = templates.ExecuteTemplate(w, "_layout.html", p)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		env.Logger.Println(err)
		return
	}
}

func includeFile(filepath string) template.HTML {
	// Add prefix to filepath
	filepath = "app/views" + filepath
	// Check file exists
	info, err := os.Stat(filepath)
	if os.IsNotExist(err) || info.IsDir() {
		log.Printf("includeFile - error file not found: %v\n", err)
		return ""
	}
	// Load file contents
	file, err := ioutil.ReadFile(filepath)
	if err != nil {
		log.Printf("includeFile - error reading file: %v\n", err)
		return ""
	}
	// Return content
	return template.HTML(file)
}

func currentDate(format string) template.HTML {
	// Return content
	return template.HTML(time.Now().Format(format))
}
