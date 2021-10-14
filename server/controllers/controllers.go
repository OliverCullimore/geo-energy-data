package controllers

import (
	"encoding/json"
	"github.com/olivercullimore/geo-energy-data-client"
	"github.com/olivercullimore/geo-energy-data/server/models"
	"log"
	"net/http"
	"runtime"
)

func APIGetCurrentUsage(env *models.Env, w http.ResponseWriter, r *http.Request) {
	// Get an access token
	accessToken, err := geo.GetAccessToken(env.GeoUser, env.GeoPass)
	checkErr(err, env.Logger)

	// Get live meter data
	liveData, err := geo.GetLiveMeterData(accessToken, env.Config.GeoSystemID)
	checkErr(err, env.Logger)

	// Set available power readings
	liveUsage := models.LiveUsage{Electricity: models.LiveUsageData{}, Gas: models.LiveUsageData{}}
	if liveData.PowerTimestamp > 0 && len(liveData.Power) > 0 {
		for _, item := range liveData.Power {
			if item.ValueAvailable {
				if item.Type == "GAS_ENERGY" {
					liveUsage.Gas.Watts = item.Watts
					liveUsage.Gas.LastUpdated = liveData.PowerTimestamp
				} else if item.Type == "ELECTRICITY" {
					liveUsage.Electricity.Watts = item.Watts
					liveUsage.Electricity.LastUpdated = liveData.PowerTimestamp
				}
			}
		}
	}

	// Return available power readings
	if liveUsage.Electricity.LastUpdated > 0 || liveUsage.Gas.LastUpdated > 0 {
		// Debug output
		if env.DebugMode {
			outputJSON(liveUsage, "Current usage data", env.Logger)
		}
		err = respondWithJSON(w, http.StatusOK, liveUsage)
		if err != nil {
			env.Logger.Fatal(err)
			return
		}
	} else {
		err := respondWithError(w, http.StatusNoContent, "No live usage data available")
		if err != nil {
			env.Logger.Fatal(err)
			return
		}
	}
}

func APIGetMeterReadings(env *models.Env, w http.ResponseWriter, r *http.Request) {
	// Get an access token
	accessToken, err := geo.GetAccessToken(env.GeoUser, env.GeoPass)
	checkErr(err, env.Logger)

	// Get periodic meter data
	periodicData, err := geo.GetPeriodicMeterData(accessToken, env.Config.GeoSystemID)
	checkErr(err, env.Logger)

	// Set available power readings
	periodicUsage := models.PeriodicUsage{Electricity: models.PeriodicUsageData{}, Gas: models.PeriodicUsageData{}}
	if periodicData.TotalConsumptionTimestamp > 0 && len(periodicData.TotalConsumptionList) > 0 {
		for _, item := range periodicData.TotalConsumptionList {
			if item.ValueAvailable {
				if item.CommodityType == "GAS_ENERGY" {
					periodicUsage.Gas.TotalConsumption = item.TotalConsumption
					periodicUsage.Gas.ReadingTime = item.ReadingTime
					periodicUsage.Gas.Unit = "m3"
				} else if item.CommodityType == "ELECTRICITY" {
					periodicUsage.Electricity.TotalConsumption = item.TotalConsumption
					periodicUsage.Electricity.ReadingTime = item.ReadingTime
					periodicUsage.Electricity.Unit = "watts"
				}
			}
		}
	}

	// Return available power readings
	if periodicUsage.Electricity.ReadingTime > 0 || periodicUsage.Gas.ReadingTime > 0 {
		// Debug output
		if env.DebugMode {
			outputJSON(periodicUsage, "Meter readings data", env.Logger)
		}
		err = respondWithJSON(w, http.StatusOK, periodicUsage)
		if err != nil {
			env.Logger.Fatal(err)
			return
		}
	} else {
		err := respondWithError(w, http.StatusNoContent, "No periodic usage data available")
		if err != nil {
			env.Logger.Fatal(err)
			return
		}
	}
}

func APIGetLiveData(env *models.Env, w http.ResponseWriter, r *http.Request) {
	// Get an access token
	accessToken, err := geo.GetAccessToken(env.GeoUser, env.GeoPass)
	checkErr(err, env.Logger)

	// Get live meter data
	liveData, err := geo.GetLiveMeterData(accessToken, env.Config.GeoSystemID)
	checkErr(err, env.Logger)

	// Debug output
	if env.DebugMode {
		outputJSON(liveData, "Live meter data", env.Logger)
	}

	// Return data
	if liveData.ID != "" {
		err = respondWithJSON(w, http.StatusOK, liveData)
		if err != nil {
			env.Logger.Fatal(err)
			return
		}
	} else {
		err := respondWithError(w, http.StatusNoContent, "No data available")
		if err != nil {
			env.Logger.Fatal(err)
			return
		}
	}
}

func APIGetPeriodicData(env *models.Env, w http.ResponseWriter, r *http.Request) {
	// Get an access token
	accessToken, err := geo.GetAccessToken(env.GeoUser, env.GeoPass)
	checkErr(err, env.Logger)

	// Get periodic meter data
	periodicData, err := geo.GetPeriodicMeterData(accessToken, env.Config.GeoSystemID)
	checkErr(err, env.Logger)

	// Debug output
	if env.DebugMode {
		outputJSON(periodicData, "Periodic meter data", env.Logger)
	}

	// Return data
	if periodicData.ID != "" {
		err = respondWithJSON(w, http.StatusOK, periodicData)
		if err != nil {
			env.Logger.Fatal(err)
			return
		}
	} else {
		err := respondWithError(w, http.StatusNoContent, "No data available")
		if err != nil {
			env.Logger.Fatal(err)
			return
		}
	}
}

func APIStatus(env *models.Env, w http.ResponseWriter, r *http.Request) {
	err := respondWithJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	if err != nil {
		env.Logger.Fatal(err)
		return
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

func checkErr(err error, logger *log.Logger) {
	if err != nil {
		_, file, line, ok := runtime.Caller(1)
		_, file2, line2, ok2 := runtime.Caller(2)
		if ok && ok2 {
			logger.Fatalf("%s:%d -> %s:%d: %s\n", file2, line2, file, line, err.Error())
		} else if ok {
			logger.Fatalf("%s:%d: %s\n", file, line, err.Error())
		} else {
			logger.Fatal(err)
		}
	}
}

func outputJSON(data interface{}, msg string, logger *log.Logger) {
	dataParsed, err := json.MarshalIndent(data, "", "  ")
	checkErr(err, logger)
	logger.Printf("%s: \n%s", msg, string(dataParsed))
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
