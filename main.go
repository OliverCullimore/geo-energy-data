package main

import (
	"encoding/json"
	"fmt"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/olivercullimore/go-energy-data/app/geotogether"
	"github.com/olivercullimore/go-utils/configfile"
	"github.com/olivercullimore/go-utils/env"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

type Config struct {
	GeoSystemID string
}

func main() {
	log.Println("Starting geo Energy Data")

	// Load environment variables
	err := env.Load(".env")
	if err != nil {
		log.Println(err)
	}
	// Get environment variables
	liveDataFetchInterval := env.Get("LIVE_DATA_FETCH_INTERVAL", "10")
	checkEnv(liveDataFetchInterval, "Invalid live data fetch interval")
	periodicDataFetchInterval := env.Get("PERIODIC_DATA_FETCH_INTERVAL", "300")
	checkEnv(periodicDataFetchInterval, "Invalid periodic data fetch interval")
	configFile := env.Get("CONFIG_FILE", "/config/config.json")
	checkEnv(configFile, "Invalid config file")
	geoUser := env.Get("GEO_USER", "")
	checkEnv(geoUser, "Invalid geo user")
	geoPass := env.Get("GEO_PASS", "")
	checkEnv(geoPass, "Invalid geo pass")
	calorificValueStr := env.Get("CALORIFIC_VALUE", "39.5")
	checkEnv(calorificValueStr, "Invalid calorific value")
	influxDBHost := env.Get("INFLUXDB_HOST", "")
	checkEnv(influxDBHost, "Invalid InfluxDB Host")
	influxDBPort := env.Get("INFLUXDB_PORT", "")
	checkEnv(influxDBPort, "Invalid InfluxDB Port")
	influxDBOrg := env.Get("INFLUXDB_ORG", "")
	checkEnv(influxDBOrg, "Invalid InfluxDB Organization")
	influxDBBucket := env.Get("INFLUXDB_BUCKET", "")
	checkEnv(influxDBBucket, "Invalid InfluxDB Bucket")
	influxDBToken := env.Get("INFLUXDB_TOKEN", "")
	checkEnv(influxDBToken, "Invalid InfluxDB Token")

	// Load config
	config := Config{}
	err = configfile.Load(configFile, &config)
	if err == nil {
		log.Println("Loaded config")
	}

	// Check if system ID is set
	if config.GeoSystemID == "" {
		// Get an access token
		accessToken, err := geotogether.GetAccessToken(geoUser, geoPass)
		checkErr(err)

		// Get device data to get the system ID
		deviceData, err := geotogether.GetDeviceData(accessToken)
		checkErr(err)
		//log.Println(deviceData)

		// Set system ID and save config
		config.GeoSystemID = deviceData.SystemDetails[0].SystemID
		log.Println("Saving config")
		err = configfile.Save(configFile, &config)
		checkErr(err)
	}

	// Convert calorific value to float
	calorificValue, err := strconv.ParseFloat(calorificValueStr, 64)
	checkErr(err)

	// Convert fetch intervals to time period intervals
	liveInterval, err := strconv.Atoi(liveDataFetchInterval)
	checkErr(err)
	periodicInterval, err := strconv.Atoi(periodicDataFetchInterval)
	checkErr(err)

	// Run get meter data task periodically
	tick := time.NewTicker(time.Second * time.Duration(liveInterval))
	tick2 := time.NewTicker(time.Second * time.Duration(periodicInterval))
	done := make(chan bool)
	log.Println("Starting schedulers")
	go scheduler(tick, tick2, done, influxDBHost, influxDBPort, influxDBToken, influxDBOrg, influxDBBucket, geoUser, geoPass, config.GeoSystemID, calorificValue)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	done <- true
}

func scheduler(tick *time.Ticker, tick2 *time.Ticker, done chan bool, influxDBHost, influxDBPort, influxDBToken, influxDBOrg, influxDBBucket, geoUser, geoPass, geoSystemID string, calorificValue float64) {
	// Run once when first started
	getMeterData(time.Now(), influxDBHost, influxDBPort, influxDBToken, influxDBOrg, influxDBBucket, geoUser, geoPass, geoSystemID, calorificValue, true, true)
	for {
		select {
		case t := <-tick.C:
			// Run live at interval
			getMeterData(t, influxDBHost, influxDBPort, influxDBToken, influxDBOrg, influxDBBucket, geoUser, geoPass, geoSystemID, calorificValue, true, false)
		case t2 := <-tick2.C:
			// Run periodic at interval
			getMeterData(t2, influxDBHost, influxDBPort, influxDBToken, influxDBOrg, influxDBBucket, geoUser, geoPass, geoSystemID, calorificValue, false, true)
		case <-done:
			return
		}
	}
}

func getMeterData(t time.Time, influxDBHost, influxDBPort, influxDBToken, influxDBOrg, influxDBBucket, geoUser, geoPass, geoSystemID string, calorificValue float64, runLive, runPeriodic bool) {
	//fmt.Println("Running at ", t)

	// Get an access token
	accessToken, err := geotogether.GetAccessToken(geoUser, geoPass)
	checkErr(err)

	var data []string

	if runLive {
		// Get live meter data
		lData := getLiveMeterData(accessToken, geoSystemID)
		if len(lData) > 0 {
			outputJSON(lData, "Writing records")
			data = append(data, lData...)
		}
	}

	if runPeriodic {
		// Get periodic meter data
		pData := getPeriodicMeterData(accessToken, geoSystemID, calorificValue)
		if len(pData) > 0 {
			outputJSON(pData, "Writing records")
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

func getPeriodicMeterData(accessToken, geoSystemID string, calorificValue float64) []string {

	// Get periodic meter data
	periodicData, err := geotogether.GetPeriodicMeterData(accessToken, geoSystemID)
	checkErr(err)
	// Debug output
	//outputJSON(periodicData, "Periodic meter data")

	var pData []string

	// Add consumption readings
	if periodicData.TotalConsumptionTimestamp > 0 && len(periodicData.TotalConsumptionList) > 0 {
		for _, item := range periodicData.TotalConsumptionList {
			if item.ValueAvailable {
				totalConsumption := item.TotalConsumption
				if item.CommodityType == "GAS_ENERGY" {
					pData = append(pData, fmt.Sprintf("meterdata,source=periodic,unit=m3,type=%s val=%f %d", item.CommodityType, item.TotalConsumption, item.ReadingTime))
					totalConsumption = geotogether.ConvertToKWH(item.TotalConsumption, calorificValue)
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

func getLiveMeterData(accessToken, geoSystemID string) []string {
	// Get live meter data
	liveData, err := geotogether.GetLiveMeterData(accessToken, geoSystemID)
	checkErr(err)
	// Debug output
	//outputJSON(liveData, "Live meter data")

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

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func checkEnv(env, msg string) {
	if env == "" {
		log.Fatal(msg)
	}
}

func outputJSON(data interface{}, msg string) {
	dataParsed, err := json.MarshalIndent(data, "", "  ")
	checkErr(err)
	log.Printf("%s: \n%s", msg, string(dataParsed))
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
