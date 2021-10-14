package routes

import (
	"github.com/gorilla/mux"
	"github.com/olivercullimore/geo-energy-data/server/controllers"
	"github.com/olivercullimore/geo-energy-data/server/middleware"
	"github.com/olivercullimore/geo-energy-data/server/models"
	"net/http"
)

func Initialize(r *mux.Router, env *models.Env) {

	// Handle API routes (with Logging & CORS)
	apiRouter := r.PathPrefix("/api").Subrouter()
	apiRouter.Use(middleware.Logging(env))
	apiRouter.Use(middleware.CORS)
	apiRouter.NotFoundHandler = &middleware.AppHandler{env, controllers.APINotFound}
	apiRouter.MethodNotAllowedHandler = &middleware.AppHandler{env, controllers.APIMethodNotAllowed}
	apiRouter.Handle("/status", &middleware.AppHandler{env, controllers.APIStatus}).Methods(http.MethodGet)

	// Handle Authenticated API routes (with Logging & CORS & Auth)
	apiAuthRouter := apiRouter.PathPrefix("/beta").Subrouter()
	apiAuthRouter.Use(middleware.Auth(env))
	apiAuthRouter.Handle("/currentusage", &middleware.AppHandler{env, controllers.APIGetCurrentUsage}).Methods(http.MethodGet)
	apiAuthRouter.Handle("/meterreadings", &middleware.AppHandler{env, controllers.APIGetMeterReadings}).Methods(http.MethodGet)
	apiAuthRouter.Handle("/live", &middleware.AppHandler{env, controllers.APIGetLiveData}).Methods(http.MethodGet)
	apiAuthRouter.Handle("/periodic", &middleware.AppHandler{env, controllers.APIGetPeriodicData}).Methods(http.MethodGet)

}
