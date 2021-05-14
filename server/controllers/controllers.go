package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/olivercullimore/geo-energy-data/server/models"
	"github.com/olivercullimore/geo-energy-data/server/notifications"
	"github.com/olivercullimore/geo-energy-data/server/views"
	"html"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

func Index(env *models.Env, w http.ResponseWriter, r *http.Request) {
	views.Render(w, env, "index", http.StatusOK, views.Page{Title: "Test Page", View: nil})
}

func NotificationTest(env *models.Env, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	notifications.Render(w, env, vars["id"], http.StatusOK, notifications.Notification{Title: "Test Notification", Details: map[string]string{"Preheader": "Message"}})
}

/*type entry struct {
	Name string
	Done bool
}
type ToDo struct {
	User string
	List []entry
}

func Test(env *models.Env, w http.ResponseWriter, r *http.Request) {
	var list []entry

	list = append(list, entry{Name: "Task 1", Done: false})
	list = append(list, entry{Name: "Task 2", Done: false})
	list = append(list, entry{Name: "Task 3", Done: false})

	todos := ToDo{User: "test", List: list}

	views.Render(w, env, "test", http.StatusOK, views.Page{Title: "Test Page", View: todos})
}*/

func NotFound(env *models.Env, w http.ResponseWriter, r *http.Request) {
	err := displayError(env, w, r, "404", "Oops! Page Not Found", "404 Page Not Found")
	if err != nil {
		env.Logger.Fatal(err)
	}
}

func MethodNotAllowed(env *models.Env, w http.ResponseWriter, r *http.Request) {
	err := displayError(env, w, r, "405", "Oops! Method Not Allowed", "405 Method Not Allowed")
	if err != nil {
		env.Logger.Fatal(err)
	}
}

func APINotFound(env *models.Env, w http.ResponseWriter, r *http.Request) {
	err := respondWithError(w, http.StatusMethodNotAllowed, "Not Found")
	if err != nil {
		env.Logger.Fatal(err)
		return
	}
}

func APIMethodNotAllowed(env *models.Env, w http.ResponseWriter, r *http.Request) {
	err := respondWithError(w, http.StatusMethodNotAllowed, "Method Not Allowed")
	if err != nil {
		env.Logger.Fatal(err)
		return
	}
}

func APIStatus(env *models.Env, w http.ResponseWriter, r *http.Request) {
	err := respondWithJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	if err != nil {
		env.Logger.Fatal(err)
		return
	}
}

/*
func APICreateUser(env *models.Env, w http.ResponseWriter, r *http.Request) {
	var user models.User
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&user)
	if err != nil {
		err := respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		if err != nil {
			env.Logger.Fatal(err)
		}
		return
	}
	// Defer the close and handle any errors
	defer func() {
		cerr := r.Body.Close()
		if err == nil {
			err = cerr
		}
	}()
	if err != nil {
		env.Logger.Println(err)
		err := respondWithError(w, http.StatusInternalServerError, "Internal Server Error")
		if err != nil {
			env.Logger.Fatal(err)
		}
		return
	}

	// Add user to database
	err = user.Create(env)
	if err != nil {
		env.Logger.Println(err)
		err := respondWithError(w, http.StatusInternalServerError, "Internal Server Error")
		if err != nil {
			env.Logger.Fatal(err)
		}
		return
	}
	err = respondWithJSON(w, http.StatusCreated, user)
	if err != nil {
		env.Logger.Fatal(err)
		return
	}
}

func APIGetUser(env *models.Env, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		err := respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		if err != nil {
			env.Logger.Fatal(err)
		}
		return
	}

	u := models.User{ID: id}
	user, err := u.GetByID(env)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			err := respondWithError(w, http.StatusNotFound, "User not found")
			if err != nil {
				env.Logger.Fatal(err)
			}
		default:
			env.Logger.Println(err)
			err := respondWithError(w, http.StatusInternalServerError, "Internal Server Error")
			if err != nil {
				env.Logger.Fatal(err)
			}
		}
		return
	}
	err = respondWithJSON(w, http.StatusOK, user)
	if err != nil {
		env.Logger.Fatal(err)
		return
	}
}

func APIGetAllUsers(env *models.Env, w http.ResponseWriter, r *http.Request) {
	count, _ := strconv.Atoi(r.FormValue("count"))
	start, _ := strconv.Atoi(r.FormValue("start"))

	if count < 1 {
		count = 10
	}
	if start < 0 {
		start = 0
	}

	users, err := models.GetAllUsers(env, start, count)
	if err != nil {
		env.Logger.Println(err)
		err := respondWithError(w, http.StatusInternalServerError, "Internal Server Error")
		if err != nil {
			env.Logger.Fatal(err)
		}
		return
	}
	err = respondWithJSON(w, http.StatusOK, users)
	if err != nil {
		env.Logger.Fatal(err)
		return
	}
}

func APIUpdateUser(env *models.Env, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		err := respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		if err != nil {
			env.Logger.Fatal(err)
		}
		return
	}
	var user models.User
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&user)
	if err != nil {
		err := respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		if err != nil {
			env.Logger.Fatal(err)
		}
		return
	}
	user.ID = id
	// Defer the close and handle any errors
	defer func() {
		cerr := r.Body.Close()
		if err == nil {
			err = cerr
		}
	}()
	if err != nil {
		env.Logger.Println(err)
		err := respondWithError(w, http.StatusInternalServerError, "Internal Server Error")
		if err != nil {
			env.Logger.Fatal(err)
		}
		return
	}

	// Update user in database
	err = user.Update(env)
	if err != nil {
		env.Logger.Println(err)
		err := respondWithError(w, http.StatusInternalServerError, "Internal Server Error")
		if err != nil {
			env.Logger.Fatal(err)
		}
		return
	}
	err = respondWithJSON(w, http.StatusCreated, user)
	if err != nil {
		env.Logger.Fatal(err)
		return
	}

	if err != nil {
		switch err {
		case errors.New("user not found"):
			err := respondWithError(w, http.StatusNotFound, "User not found")
			if err != nil {
				env.Logger.Fatal(err)
			}
		default:
			env.Logger.Println(err)
			err := respondWithError(w, http.StatusInternalServerError, "Internal Server Error")
			if err != nil {
				env.Logger.Fatal(err)
			}
		}
		return
	}
	err = respondWithJSON(w, http.StatusNoContent, map[string]string{"message": "User updated"})
	if err != nil {
		env.Logger.Fatal(err)
		return
	}
}

func APIDeleteUser(env *models.Env, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		err := respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		if err != nil {
			env.Logger.Fatal(err)
		}
		return
	}

	u := models.User{ID: id}
	err = u.Delete(env)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			err := respondWithError(w, http.StatusNotFound, "User not found")
			if err != nil {
				env.Logger.Fatal(err)
			}
		default:
			env.Logger.Println(err)
			err := respondWithError(w, http.StatusInternalServerError, "Internal Server Error")
			if err != nil {
				env.Logger.Fatal(err)
			}
		}
		return
	}
	err = respondWithJSON(w, http.StatusNoContent, map[string]string{"message": "User deleted"})
	if err != nil {
		env.Logger.Fatal(err)
		return
	}
}
*/

// displayError will accept a ResponseWriter, Request, code, message and messagedetails and will output
// an error page in HTML format to the ResponseWriter.
func displayError(env *models.Env, w http.ResponseWriter, r *http.Request, code, message, messagedetails string) error {
	// Load error.html file if exists
	file := env.ErrorsDir + "error.html"
	errorOccurred := false
	info, err := os.Stat(file)
	if os.IsNotExist(err) || info.IsDir() {
		log.Printf("error file not found: %v\n", err)
		errorOccurred = true
	}
	if !errorOccurred {
		fileContent, err := ioutil.ReadFile(file)
		if err != nil {
			log.Printf("error reading file: %v\n", err)
			errorOccurred = true
		}
		fileContentParsed := strings.Replace(string(fileContent), "{svgerror}", code, 1)
		fileContentParsed = strings.Replace(fileContentParsed, "{errormessage}", message, 1)
		fileContentParsed = strings.Replace(fileContentParsed, "{errormessagedetails}", messagedetails, 1)
		_, err = fmt.Fprint(w, fileContentParsed)
		if err != nil {
			return err
		}
	}
	// Output basic message if error.html doesn't exist
	if errorOccurred {
		_, err := fmt.Fprintf(w, "Error, sorry the page %q was not found.", html.EscapeString(r.URL.Path))
		if err != nil {
			return err
		}
	}
	return nil
}

// respondWithError will accept a ResponseWriter, code and message and writes the code
// and message in JSON format to the ResponseWriter.
func respondWithError(w http.ResponseWriter, code int, message string) error {
	return respondWithJSON(w, code, map[string]string{"error": message})
}

// respondWithJSON will accept a ResponseWriter and a payload and writes the payload
// in JSON format to the ResponseWriter.
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) error {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, err := w.Write(response)
	if err != nil {
		return err
	}
	return nil
}
