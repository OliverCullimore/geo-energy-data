package middleware

import (
	"encoding/json"
	"fmt"
	"github.com/olivercullimore/geo-energy-data/server/models"
	"net/http"
	"strings"
	"time"
)

// The Handler struct that takes a configured Env and a function matching
// our useful signature.
type AppHandler struct {
	Env     *models.Env
	Handler func(env *models.Env, w http.ResponseWriter, r *http.Request)
}

// ServeHTTP allows your type to satisfy the http.Handler interface.
func (ah *AppHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ah.Handler(ah.Env, w, r)
}

// Logging logs the incoming HTTP request & its duration.
func Logging(env *models.Env) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			// Set request start time
			start := time.Now()
			// Call the next handler
			next.ServeHTTP(w, r)
			// Log request details and duration
			referrer := r.Referer()
			if referrer != "" {
				referrer = fmt.Sprintf(" referrer=%q", referrer)
			}
			userAgent := r.UserAgent()
			if userAgent != "" {
				userAgent = fmt.Sprintf(" userAgent=%q", userAgent)
			}
			if env.DebugMode {
				env.Logger.Printf("%s %s \"%s %s %s\"%s%s time=%s\n", r.RemoteAddr, r.Host, r.Method, r.URL.String(), r.Proto, referrer, userAgent, time.Since(start))
			}
		}
		return http.HandlerFunc(fn)
	}
}

// CORS sets the CORS headers for the request.
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

// Auth authenticates the request.
func Auth(env *models.Env) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			// Get X-Api-Key header
			var header = r.Header.Get("X-Api-Key")
			// Validate & verify access token
			if strings.TrimSpace(header) == "" {
				err := respondWithError(w, http.StatusBadRequest, "Missing API key")
				if err != nil {
					env.Logger.Println(err)
				}
				return
			} else if header != env.APIKey {
				err := respondWithError(w, http.StatusBadRequest, "Invalid API key")
				if err != nil {
					env.Logger.Println(err)
				}
				return
			}
			// Call the next handler
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
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
