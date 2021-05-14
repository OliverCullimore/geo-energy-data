package routes

import (
	"github.com/gorilla/mux"
	"github.com/olivercullimore/geo-energy-data/server/controllers"
	"github.com/olivercullimore/geo-energy-data/server/middleware"
	"github.com/olivercullimore/geo-energy-data/server/models"
	"net/http"
)

func Initialize(r *mux.Router, env *models.Env) {

	// Initialize logging middleware
	r.Use(middleware.Logging(env))

	// Handle core routes (with Logging)
	r.Handle("/", &middleware.AppHandler{env, controllers.Index})
	r.NotFoundHandler = &middleware.AppHandler{env, controllers.NotFound}
	r.MethodNotAllowedHandler = &middleware.AppHandler{env, controllers.MethodNotAllowed}

	// Handle notification routes (with Logging)
	r.Handle("/notifications/{id:[a-z0-9]+}", &middleware.AppHandler{env, controllers.NotificationTest})

	// Handle API routes (with Logging & CORS)
	apiRouter := r.PathPrefix("/api").Subrouter()
	apiRouter.NotFoundHandler = &middleware.AppHandler{env, controllers.APINotFound}
	apiRouter.MethodNotAllowedHandler = &middleware.AppHandler{env, controllers.APIMethodNotAllowed}
	apiRouter.Use(middleware.CORS)
	apiRouter.Handle("/status", &middleware.AppHandler{env, controllers.APIStatus}).Methods(http.MethodGet)

	// Handle API v1 routes (with Logging & CORS & Auth)
	/*apiAuthRouter := apiRouter.PathPrefix("/v1").Subrouter()
	apiAuthRouter.Use(middleware.Auth)
	apiAuthRouter.Handle("/users", &middleware.AppHandler{env, controllers.APIGetAllUsers}).Methods(http.MethodGet)
	apiAuthRouter.Handle("/users", &middleware.AppHandler{env, controllers.APICreateUser}).Methods(http.MethodPost)
	apiAuthRouter.Handle("/users/{id:[0-9]+}", &middleware.AppHandler{env, controllers.APIGetUser}).Methods(http.MethodGet)
	apiAuthRouter.Handle("/users/{id:[0-9]+}", &middleware.AppHandler{env, controllers.APIUpdateUser}).Methods(http.MethodPut)
	apiAuthRouter.Handle("/users/{id:[0-9]+}", &middleware.AppHandler{env, controllers.APIDeleteUser}).Methods(http.MethodDelete)*/

	// Handle static files
	r.PathPrefix("/").Handler(http.FileServer(http.Dir(env.StaticDir)))
}
