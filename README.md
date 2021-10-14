[![License](https://img.shields.io/github/license/OliverCullimore/geo-energy-data?style=for-the-badge)](https://github.com/OliverCullimore/geo-energy-data)
[![Build Status](https://img.shields.io/github/workflow/status/OliverCullimore/geo-energy-data/ci?logo=github&style=for-the-badge)](https://github.com/OliverCullimore/geo-energy-data)
[![Docker Pulls](https://img.shields.io/docker/pulls/olivercullimore/geo-energy-data?logo=docker&style=for-the-badge)](https://hub.docker.com/r/olivercullimore/geo-energy-data)

# geo Energy Data

A Go application that periodically retrieves energy data from the geotogether.com API and stores it in an InfluxDB 2.0 database and/or exposes it as an API for another application to use.


## Prerequisites

* A geo smart meter display with a [WiFi module](https://www.geotogether.com/consumer/product/wifi-module/) installed or a [geo Hub + LED Sensor](https://www.geotogether.com/consumer/product/hub_led_sensor/) and have set up an account in the geo Home app and [linked](https://touchbutton.support.geotogether.com/en/support/solutions/articles/7000051019-linking-the-geo-home-app-to-your-trio) your smart meter display to your account.
> This may also work with a [geo Hub + LED Sensor](https://www.geotogether.com/consumer/product/hub_led_sensor/) that's been [linked](https://electricity.support.geotogether.com/en/support/solutions/articles/7000058728-how-do-i-link-to-my-hub-and-app) to your geo Home app account but this is currently untested.


* A host with Docker [set-up and configured](https://www.digitalocean.com/community/tutorial_collections/how-to-install-and-use-docker).


* (if using InfluxDB mode) An [InfluxDB OSS 2.0](https://docs.influxdata.com/influxdb/v2.0/install/) database server with a bucket and access token set up to use for this application (other versions of InfluxDB may work, but are not tested).

## Usage

> Ensure you have met the prerequisites above before continuing

### Quick start for InfluxDB mode

This mode will periodically retrieve energy data and store it in an InfluxDB 2.0 database.

Run the following command after replacing the following parts with appropriate values `YOUR-GEOHOMEAPP-USER`, `YOUR-GEOHOMEAPP-PASS`, `YOUR-INFLUXDB-HOST`, `YOUR-INFLUXDB-PORT`, `YOUR-INFLUXDB-ORG`, `YOUR-INFLUXDB-BUCKET`, `YOUR-INFLUXDB-TOKEN`, please refer to the Environment variables configuration section below for further details on the values to enter.

```bash
docker run -d \
--name geo-energy-data \
-e LIVE_DATA_FETCH_INTERVAL=10 \
-e PERIODIC_DATA_FETCH_INTERVAL=30 \
-e GEO_USER="YOUR-GEOHOMEAPP-USER" \
-e GEO_PASS="YOUR-GEOHOMEAPP-PASS" \
-e INFLUXDB_HOST="YOUR-INFLUXDB-HOST" \
-e INFLUXDB_PORT="YOUR-INFLUXDB-PORT" \
-e INFLUXDB_ORG="YOUR-INFLUXDB-ORG" \
-e INFLUXDB_BUCKET="YOUR-INFLUXDB-BUCKET" \
-e INFLUXDB_TOKEN="YOUR-INFLUXDB-TOKEN" \
-e CONFIG_FILE=/config/config.json \
-e ENABLE_API=false \
-e ENABLE_INFLUXDB=true \
-e DEBUG_MODE=false \
--restart unless-stopped \
-v geo_energy_data:/config \
olivercullimore/geo-energy-data:v2.0.0
```

### Quick start for API mode

This mode will provide an API (please refer to the API details section below for further details) which can be used to request energy data and use it in another application.

Run the following command after replacing the following parts with appropriate values `YOUR-GEOHOMEAPP-USER`, `YOUR-GEOHOMEAPP-PASS`, `YOUR-API-KEY`, please refer to the Environment variables configuration section below for further details on the values to enter.

```bash
docker run -d \
--name geo-energy-data \
-e GEO_USER="YOUR-GEOHOMEAPP-USER" \
-e GEO_PASS="YOUR-GEOHOMEAPP-PASS" \
-e API_KEY="YOUR-API-KEY" \
-e CONFIG_FILE=/config/config.json \
-e ENABLE_API=true \
-e ENABLE_INFLUXDB=false \
-e DEBUG_MODE=false \
-p 8080:80 \
--restart unless-stopped \
-v geo_energy_data:/config \
olivercullimore/geo-energy-data:v2.0.0
```

### Quick start for both InfluxDB and API mode

This mode will periodically retrieve energy data and store it in an InfluxDB 2.0 database and also provide an API (please refer to the API details section below for further details) which can be used to request energy data and use it in another application.

Run the following command after replacing the following parts with appropriate values `YOUR-GEOHOMEAPP-USER`, `YOUR-GEOHOMEAPP-PASS`, `YOUR-INFLUXDB-HOST`, `YOUR-INFLUXDB-PORT`, `YOUR-INFLUXDB-ORG`, `YOUR-INFLUXDB-BUCKET`, `YOUR-INFLUXDB-TOKEN`, `YOUR-API-KEY`, please refer to the Environment variables configuration section below for further details on the values to enter.

```bash
docker run -d \
--name geo-energy-data \
-e LIVE_DATA_FETCH_INTERVAL=10 \
-e PERIODIC_DATA_FETCH_INTERVAL=30 \
-e GEO_USER="YOUR-GEOHOMEAPP-USER" \
-e GEO_PASS="YOUR-GEOHOMEAPP-PASS" \
-e INFLUXDB_HOST="YOUR-INFLUXDB-HOST" \
-e INFLUXDB_PORT="YOUR-INFLUXDB-PORT" \
-e INFLUXDB_ORG="YOUR-INFLUXDB-ORG" \
-e INFLUXDB_BUCKET="YOUR-INFLUXDB-BUCKET" \
-e INFLUXDB_TOKEN="YOUR-INFLUXDB-TOKEN" \
-e API_KEY="YOUR-API-KEY" \
-e CONFIG_FILE=/config/config.json \
-e ENABLE_API=true \
-e ENABLE_INFLUXDB=true \
-e DEBUG_MODE=false \
-p 8080:80 \
--restart unless-stopped \
-v geo_energy_data:/config \
olivercullimore/geo-energy-data:v2.0.0
```

## Environment variables configuration

|            Variable            |                                               Description                                                                                     |
| :----------------------------: | --------------------------------------------------------------------------------------------------------------------------------------------- |
| `GEO_USER`                     | Specify the geo Home app username to use                                                                                                      |
| `GEO_PASS`                     | Specify the geo Home app password to use                                                                                                      |
| `LIVE_DATA_FETCH_INTERVAL`     | Specify the live data fetch interval to use in seconds e.g. 10 (only if ENABLE_INFLUXDB is set to true)                                       |
| `PERIODIC_DATA_FETCH_INTERVAL` | Specify the periodic data fetch interval to use in seconds e.g. 30 for 30 seconds, 300 for 5 minutes (only if ENABLE_INFLUXDB is set to true) |
| `INFLUXDB_HOST`                | Specify the InfluxDB host domain/IP to use including the protocol e.g. http://192.168.1.50 (only if ENABLE_INFLUXDB is set to true)           |
| `INFLUXDB_PORT`                | Specify the InfluxDB port number to use e.g. 8086 (only if ENABLE_INFLUXDB is set to true)                                                    |
| `INFLUXDB_ORG`                 | Specify the InfluxDB organization to use (only if ENABLE_INFLUXDB is set to true)                                                             |
| `INFLUXDB_BUCKET`              | Specify the InfluxDB bucket to use (only if ENABLE_INFLUXDB is set to true)                                                                   |
| `INFLUXDB_TOKEN`               | Specify the InfluxDB token to use (only if ENABLE_INFLUXDB is set to true)                                                                    |
| `API_KEY`                      | Specify the API key to use. This can be any value e.g.  (only if ENABLE_API is set to true)                                                   |
| `CONFIG_FILE`                  | Specify the config file path to use. Leave blank to use default config file path of `/config/config.json`                                     |
| `ENABLE_API`                   | Specify if the API functionality should be enabled. Leave blank to use default value of `false`                                               |
| `ENABLE_INFLUXDB`              | Specify if the InfluxDB functionality should be enabled. Leave blank to use default value of `true`                                           |
| `DEBUG_MODE`                   | Specify if the debug mode should be enabled. Leave blank to use default value of `false`                                                      |

## Troubleshooting

|      Message       |                                       Description                                          |
| :----------------: | ------------------------------------------------------------------------------------------ |
| No system ID found | Please ensure you have linked your smart meter display to your account in the geo Home app |

> If your issue isn't listed in the troubleshooting table above please try running the latest version of the docker image, if the issue persists please open an issue

## API details

### Authorization

Header `X-Api-Key: YOUR-API-KEY`

### Endpoints

GET `/api/status` Health check

GET `/api/beta/currentusage` Get current usage data

GET `/api/beta/meterreadings` Get meter readings data

GET `/api/beta/live` Get live data

GET `/api/beta/periodic` Get periodic data

### Example request

Replace the following parts with appropriate values:
`YOUR-HOST-ADDRESS` The hostname/IP of your Docker host e.g. 192.168.1.50
`YOUR-API-KEY` The API key set in your environment variables above

```curl
curl --location --request GET 'http://YOUR-HOST-ADDRESS:8080/api/beta/currentusage' \
--header 'X-Api-Key: YOUR-API-KEY'
```