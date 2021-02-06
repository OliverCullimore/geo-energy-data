package geotogether

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
)

const (
	geoBaseURL = "https://api.geotogether.com" // API Base URL (without trailing slash)
)

type EndpointDataEntry struct {
	ID int
}

type EndpointData struct {
	Type    string
	Entries []EndpointDataEntry
}

type Endpoint struct {
	Name string
	Data []EndpointData
}

/*
{
	"username": "example-user-name",
	"email": "example@gmail.com",
	"displayName": "example-display-name",
	"validated": true,
	"accessToken": "example-access-token"
}
*/
type AuthData struct {
	Username    string `json:"username"`
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
	Validated   bool   `json:"validated"`
	AccessToken string `json:"accessToken"`
}

// Login retrieves a users details and an access token to use in other requests
func Login(user, pass string) (AuthData, error) {
	// Define request parameters
	requestBody := []byte(`{"identity":"` + user + `","password":"` + pass + `"}`)

	// Define request headers
	requestHeaders := map[string][]string{
		"Content-Type": {"application/json"},
	}

	// Make login request
	body, err := makeRequest("/usersservice/v2/login", "POST", nil, requestHeaders, bytes.NewBuffer(requestBody))
	if err != nil {
		return AuthData{}, err
	}
	//fmt.Println("Response: " + string(body))

	var authData AuthData
	err = json.Unmarshal(body, &authData)
	if err != nil {
		log.Fatal(err)
		return AuthData{}, err
	}

	return authData, err
}

// GetAccessToken retrieves an access token to use in other requests
func GetAccessToken(user, pass string) (string, error) {
	// Get an access token by logging in
	authData, err := Login(user, pass)
	if err != nil {
		log.Println(err)
	}
	//log.Println(authData)
	accessToken := authData.AccessToken

	return accessToken, err
}

/*
{
    "systemRoles": [
        {
            "name": "6fb578fd-5295-40aa-8188-918f90ca1a3a",
            "systemId": "ef9272fd-2a72-4306-a96f-7d624c0fbd31",
            "roles": [
                "READ",
                "WRITE"
            ]
        }
    ],
    "systemDetails": [
        {
            "name": "Cullimore",
            "devices": [
                {
                    "deviceType": "TRIO_II_TB_GEO",
                    "sensorType": 94,
                    "nodeId": 0,
                    "versionNumber": {
                        "major": 4,
                        "minor": 2
                    },
                    "pairedTimestamp": 0,
                    "pairingCode": "AAAAAA",
                    "upgradeRequired": false
                },
                {
                    "deviceType": "WIFI_MODULE",
                    "sensorType": 81,
                    "nodeId": 64,
                    "versionNumber": {
                        "major": 2,
                        "minor": 6
                    },
                    "pairedTimestamp": 0,
                    "pairingCode": "AAAAAA",
                    "upgradeRequired": false
                }
            ],
            "systemId": "ef9272fd-2a72-4306-a96f-7d624c0fbd31"
        }
    ]
}
*/
type DeviceDataSystemRoles struct {
	Name     string   `json:"name"`
	SystemID string   `json:"systemId"`
	Roles    []string `json:"roles"`
}
type DeviceDataSystemDetailsDeviceVersionNumber struct {
	Major int64 `json:"major"`
	Minor int64 `json:"minor"`
}
type DeviceDataSystemDetailsDevice struct {
	DeviceType      string                                     `json:"deviceType"`
	SensorType      float64                                    `json:"sensorType"`
	NodeID          float64                                    `json:"nodeId"`
	VersionNumber   DeviceDataSystemDetailsDeviceVersionNumber `json:"versionNumber"`
	PairedTimestamp int64                                      `json:"pairedTimestamp"`
	PairingCode     string                                     `json:"pairingCode"`
	UpgradeRequired bool                                       `json:"upgradeRequired"`
}
type DeviceDataSystemDetails struct {
	Name     string                          `json:"name"`
	Devices  []DeviceDataSystemDetailsDevice `json:"devices"`
	SystemID string                          `json:"systemId"`
}
type DeviceData struct {
	SystemRoles   []DeviceDataSystemRoles   `json:"systemRoles"`
	SystemDetails []DeviceDataSystemDetails `json:"systemDetails"`
	LatestUTC     time.Time                 `json:"latestUtc"`
}

// GetDeviceData retrieves device data
func GetDeviceData(accessToken string) (DeviceData, error) {
	// Define request headers
	requestHeaders := map[string][]string{
		"Authorization": {"Bearer " + accessToken},
	}

	// Get data from geo API
	body, err := makeRequest("/api/userapi/v2/user/detail-systems?systemDetails=true", "GET", nil, requestHeaders, nil)
	if err != nil {
		return DeviceData{}, err
	}
	//fmt.Println("Response: " + string(body))

	var deviceData DeviceData
	err = json.Unmarshal(body, &deviceData)
	if err != nil {
		log.Fatal(err)
		return DeviceData{}, err
	}

	return deviceData, err
}

/*
{
    "ttl": 1800,
    "latestUtc": 1612126162,
    "id": "ef9272fd-2a72-4306-a96f-7d624c0fbd31",
    "totalConsumptionList": [
        {
            "commodityType": "ELECTRICITY",
            "readingTime": 1612126162,
            "totalConsumption": 16968.136,
            "valueAvailable": true
        },
        {
            "commodityType": "GAS_ENERGY",
            "readingTime": 1612125000,
            "totalConsumption": 1344743.0,
            "valueAvailable": true
        }
    ],
    "totalConsumptionTimestamp": 1612126162,
    "supplyStatusList": [
        {
            "commodityType": "ELECTRICITY",
            "supplyStatus": "SUPPLYOFF"
        },
        {
            "commodityType": "GAS_ENERGY",
            "supplyStatus": "SUPPLYOFF"
        }
    ],
    "supplyStatusTimestamp": 1612126162,
    "billToDateList": [
        {
            "commodityType": "ELECTRICITY",
            "billToDate": 11988.49,
            "validUTC": 1612126162,
            "startUTC": 946684800,
            "duration": 0,
            "valueAvailable": true
        },
        {
            "commodityType": "GAS_ENERGY",
            "billToDate": 4267.0,
            "validUTC": 1612125000,
            "startUTC": 946684800,
            "duration": 0,
            "valueAvailable": true
        }
    ],
    "billToDateTimestamp": 1612126162,
    "activeTariffList": [
        {
            "commodityType": "ELECTRICITY",
            "valueAvailable": true,
            "nextTariffStartTime": 0,
            "activeTariffPrice": 10.0,
            "nextTariffPrice": 0.0,
            "nextPriceAvailable": false
        },
        {
            "commodityType": "GAS_ENERGY",
            "valueAvailable": true,
            "nextTariffStartTime": 0,
            "activeTariffPrice": 2.229,
            "nextTariffPrice": 0.0,
            "nextPriceAvailable": false
        }
    ],
    "activeTariffTimestamp": 1612126162,
    "currentCostsElec": [
        {
            "commodityType": "ELECTRICITY",
            "duration": "DAY",
            "period": 0,
            "costAmount": 326.68,
            "energyAmount": 31.669
        },
        {
            "commodityType": "ELECTRICITY",
            "duration": "WEEK",
            "period": 0,
            "costAmount": 2567.5,
            "energyAmount": 249.751
        },
        {
            "commodityType": "ELECTRICITY",
            "duration": "MONTH",
            "period": 0,
            "costAmount": 11998.5,
            "energyAmount": 1168.851
        }
    ],
    "currentCostsElecTimestamp": 1612126162,
    "currentCostsGas": [
        {
            "commodityType": "GAS_ENERGY",
            "duration": "DAY",
            "period": 0,
            "costAmount": 129.0,
            "energyAmount": 48.19
        },
        {
            "commodityType": "GAS_ENERGY",
            "duration": "WEEK",
            "period": 0,
            "costAmount": 678.0,
            "energyAmount": 236.859
        },
        {
            "commodityType": "GAS_ENERGY",
            "duration": "MONTH",
            "period": 0,
            "costAmount": 4267.0,
            "energyAmount": 1618.241
        }
    ],
    "currentCostsGasTimestamp": 1612126162,
    "prePayDebtList": null,
    "prePayDebtTimestamp": 0,
    "billingMode": [
        {
            "billingMode": "CREDIT",
            "commodityType": "ELECTRICITY",
            "valueAvailable": true
        },
        {
            "billingMode": "CREDIT",
            "commodityType": "GAS_ENERGY",
            "valueAvailable": true
        }
    ],
    "billingModeTimestamp": 1612126162,
    "budgetRagStatusDetails": [
        {
            "currDay": "GREEN",
            "yesterDay": "RED",
            "currWeek": "GREEN",
            "lastWeek": "GREEN",
            "currMth": "GREEN",
            "lastMth": "GREEN",
            "thisYear": "GREEN",
            "valueAvailable": true,
            "commodityType": "ELECTRICITY"
        },
        {
            "currDay": "GREEN",
            "yesterDay": "GREEN",
            "currWeek": "GREEN",
            "lastWeek": "GREEN",
            "currMth": "GREEN",
            "lastMth": "GREEN",
            "thisYear": "GREEN",
            "valueAvailable": true,
            "commodityType": "GAS_ENERGY"
        }
    ],
    "budgetRagStatusDetailsTimestamp": 1612126162,
    "budgetSettingDetails": [
        {
            "valueAvailable": true,
            "energyAmount": 1200.0,
            "costAmount": 1200.0,
            "budgetToC": 1612099325,
            "commodityType": "ELECTRICITY"
        },
        {
            "valueAvailable": true,
            "energyAmount": 420.0,
            "costAmount": 420.0,
            "budgetToC": 1612099325,
            "commodityType": "GAS_ENERGY"
        }
    ],
    "budgetSettingDetailsTimestamp": 1612126162,
    "setPoints": {
        "daySetPoint": {
            "temperatureSetPoint": 190,
            "timeOfChange": 0
        },
        "nightSetPoint": {
            "temperatureSetPoint": 150,
            "timeOfChange": 0
        }
    },
    "seasonalAdjustments": [
        {
            "valueAvailable": true,
            "commodityType": "ELECTRICITY",
            "adjustment": true,
            "timeOfChange": 1567102601
        },
        {
            "valueAvailable": true,
            "commodityType": "GAS_ENERGY",
            "adjustment": true,
            "timeOfChange": 1567102601
        }
    ]
}
*/

type PeriodicMeterDataConsumption struct {
	CommodityType    string  `json:"commodityType"`
	ReadingTime      int64   `json:"readingTime"`
	TotalConsumption float64 `json:"totalConsumption"`
	ValueAvailable   bool    `json:"valueAvailable"`
}
type PeriodicMeterDataSupplyStatus struct {
	CommodityType string `json:"commodityType"`
	SupplyStatus  string `json:"supplyStatus"`
}
type PeriodicMeterDataBillToDate struct {
	CommodityType  string  `json:"commodityType"`
	BillToDate     float64 `json:"billToDate"`
	ValidUTC       int64   `json:"validUTC"`
	StartUTC       int64   `json:"startUTC"`
	Duration       float64 `json:"duration"`
	ValueAvailable bool    `json:"valueAvailable"`
}
type PeriodicMeterDataActiveTariff struct {
	CommodityType       string  `json:"commodityType"`
	ValueAvailable      bool    `json:"valueAvailable"`
	NextTariffStartTime float64 `json:"nextTariffStartTime"`
	ActiveTariffPrice   float64 `json:"activeTariffPrice"`
	NextTariffPrice     float64 `json:"nextTariffPrice"`
	NextPriceAvailable  bool    `json:"nextPriceAvailable"`
}
type PeriodicMeterDataCurrentCost struct {
	CommodityType string  `json:"commodityType"`
	Duration      string  `json:"duration"`
	Period        float64 `json:"period"`
	CostAmount    float64 `json:"costAmount"`
	EnergyAmount  float64 `json:"energyAmount"`
}
type PeriodicMeterDataBillingMode struct {
	BillingMode    string `json:"billingMode"`
	CommodityType  string `json:"commodityType"`
	ValueAvailable bool   `json:"valueAvailable"`
}
type PeriodicMeterDataBudgetRagStatusDetails struct {
	CurrDay        string `json:"currDay"`
	YesterDay      string `json:"yesterDay"`
	CurrWeek       string `json:"currWeek"`
	LastWeek       string `json:"lastWeek"`
	CurrMth        string `json:"currMth"`
	LastMth        string `json:"lastMth"`
	ThisYear       string `json:"thisYear"`
	ValueAvailable bool   `json:"valueAvailable"`
	CommodityType  string `json:"commodityType"`
}
type PeriodicMeterDataBudgetSettingDetails struct {
	ValueAvailable bool    `json:"valueAvailable"`
	EnergyAmount   float64 `json:"energyAmount"`
	CostAmount     float64 `json:"costAmount"`
	BudgetToC      float64 `json:"budgetToC"`
	CommodityType  string  `json:"commodityType"`
}
type PeriodicMeterDataSetPointsPoint struct {
	TemperatureSetPoint float64 `json:"temperatureSetPoint"`
	TimeOfChange        int64   `json:"timeOfChange"`
}
type PeriodicMeterDataSetPoints struct {
	Day   PeriodicMeterDataSetPointsPoint `json:"daySetPoint"`
	Night PeriodicMeterDataSetPointsPoint `json:"nightSetPoint"`
}
type PeriodicMeterDataSeasonalAdjustments struct {
	ValueAvailable bool   `json:"valueAvailable"`
	CommodityType  string `json:"commodityType"`
	Adjustment     bool   `json:"adjustment"`
	TimeOfChange   int64  `json:"timeOfChange"`
}
type PeriodicMeterData struct {
	TTL                             int64                                     `json:"ttl"`
	LatestUTC                       int64                                     `json:"latestUtc"`
	ID                              string                                    `json:"id"`
	TotalConsumptionList            []PeriodicMeterDataConsumption            `json:"totalConsumptionList"`
	TotalConsumptionTimestamp       int64                                     `json:"totalConsumptionTimestamp"`
	SupplyStatusList                []PeriodicMeterDataSupplyStatus           `json:"supplyStatusList"`
	SupplyStatusTimestamp           int64                                     `json:"supplyStatusTimestamp"`
	BillToDateList                  []PeriodicMeterDataBillToDate             `json:"billToDateList"`
	BillToDateTimestamp             int64                                     `json:"billToDateTimestamp"`
	ActiveTariffList                []PeriodicMeterDataActiveTariff           `json:"activeTariffList"`
	ActiveTariffTimestamp           int64                                     `json:"activeTariffTimestamp"`
	CurrentCostsElec                []PeriodicMeterDataCurrentCost            `json:"currentCostsElec"`
	CurrentCostsElecTimestamp       int64                                     `json:"currentCostsElecTimestamp"`
	CurrentCostsGas                 []PeriodicMeterDataCurrentCost            `json:"currentCostsGas"`
	CurrentCostsGasTimestamp        int64                                     `json:"currentCostsGasTimestamp"`
	PrePayDebtList                  []string                                  `json:"prePayDebtList"`
	PrePayDebtTimestamp             int64                                     `json:"prePayDebtTimestamp"`
	BillingMode                     []PeriodicMeterDataBillingMode            `json:"billingMode"`
	BillingModeTimestamp            int64                                     `json:"billingModeTimestamp"`
	BudgetRagStatusDetails          []PeriodicMeterDataBudgetRagStatusDetails `json:"budgetRagStatusDetails"`
	BudgetRagStatusDetailsTimestamp int64                                     `json:"budgetRagStatusDetailsTimestamp"`
	BudgetSettingDetails            []PeriodicMeterDataBudgetSettingDetails   `json:"budgetSettingDetails"`
	BudgetSettingDetailsTimestamp   int64                                     `json:"budgetSettingDetailsTimestamp"`
	SetPoints                       PeriodicMeterDataSetPoints                `json:"setPoints"`
	SeasonalAdjustments             []PeriodicMeterDataSeasonalAdjustments    `json:"seasonalAdjustments"`
}

// GetPeriodicMeterData retrieves periodic meter data
func GetPeriodicMeterData(accessToken, systemID string) (PeriodicMeterData, error) {
	// Define request endpoint URL
	safeSystemID := url.QueryEscape(systemID)
	requestURL := fmt.Sprintf("/api/userapi/system/smets2-periodic-data/%s", safeSystemID)

	// Define request headers
	requestHeaders := map[string][]string{
		"Authorization": {"Bearer " + accessToken},
	}

	// Get data from geo API
	body, err := makeRequest(requestURL, "GET", nil, requestHeaders, nil)
	if err != nil {
		return PeriodicMeterData{}, err
	}
	//fmt.Println("Response: " + string(body))

	var periodicMeterData PeriodicMeterData
	err = json.Unmarshal(body, &periodicMeterData)
	if err != nil {
		log.Fatal(err)
		return PeriodicMeterData{}, err
	}

	return periodicMeterData, err
}

/*
{
    "latestUtc": 1612126140,
    "id": "ef9272fd-2a72-4306-a96f-7d624c0fbd31",
    "power": [
        {
            "type": "ELECTRICITY",
            "watts": 887,
            "valueAvailable": true
        },
        {
            "type": "GAS_ENERGY",
            "watts": 0,
            "valueAvailable": true
        }
    ],
    "powerTimestamp": 1612126140,
    "localTime": 1612126140,
    "localTimeTimestamp": 1612126140,
    "creditStatus": null,
    "creditStatusTimestamp": 0,
    "remainingCredit": null,
    "remainingCreditTimestamp": 0,
    "zigbeeStatus": {
        "electricityClusterStatus": "CONNECTED",
        "gasClusterStatus": "CONNECTED",
        "hanStatus": "CONNECTED",
        "networkRssi": -46
    },
    "zigbeeStatusTimestamp": 1612126140,
    "emergencyCredit": null,
    "emergencyCreditTimestamp": 0,
    "systemStatus": [
        {
            "component": "DISPLAY",
            "statusType": "STATUS_OK",
            "systemErrorCode": "ERROR_CODE_NONE",
            "systemErrorNumber": 0
        },
        {
            "component": "ZIGBEE",
            "statusType": "STATUS_OK",
            "systemErrorCode": "ERROR_CODE_NONE",
            "systemErrorNumber": 0
        },
        {
            "component": "ELECTRICITY",
            "statusType": "STATUS_OK",
            "systemErrorCode": "ERROR_CODE_NONE",
            "systemErrorNumber": 0
        },
        {
            "component": "GAS",
            "statusType": "STATUS_OK",
            "systemErrorCode": "ERROR_CODE_NONE",
            "systemErrorNumber": 0
        }
    ],
    "systemStatusTimestamp": 1612126140,
    "temperature": 0.0,
    "temperatureTimestamp": 0,
    "ttl": 120
}
*/
type LiveMeterDataPower struct {
	Type           string  `json:"type"`
	Watts          float64 `json:"watts"`
	ValueAvailable bool    `json:"valueAvailable"`
}
type LiveMeterDataZigbeeStatus struct {
	ElectricityClusterStatus string  `json:"electricityClusterStatus"`
	GasClusterStatus         string  `json:"gasClusterStatus"`
	HanStatus                string  `json:"hanStatus"`
	NetworkRssi              float64 `json:"networkRssi"`
}
type LiveMeterDataSystemStatus struct {
	Component         string  `json:"component"`
	StatusType        string  `json:"statusType"`
	SystemErrorCode   string  `json:"systemErrorCode"`
	SystemErrorNumber float64 `json:"systemErrorNumber"`
}
type LiveMeterData struct {
	LatestUTC                int64                       `json:"latestUtc"`
	ID                       string                      `json:"id"`
	Power                    []LiveMeterDataPower        `json:"power"`
	PowerTimestamp           int64                       `json:"powerTimestamp"`
	LocalTime                int64                       `json:"localTime"`
	LocalTimeTimestamp       int64                       `json:"localTimeTimestamp"`
	CreditStatus             int64                       `json:"creditStatus"`
	CreditStatusTimestamp    int64                       `json:"creditStatusTimestamp"`
	RemainingCredit          int64                       `json:"remainingCredit"`
	RemainingCreditTimestamp int64                       `json:"remainingCreditTimestamp"`
	ZigbeeStatus             LiveMeterDataZigbeeStatus   `json:"zigbeeStatus"`
	ZigbeeStatusTimestamp    int64                       `json:"zigbeeStatusTimestamp"`
	EmergencyCredit          int64                       `json:"emergencyCredit"`
	EmergencyCreditTimestamp int64                       `json:"emergencyCreditTimestamp"`
	SystemStatus             []LiveMeterDataSystemStatus `json:"systemStatus"`
	SystemStatusTimestamp    int64                       `json:"systemStatusTimestamp"`
	Temperature              float64                     `json:"temperature"`
	TemperatureTimestamp     int64                       `json:"temperatureTimestamp"`
	TTL                      int64                       `json:"ttl"`
}

// GetLiveMeterData retrieves live meter data
func GetLiveMeterData(accessToken, systemID string) (LiveMeterData, error) {
	// Define request endpoint URL
	safeSystemID := url.QueryEscape(systemID)
	requestURL := fmt.Sprintf("/api/userapi/system/smets2-live-data/%s", safeSystemID)

	// Define request headers
	requestHeaders := map[string][]string{
		"Authorization": {"Bearer " + accessToken},
	}

	// Get data from geo API
	body, err := makeRequest(requestURL, "GET", nil, requestHeaders, nil)
	if err != nil {
		return LiveMeterData{}, err
	}
	//fmt.Println("Response: " + string(body))

	var liveMeterData LiveMeterData
	err = json.Unmarshal(body, &liveMeterData)
	if err != nil {
		log.Fatal(err)
		return LiveMeterData{}, err
	}

	return liveMeterData, err
}

// ConvertToKWH converts m3 to kWh
func ConvertToKWH(m3 float64) float64 {
	return (((m3 / 1000) * 39.5) * 1.02264) / 3.6
}

// makeRequest function
func makeRequest(url string, method string, params map[string]string, headers map[string][]string, body io.Reader) ([]byte, error) {
	if method == "" {
		method = "GET"
	}

	req, err := http.NewRequest(method, geoBaseURL+url, body)
	if err != nil {
		return nil, err
	}

	// Define the query
	q := req.URL.Query()
	// Add query parameters
	for key, value := range params {
		q.Add(key, value)
	}
	req.URL.RawQuery = q.Encode()
	req.Header = headers

	// Make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	// Defer the close and handle any errors
	defer func() {
		cerr := resp.Body.Close()
		if err == nil {
			err = cerr
		}
	}()
	if err != nil {
		return nil, err
	}

	// Check response status
	if resp.Status == "200 OK" {
		// Read response body
		responseData, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return responseData, nil
	}
	return nil, err
}
