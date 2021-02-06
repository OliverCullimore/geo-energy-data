[![License](https://img.shields.io/github/license/OliverCullimore/geo-energy-data?style=for-the-badge)](https://github.com/OliverCullimore/geo-energy-data)
[![Build Status](https://img.shields.io/github/workflow/status/OliverCullimore/geo-energy-data/ci?logo=github&style=for-the-badge)](https://github.com/OliverCullimore/geo-energy-data)
[![Docker Pulls](https://img.shields.io/docker/pulls/olivercullimore/geo-energy-data?logo=docker&style=for-the-badge)](https://hub.docker.com/r/olivercullimore/geo-energy-data)

# geo Energy Data

A Go application that periodically gets energy data from the geotogether.com API and stores it in an InfluxDB database.


## Prerequisites

* A geo smart meter display with a [WiFi module](https://www.geotogether.com/consumer/product/wifi-module/) installed and have set up an account in the geo Home app and linked your smart meter display to your account.

* An InfluxDB database server with a bucket and access token set up to use for this application.

## Set Up
> Ensure you have met the prerequisites above before continuing

### Docker

[Docker Hub](https://hub.docker.com/repository/docker/olivercullimore/geo-energy-data)

### Standalone

1. Rename/Copy the .env.example file to .env

2. Configure the geotogether environment variables
    * Replace `YOUR-GEOTOGETHER-USER` in the .env file with your geo Home app username 
    * Replace `YOUR-GEOTOGETHER-PASS` in the .env file with your geo Home app password


3. Configure the InfluxDB environment variables
    * Replace `YOUR-INFLUXDB-HOST` in the .env file with your InfluxDB hostname
    * Replace `YOUR-INFLUXDB-PORT` in the .env file with your InfluxDB port
    * Replace `YOUR-INFLUXDB-ORGANIZATION` in the .env file with your InfluxDB organization
    * Replace `YOUR-INFLUXDB-BUCKET` in the .env file with your InfluxDB bucket name
    * Replace `YOUR-INFLUXDB-TOKEN` in the .env file with your InfluxDB access token


4. Start the application by running `make run`