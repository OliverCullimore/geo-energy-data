package middleware

import (
	"encoding/json"
	"fmt"
	"github.com/olivercullimore/geo-energy-data/server/models"
	"log"
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
			// Recover from and log errors
			/*defer func() {
				if err := recover(); err != nil {
					//w.WriteHeader(http.StatusInternalServerError)
					env.Logger.Printf("%v %v\n", err, debug.Stack())
				}
			}()*/

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
			env.Logger.Printf("%s %s \"%s %s %s\"%s%s time=%s\n", r.RemoteAddr, r.Host, r.Method, r.URL.String(), r.Proto, referrer, userAgent, time.Since(start))
		}

		return http.HandlerFunc(fn)
	}
}

// Headers sets the headers for the request.
func Headers(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set security headers
		w.Header().Set("X-Frame-Options", "SAMEORIGIN")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "same-origin")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		w.Header().Set("Permissions-Policy", "accelerometer=(), camera=(), geolocation=(), gyroscope=(), magnetometer=(), microphone=(), payment=(), usb=()")

		// Set CSP security headers
		//w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'none'; object-src 'none'; style-src 'none'; img-src 'none'; media-src 'none'; frame-src 'none'; font-src 'none'; connect-src 'none'")
		//w.Header().Set("Content-Security-Policy", "script-src 'self' 'unsafe-eval' 'unsafe-inline'")
		//w.Header().Set("Content-Security-Policy-Report-Only", "report-uri https://REPORTING_URL")

		// Set cache headers
		if strings.HasPrefix(r.URL.Path, "/static/") {
			w.Header().Set("Cache-Control", "max-age=86400")
		} else {
			w.Header().Set("Cache-Control", "no-store, no-cache")
		}

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

// CORS sets the CORS headers for the request.
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		/*
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, Access-Control-Request-Headers, Access-Control-Request-Method, Connection, Host, Origin, User-Agent, Referer, Cache-Control, X-header")
		*/
		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

// Auth authenticates the request.
func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get X-Api-Key header
		var header = r.Header.Get("X-Api-Key")
		// Validate & verify access token
		if strings.TrimSpace(header) == "" {
			err := respondWithError(w, http.StatusBadRequest, "Missing auth token")
			if err != nil {
				log.Fatal(err)
			}
			return
		} else {
			// TODO: Verify access token
			/*session, _ := sessions.Store.Get(r, "session")
			_, ok := session.Values["user_id"]
			if !ok {
				http.Redirect(w, r, "/login", 302)
				return
			}*/
		}
		// Call the next handler
		next.ServeHTTP(w, r)
	})
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
