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
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

func Run(liveDataFetchInterval, periodicDataFetchInterval, configFile, geoUser, geoPass, calorificValueStr, dbHost, dbPort, dbOrg, dbBucket, dbToken, httpHost, httpPort string) {

	// Initialize logger
	logger := log.New(os.Stdout, "app: ", log.LstdFlags|log.Lshortfile)

	// Validate variables
	checkEnv(liveDataFetchInterval, "Invalid live data fetch interval", logger)
	checkEnv(periodicDataFetchInterval, "Invalid periodic data fetch interval", logger)
	checkEnv(configFile, "Invalid config file", logger)
	checkEnv(geoUser, "Invalid geo user", logger)
	checkEnv(geoPass, "Invalid geo pass", logger)
	checkEnv(calorificValueStr, "Invalid calorific value", logger)
	checkEnv(dbHost, "Invalid InfluxDB Host", logger)
	checkEnv(dbPort, "Invalid InfluxDB Port", logger)
	checkEnv(dbOrg, "Invalid InfluxDB Organization", logger)
	checkEnv(dbBucket, "Invalid InfluxDB Bucket", logger)
	checkEnv(dbToken, "Invalid InfluxDB Token", logger)

	// Load config
	config := models.Config{}
	err := configfile.Load(configFile, &config)
	if err == nil {
		log.Println("Loaded config")
	}

	// Check if system ID is set
	if config.GeoSystemID == "" {
		// Get an access token
		accessToken, err := geo.GetAccessToken(geoUser, geoPass)
		checkErr(logger, err)

		// Get device data to get the system ID
		deviceData, err := geo.GetDeviceData(accessToken)
		checkErr(logger, err)
		//log.Println(deviceData)

		// Set system ID and save config
		config.GeoSystemID = deviceData.SystemDetails[0].SystemID
		log.Println("Saving config")
		err = configfile.Save(configFile, &config)
		checkErr(logger, err)
	}

	// Convert calorific value to float
	calorificValue, err := strconv.ParseFloat(calorificValueStr, 64)
	checkErr(logger, err)

	// Convert fetch intervals to time period intervals
	liveInterval, err := strconv.Atoi(liveDataFetchInterval)
	checkErr(logger, err)
	periodicInterval, err := strconv.Atoi(periodicDataFetchInterval)
	checkErr(logger, err)

	// Initialise env
	env := &models.Env{
		Config:                  config,
		Logger:                  logger,
		NotificationTemplateDir: "server/notifications/templates/",
		TemplateDir:             "server/views/templates/",
		StaticDir:               "server/views/static/",
		ErrorsDir:               "server/views/errors/",
	}

	// Run get meter data task periodically
	go func() {
		liveTicker := time.NewTicker(time.Second * time.Duration(liveInterval))
		periodicTicker := time.NewTicker(time.Second * time.Duration(periodicInterval))
		notificationTicker := time.NewTicker(time.Second * 300)
		done := make(chan bool)
		log.Println("Starting schedulers")
		go scheduler(env, liveTicker, periodicTicker, notificationTicker, done, dbHost, dbPort, dbToken, dbOrg, dbBucket, geoUser, geoPass, config.GeoSystemID, calorificValue)
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		<-sigs
		done <- true
	}()

	// Initialize database client
	dbClient := influxDBClient(dbHost, dbPort, dbToken)
	// Set Env DB client
	env.DB = dbClient

	// Initialize router
	r := mux.NewRouter().StrictSlash(true)

	// Initialize routes
	routes.Initialize(r, env)

	// Initialize http server
	env.Logger.Println("Starting server at http://" + httpHost + ":" + httpPort)
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
	log.Println("Got signal:", sig)

	tc, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	err = s.Shutdown(tc)
	if err != nil {
		env.Logger.Fatal(err)
	} else {
		env.Logger.Println("Shutdown Server")
	}
}

func checkErr(logger *log.Logger, err error) {
	if err != nil {
		logger.Fatal(err)
	}
}

func checkEnv(env, msg string, logger *log.Logger) {
	if env == "" {
		logger.Fatal(msg)
	}
}

func outputJSON(logger *log.Logger, data interface{}, msg string) {
	dataParsed, err := json.MarshalIndent(data, "", "  ")
	checkErr(logger, err)
	log.Printf("%s: \n%s", msg, string(dataParsed))
}

func scheduler(env *models.Env, liveTicker *time.Ticker, periodicTicker *time.Ticker, notificationTicker *time.Ticker, done chan bool, influxDBHost, influxDBPort, influxDBToken, influxDBOrg, influxDBBucket, geoUser, geoPass, geoSystemID string, calorificValue float64) {
	// Run once when first started
	getMeterData(env, influxDBHost, influxDBPort, influxDBToken, influxDBOrg, influxDBBucket, geoUser, geoPass, geoSystemID, calorificValue, true, true)
	sendNotifications(env, influxDBHost, influxDBPort, influxDBToken, influxDBOrg, influxDBBucket)
	for {
		select {
		case <-liveTicker.C:
			// Run live at interval
			getMeterData(env, influxDBHost, influxDBPort, influxDBToken, influxDBOrg, influxDBBucket, geoUser, geoPass, geoSystemID, calorificValue, true, false)
		case <-periodicTicker.C:
			// Run periodic at interval
			getMeterData(env, influxDBHost, influxDBPort, influxDBToken, influxDBOrg, influxDBBucket, geoUser, geoPass, geoSystemID, calorificValue, false, true)
		case <-notificationTicker.C:
			// Run send notifications at interval
			sendNotifications(env, influxDBHost, influxDBPort, influxDBToken, influxDBOrg, influxDBBucket)
		case <-done:
			return
		}
	}
}

func getMeterData(env *models.Env, influxDBHost, influxDBPort, influxDBToken, influxDBOrg, influxDBBucket, geoUser, geoPass, geoSystemID string, calorificValue float64, runLive, runPeriodic bool) {
	// Get an access token
	accessToken, err := geo.GetAccessToken(geoUser, geoPass)
	checkErr(env.Logger, err)

	var data []string

	if runLive {
		// Get live meter data
		env.Logger.Println("Getting live meter data")
		lData := getLiveMeterData(env, accessToken, geoSystemID)
		if len(lData) > 0 {
			//outputJSON(env.Logger, lData, "Writing records")
			data = append(data, lData...)
		}
	}

	if runPeriodic {
		// Get periodic meter data
		env.Logger.Println("Getting periodic meter data")
		pData := getPeriodicMeterData(env, accessToken, geoSystemID, calorificValue)
		if len(pData) > 0 {
			//outputJSON(env.Logger, pData, "Writing records")
			data = append(data, pData...)
		}
	}

	// Write data to InfluxDB if exists
	if len(data) > 0 {
		// Init InfluxDB client
		client := influxDBClient(influxDBHost, influxDBPort, influxDBToken)
		// Write records
		influxDBWriteRecords(data, client, influxDBOrg, influxDBBucket)
	}
}

func getPeriodicMeterData(env *models.Env, accessToken, geoSystemID string, calorificValue float64) []string {
	// Get periodic meter data
	periodicData, err := geo.GetPeriodicMeterData(accessToken, geoSystemID)
	checkErr(env.Logger, err)
	// Debug output
	//outputJSON(env.Logger, periodicData, "Periodic meter data")

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

func getLiveMeterData(env *models.Env, accessToken, geoSystemID string) []string {
	// Get live meter data
	liveData, err := geo.GetLiveMeterData(accessToken, geoSystemID)
	checkErr(env.Logger, err)
	// Debug output
	//outputJSON(env.Logger, liveData, "Live meter data")

	// Check system status
	if len(liveData.SystemStatus) > 0 {
		for _, item := range liveData.SystemStatus {
			if item.StatusType != "STATUS_OK" || item.SystemErrorCode != "ERROR_CODE_NONE" || item.SystemErrorNumber != 0 {
				// Debug output
				outputJSON(env.Logger, liveData, "Live meter data")
			}
		}
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

func sendNotifications(env *models.Env, influxDBHost, influxDBPort, influxDBToken, influxDBOrg, influxDBBucket string) {
	env.Logger.Println("Checking for notifications to send")

	// TODO: Check notification schedule

	// TODO: Send scheduled notifications
	/*err := notifications.Send(env, "Test notification")
	if err != nil {
		env.Logger.Fatal(err)
	}*/

	return
}

func influxDBClient(influxDBHost, influxDBPort, influxDBToken string) influxdb2.Client {
	// Init InfluxDB client and set the timestamp precision
	c := influxdb2.NewClientWithOptions(influxDBHost+":"+influxDBPort, influxDBToken, influxdb2.DefaultOptions().SetPrecision(time.Second))
	// Always close client at the end
	defer c.Close()
	// Return InfluxDB client
	return c
}

func influxDBWriteRecords(recordsList []string, client influxdb2.Client, influxDBOrg, influxDBBucket string) {
	// get non-blocking write client
	writeAPI := client.WriteAPI(influxDBOrg, influxDBBucket)

	// Get errors channel
	errorsCh := writeAPI.Errors()
	// Create go proc for reading and logging errors
	go func() {
		for err := range errorsCh {
			fmt.Printf("InfluxDB write error: %s\n", err.Error())
		}
	}()

	// write records in line protocol
	for _, record := range recordsList {
		writeAPI.WriteRecord(record)
	}

	// Flush writes
	writeAPI.Flush()
}
