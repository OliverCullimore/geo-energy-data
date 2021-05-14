package notifications

import (
	"errors"
	"github.com/containrrr/shoutrrr"
	"github.com/containrrr/shoutrrr/pkg/types"
	"github.com/olivercullimore/geo-energy-data/server/models"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

func Send(env *models.Env, message string) error {
	if len(env.Config.ShoutrrrServices) > 0 {
		for _, serviceURL := range env.Config.ShoutrrrServices {
			// Create sender
			sender, err := shoutrrr.NewSender(env.Logger, serviceURL)
			if err != nil {
				return err
			}
			// Send notifications instantly to all sender's services
			sender.Send(message, (*types.Params)(&map[string]string{"title": "Message"}))
		}
	} else {
		return errors.New("no Shoutrrr services configured")
	}
	return nil
}

type Notification struct {
	Title   string
	Details interface{}
}

// Render will accept a ResponseWriter, environment, template and page interface and writes the code
// and rendered template in HTML format to the ResponseWriter.
//func Render(w http.ResponseWriter, env *models.Env, notification string, code int, p Page) {
func Render(w http.ResponseWriter, env *models.Env, notificationTemplate string, code int, n Notification) {
	// Check notification template exists
	notificationTemplateFile := env.NotificationTemplateDir + notificationTemplate + ".html"
	info, err := os.Stat(notificationTemplateFile)
	if os.IsNotExist(err) || info.IsDir() {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		env.Logger.Println(err)
		return
	}
	// Check notification layout template exists
	notificationLayoutTemplateFile := env.NotificationTemplateDir + "_email.html"
	info2, err := os.Stat(notificationLayoutTemplateFile)
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
	}).ParseFiles(notificationTemplateFile, notificationLayoutTemplateFile)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		env.Logger.Println(err)
		return
	}
	// Execute template
	err = templates.ExecuteTemplate(w, "_email.html", n)
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
