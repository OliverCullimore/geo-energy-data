[![License](https://img.shields.io/github/license/OliverCullimore/geo-energy-data?style=for-the-badge)](https://github.com/OliverCullimore/geo-energy-data)
[![Build Status](https://img.shields.io/github/workflow/status/OliverCullimore/geo-energy-data/ci?logo=github&style=for-the-badge)](https://github.com/OliverCullimore/geo-energy-data)
[![Docker Pulls](https://img.shields.io/docker/pulls/olivercullimore/geo-energy-data?logo=docker&style=for-the-badge)](https://hub.docker.com/r/olivercullimore/geo-energy-data)

# geo Energy Data

A Go application that periodically retrieves energy data from the geotogether.com API and stores it in an InfluxDB 2.0 database.


## Prerequisites

* A geo smart meter display with a [WiFi module](https://www.geotogether.com/consumer/product/wifi-module/) installed and have set up an account in the geo Home app and linked your smart meter display to your account.

* An [InfluxDB OSS 2.0](https://docs.influxdata.com/influxdb/v2.0/install/) database server with a bucket and access token set up to use for this application (other versions of InfluxDB may work, but are not tested).

## Usage

> Ensure you have met the prerequisites above before continuing

### Quick start

This quick start is distributed as a docker image, please ensure you have docker [set-up and configured](https://www.digitalocean.com/community/tutorial_collections/how-to-install-and-use-docker) before continuing.

Run the following command after replacing the following parts with appropriate values `YOUR-GEOHOMEAPP-USER`, `YOUR-GEOHOMEAPP-PASS`, `YOUR-INFLUXDB-HOST`, `YOUR-INFLUXDB-PORT`, `YOUR-INFLUXDB-ORG`, `YOUR-INFLUXDB-BUCKET`, `YOUR-INFLUXDB-TOKEN`, please refer to the Environment variables configuration section below for further details on the values to enter.

```bash
docker run -d \
--name geo-energy-data \
-e LIVE_DATA_FETCH_INTERVAL=10 \
-e PERIODIC_DATA_FETCH_INTERVAL=30 \
-e GEO_USER=YOUR-GEOHOMEAPP-USER \
-e GEO_PASS=YOUR-GEOHOMEAPP-PASS \
-e INFLUXDB_HOST=YOUR-INFLUXDB-HOST \
-e INFLUXDB_PORT=YOUR-INFLUXDB-PORT \
-e INFLUXDB_ORG=YOUR-INFLUXDB-ORG \
-e INFLUXDB_BUCKET=YOUR-INFLUXDB-BUCKET \
-e INFLUXDB_TOKEN=YOUR-INFLUXDB-TOKEN \
-e CONFIG_FILE=/config/config.json \
-e DEBUG_MODE=false \
--restart unless-stopped \
-v geo_energy_data:/config \
olivercullimore/geo-energy-data
```

## Environment variables configuration

|            Variable            |                                               Description                                                 |
| :----------------------------: | --------------------------------------------------------------------------------------------------------- |
| `LIVE_DATA_FETCH_INTERVAL`     | Specify the live data fetch interval to use in seconds e.g. 10                                            |
| `PERIODIC_DATA_FETCH_INTERVAL` | Specify the periodic data fetch interval to use in seconds e.g. 30 for 30 seconds, 300 for 5 minutes      |
| `GEO_USER`                     | Specify the geo Home app username to use                                                                  |
| `GEO_PASS`                     | Specify the geo Home app password to use                                                                  |
| `INFLUXDB_HOST`                | Specify the InfluxDB host domain/IP to use including the protocol e.g. http://192.168.1.50                |
| `INFLUXDB_PORT`                | Specify the InfluxDB port number to use e.g. 8086                                                         |
| `INFLUXDB_ORG`                 | Specify the InfluxDB organization to use                                                                  |
| `INFLUXDB_BUCKET`              | Specify the InfluxDB bucket to use                                                                        |
| `INFLUXDB_TOKEN`               | Specify the InfluxDB token to use                                                                         |
| `CONFIG_FILE`                  | Specify the config file path to use. Leave blank to use default config file path of `/config/config.json` |
| `DEBUG_MODE`                   | Specify if the debug mode should be enabled. Leave blank to use default value of `false`                  |

## For development / running standalone

1. Rename/Copy the .env.example file to .env

2. Configure the geotogether environment variables
    * Replace `YOUR-GEOHOMEAPP-USER` in the .env file with your geo Home app username 
    * Replace `YOUR-GEOHOMEAPP-PASS` in the .env file with your geo Home app password


3. Configure the InfluxDB environment variables
    * Replace `YOUR-INFLUXDB-HOST` in the .env file with your InfluxDB hostname
    * Replace `YOUR-INFLUXDB-PORT` in the .env file with your InfluxDB port
    * Replace `YOUR-INFLUXDB-ORG` in the .env file with your InfluxDB organization
    * Replace `YOUR-INFLUXDB-BUCKET` in the .env file with your InfluxDB bucket name
    * Replace `YOUR-INFLUXDB-TOKEN` in the .env file with your InfluxDB access token


4. Start the application by running `make run`