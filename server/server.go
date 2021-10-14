package server

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/olivercullimore/geo-energy-data-client"
	"github.com/olivercullimore/geo-energy-data/server/models"
	"github.com/olivercullimore/geo-energy-data/server/routes"
	"github.com/olivercullimore/go-utils/configfile"
	envs "github.com/olivercullimore/go-utils/env"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"syscall"
	"time"
)

func Run() {

	// Initialize logger
	logger := log.New(os.Stdout, "app: ", log.LstdFlags)

	logger.Println("Starting geo Energy Data")

	// Load environment variables if .env file exists
	info, err := os.Stat(".env")
	if !os.IsNotExist(err) && !info.IsDir() {
		err := envs.Load(".env")
		if err != nil {
			logger.Println(err)
		} else {
			logger.Println("Loaded .env file")
		}
	}
	// Get debug mode
	debugModeStr := checkConfig("DEBUG_MODE", "false", "Debug mode", "", logger)
	// Debug mode enabled?
	debugMode := false
	if debugModeStr == "true" {
		logger.SetFlags(log.LstdFlags | log.Lshortfile)
		debugMode = true
	}
	// Get main environment variables
	configFile := checkConfig("CONFIG_FILE", "/config/config.json", "config file", "", logger)
	enableAPIStr := checkConfig("ENABLE_API", "false", "Enable API", "", logger)
	enableInfluxDBStr := checkConfig("ENABLE_INFLUXDB", "true", "Enable InfluxDB", "", logger)
	geoUser := checkConfig("GEO_USER", "", "geo user", "", logger)
	geoPass := checkConfig("GEO_PASS", "", "geo pass", "", logger)
	calorificValueStr := checkConfig("CALORIFIC_VALUE", "39.5", "calorific value", "", logger)
	httpPort := ""
	apiKey := ""
	liveDataFetchInterval := "30"
	periodicDataFetchInterval := "300"
	influxDBHost := ""
	influxDBPort := ""
	influxDBOrg := ""
	influxDBBucket := ""
	influxDBToken := ""

	// API enabled?
	enableAPI := false
	if enableAPIStr == "true" {
		enableAPI = true
		httpPort = checkConfig("HTTP_PORT", "80", "HTTP port", "numeric", logger)
		apiKey = checkConfig("API_KEY", "", "API key", "", logger)
	}
	// InfluxDB enabled?
	enableInfluxDB := false
	if enableInfluxDBStr == "true" {
		enableInfluxDB = true
		liveDataFetchInterval = checkConfig("LIVE_DATA_FETCH_INTERVAL", "10", "live data fetch interval", "numeric", logger)
		periodicDataFetchInterval = checkConfig("PERIODIC_DATA_FETCH_INTERVAL", "300", "periodic data fetch interval", "numeric", logger)
		influxDBHost = checkConfig("INFLUXDB_HOST", "", "InfluxDB host", "url", logger)
		influxDBPort = checkConfig("INFLUXDB_PORT", "8086", "InfluxDB port", "numeric", logger)
		influxDBOrg = checkConfig("INFLUXDB_ORG", "", "InfluxDB organization", "", logger)
		influxDBBucket = checkConfig("INFLUXDB_BUCKET", "", "InfluxDB bucket", "", logger)
		influxDBToken = checkConfig("INFLUXDB_TOKEN", "", "InfluxDB token", "", logger)
	}

	// Load config
	config := models.Config{}
	err = configfile.Load(configFile, &config)
	if err == nil {
		logger.Println("Loaded config")
	}

	// Check if system ID is set
	if config.GeoSystemID == "" && geoUser != "" && geoPass != "" {
		// Get an access token
		accessToken, err := geo.GetAccessToken(geoUser, geoPass)
		checkErr(err, logger)
		if accessToken == "" {
			logger.Fatalf("Unable to retrieve an access token. Please check your login details are correct\n")
		}

		// Get device data to get the system ID
		deviceData, err := geo.GetDeviceData(accessToken)
		checkErr(err, logger)
		if debugMode {
			logger.Println(deviceData)
		}

		// Set system ID and save config
		if len(deviceData.SystemDetails) > 0 && deviceData.SystemDetails[0].SystemID != "" {
			config.GeoSystemID = deviceData.SystemDetails[0].SystemID
			logger.Println("Saving config")
			err = configfile.Save(configFile, &config)
			checkErr(err, logger)
		} else {
			if debugMode {
				logger.Fatalf("No system ID found in: %v\n", deviceData)
			} else {
				logger.Fatalf("No system ID found\n")
			}
		}
	} else {
		// Check login details are still valid
		authData, err := geo.Login(geoUser, geoPass)
		checkErr(err, logger)
		if authData.AccessToken == "" {
			logger.Fatalf("Unable to retrieve an access token. Please check your login details are correct\n")
		}
	}

	// Convert fetch intervals to time period intervals
	liveInterval, err := strconv.Atoi(liveDataFetchInterval)
	checkErr(err, logger)
	periodicInterval, err := strconv.Atoi(periodicDataFetchInterval)
	checkErr(err, logger)

	// Convert calorific value to float and save config
	calorificValue, err := strconv.ParseFloat(calorificValueStr, 64)
	checkErr(err, logger)
	if calorificValue != config.CalorificValue {
		config.CalorificValue = calorificValue
		logger.Println("Saving config")
		err = configfile.Save(configFile, &config)
		checkErr(err, logger)
	}

	// Initialise env
	env := &models.Env{
		Config:         config,
		Logger:         logger,
		GeoUser:        geoUser,
		GeoPass:        geoPass,
		APIKey:         apiKey,
		EnableAPI:      enableAPI,
		EnableInfluxDB: enableInfluxDB,
		DebugMode:      debugMode,
	}

	// Check System ID is set
	if config.GeoSystemID != "" {
		// Initialize get meter data periodic tasks
		if env.EnableInfluxDB {
			tick := time.NewTicker(time.Second * time.Duration(liveInterval))
			tick2 := time.NewTicker(time.Second * time.Duration(periodicInterval))
			go func() {
				done := make(chan bool)
				env.Logger.Println("Starting schedulers")
				go scheduler(tick, tick2, done, influxDBHost, influxDBPort, influxDBToken, influxDBOrg, influxDBBucket, geoUser, geoPass, env.Config.GeoSystemID, env.Config.CalorificValue, env.DebugMode, env.Logger)
				sigs := make(chan os.Signal, 1)
				signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
				<-sigs
				done <- true
			}()
		}
	}

	// Initialize API
	if env.EnableAPI {
		// Initialize router
		r := mux.NewRouter().StrictSlash(true)

		// Initialize routes
		routes.Initialize(r, env)

		// Initialize http server
		env.Logger.Println("Starting API server")
		s := &http.Server{
			Addr:         ":" + httpPort,    // configure the bind address
			Handler:      r,                 // set the default handler
			ErrorLog:     env.Logger,        // set the logger for the server
			IdleTimeout:  120 * time.Second, // max time to read request from the client
			ReadTimeout:  5 * time.Second,   // max time to write response to the client
			WriteTimeout: 10 * time.Second,  // max time for connections using TCP Keep-Alive
		}
		// Run http server
		go func() {
			err := s.ListenAndServe()
			if err != nil {
				env.Logger.Printf("Error starting server: %s\n", err)
				os.Exit(1)
			}
		}()

		// Listen for interrupts
		sigChan := make(chan os.Signal)
		signal.Notify(sigChan, os.Interrupt)
		signal.Notify(sigChan, os.Kill)

		sig := <-sigChan
		env.Logger.Println("Got signal:", sig)

		tc, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		err = s.Shutdown(tc)
		if err != nil {
			env.Logger.Fatal(err)
		} else {
			env.Logger.Println("Shutdown Server")
		}
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

func checkConfig(envKey, defaultValue, name, validationType string, logger *log.Logger) string {
	valid := true
	//logger.Printf("n: %s c: %s e: %s d: %s", name, configValue, envValue, defaultValue)
	checkVal := envs.Get(envKey, "")
	if checkVal == "" {
		checkVal = defaultValue
	}
	switch validationType {
	case "numeric":
		_, err := strconv.Atoi(checkVal)
		if err != nil {
			valid = false
		}
	case "url":
		_, err := url.ParseRequestURI(checkVal)
		if err != nil {
			valid = false
		}
	case "":
		if checkVal == "" {
			valid = false
		}
	}
	if valid != true {
		logger.Fatalf("Invalid %s value", name)
	}
	return checkVal
}

func outputJSON(data interface{}, msg string, logger *log.Logger) {
	dataParsed, err := json.MarshalIndent(data, "", "  ")
	checkErr(err, logger)
	logger.Printf("%s: \n%s", msg, string(dataParsed))
}

func scheduler(tick *time.Ticker, tick2 *time.Ticker, done chan bool, influxDBHost, influxDBPort, influxDBToken, influxDBOrg, influxDBBucket, geoUser, geoPass, geoSystemID string, calorificValue float64, debugMode bool, logger *log.Logger) {
	// Run once when first started
	getMeterData(time.Now(), influxDBHost, influxDBPort, influxDBToken, influxDBOrg, influxDBBucket, geoUser, geoPass, geoSystemID, calorificValue, true, true, debugMode, logger)
	for {
		select {
		case t := <-tick.C:
			// Run live at interval
			getMeterData(t, influxDBHost, influxDBPort, influxDBToken, influxDBOrg, influxDBBucket, geoUser, geoPass, geoSystemID, calorificValue, true, false, debugMode, logger)
		case t2 := <-tick2.C:
			// Run periodic at interval
			getMeterData(t2, influxDBHost, influxDBPort, influxDBToken, influxDBOrg, influxDBBucket, geoUser, geoPass, geoSystemID, calorificValue, false, true, debugMode, logger)
		case <-done:
			return
		}
	}
}

func getMeterData(t time.Time, influxDBHost, influxDBPort, influxDBToken, influxDBOrg, influxDBBucket, geoUser, geoPass, geoSystemID string, calorificValue float64, runLive, runPeriodic, debugMode bool, logger *log.Logger) {
	// Debug output
	if debugMode {
		if runLive {
			logger.Println("Running get live data at", t)
		} else {
			logger.Println("Running get periodic data at", t)
		}
	}

	// Get an access token
	accessToken, err := geo.GetAccessToken(geoUser, geoPass)
	checkErr(err, logger)

	var data []string

	if runLive {
		// Get live meter data
		lData := getLiveMeterData(accessToken, geoSystemID, debugMode, logger)
		if len(lData) > 0 {
			// Debug output
			if debugMode {
				outputJSON(lData, "Writing records", logger)
			}
			data = append(data, lData...)
		}
	}

	if runPeriodic {
		// Get periodic meter data
		pData := getPeriodicMeterData(accessToken, geoSystemID, calorificValue, debugMode, logger)
		if len(pData) > 0 {
			// Debug output
			if debugMode {
				outputJSON(pData, "Writing records", logger)
			}
			data = append(data, pData...)
		}
	}

	// Write data to InfluxDB if exists
	if len(data) > 0 {
		// Init InfluxDB client
		client := influxDBClient(influxDBHost, influxDBPort, influxDBToken)
		// Write records
		influxDBWriteRecords(data, client, influxDBOrg, influxDBBucket, logger)
	}
}

func getPeriodicMeterData(accessToken, geoSystemID string, calorificValue float64, debugMode bool, logger *log.Logger) []string {

	// Get periodic meter data
	periodicData, err := geo.GetPeriodicMeterData(accessToken, geoSystemID)
	checkErr(err, logger)
	// Debug output
	if debugMode {
		outputJSON(periodicData, "Periodic meter data", logger)
	}

	var pData []string

	// Add consumption readings
	if periodicData.TotalConsumptionTimestamp > 0 && len(periodicData.TotalConsumptionList) > 0 {
		for _, item := range periodicData.TotalConsumptionList {
			if item.ValueAvailable {
				totalConsumption := item.TotalConsumption
				if item.CommodityType == "GAS_ENERGY" {
					pData = append(pData, fmt.Sprintf("meterdata,source=periodic,unit=m3,type=%s val=%f %d", item.CommodityType, item.TotalConsumption, item.ReadingTime))
					totalConsumption = geo.ConvertToKWH(item.TotalConsumption, calorificValue)
				}
				pData = append(pData, fmt.Sprintf("meterdata,source=periodic,unit=watts,type=%s val=%f %d", item.CommodityType, totalConsumption, item.ReadingTime))
			}
		}
	}

	// Add bill to date
	if periodicData.BillToDateTimestamp > 0 && len(periodicData.BillToDateList) > 0 {
		for _, item := range periodicData.BillToDateList {
			pData = append(pData, fmt.Sprintf("meterdata_bill,source=periodic,type=%s,validutc=%d,startutc=%d val=%f %d", item.CommodityType, item.ValidUTC, item.StartUTC, item.BillToDate, periodicData.BillToDateTimestamp))
		}
	}

	// Add active tariff data
	if periodicData.ActiveTariffTimestamp > 0 && len(periodicData.ActiveTariffList) > 0 {
		for _, item := range periodicData.ActiveTariffList {
			if item.ValueAvailable {
				pData = append(pData, fmt.Sprintf("meterdata_tariff,source=periodic,type=%s val=%f %d", item.CommodityType, item.ActiveTariffPrice, periodicData.ActiveTariffTimestamp))
			}
		}
	}

	// Add current electricity costs
	if periodicData.CurrentCostsElecTimestamp > 0 && len(periodicData.CurrentCostsElec) > 0 {
		for _, item := range periodicData.CurrentCostsElec {
			pData = append(pData, fmt.Sprintf("meterdata_currentcosts,source=periodic,type=%s,duration=%s,subtype=cost val=%f %d", item.CommodityType, item.Duration, item.CostAmount, periodicData.CurrentCostsElecTimestamp))
			pData = append(pData, fmt.Sprintf("meterdata_currentcosts,source=periodic,type=%s,duration=%s,subtype=energy val=%f %d", item.CommodityType, item.Duration, item.EnergyAmount, periodicData.CurrentCostsElecTimestamp))
		}
	}

	// Add current gas costs
	if periodicData.CurrentCostsGasTimestamp > 0 && len(periodicData.CurrentCostsGas) > 0 {
		for _, item := range periodicData.CurrentCostsGas {
			pData = append(pData, fmt.Sprintf("meterdata_currentcosts,source=periodic,type=%s,duration=%s,subtype=cost val=%f %d", item.CommodityType, item.Duration, item.CostAmount, periodicData.CurrentCostsGasTimestamp))
			pData = append(pData, fmt.Sprintf("meterdata_currentcosts,source=periodic,type=%s,duration=%s,subtype=energy val=%f %d", item.CommodityType, item.Duration, item.EnergyAmount, periodicData.CurrentCostsGasTimestamp))
		}
	}

	return pData
}

func getLiveMeterData(accessToken, geoSystemID string, debugMode bool, logger *log.Logger) []string {
	// Get live meter data
	liveData, err := geo.GetLiveMeterData(accessToken, geoSystemID)
	checkErr(err, logger)
	// Debug output
	if debugMode {
		outputJSON(liveData, "Live meter data", logger)
	}

	var lData []string

	// Add power readings
	if liveData.PowerTimestamp > 0 && len(liveData.Power) > 0 {
		for _, item := range liveData.Power {
			if item.ValueAvailable {
				lData = append(lData, fmt.Sprintf("meterdata,source=live,unit=watts,type=%s val=%f %d", item.Type, item.Watts, liveData.PowerTimestamp))
			}
		}
	}

	return lData
}

func influxDBClient(influxDBHost, influxDBPort, influxDBToken string) influxdb2.Client {
	// Init InfluxDB client and set the timestamp precision
	c := influxdb2.NewClientWithOptions(influxDBHost+":"+influxDBPort, influxDBToken, influxdb2.DefaultOptions().SetPrecision(time.Second))
	// Always close client at the end
	defer c.Close()
	// Return InfluxDB client
	return c
}

func influxDBWriteRecords(recordsList []string, client influxdb2.Client, influxDBOrg, influxDBBucket string, logger *log.Logger) {
	// get non-blocking write client
	writeAPI := client.WriteAPI(influxDBOrg, influxDBBucket)

	// Get errors channel
	errorsCh := writeAPI.Errors()
	// Create go proc for reading and logging errors
	go func() {
		for err := range errorsCh {
			logger.Printf("InfluxDB write error: %s\n", err.Error())
		}
	}()

	// write records in line protocol
	for _, record := range recordsList {
		writeAPI.WriteRecord(record)
	}

	// Flush writes
	writeAPI.Flush()
}
