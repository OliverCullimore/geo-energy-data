package main

import (
	"fmt"
	"github.com/olivercullimore/geo-energy-data/server"
	"github.com/olivercullimore/go-utils/env"
	"log"
)

func main() {
	fmt.Println("Starting Server")

	// Load environment variables from file
	err := env.Load(".env")
	if err != nil {
		log.Println(err)
	}

	// Run server using environment variables
	server.Run(
		env.Get("LIVE_DATA_FETCH_INTERVAL", "10"),
		env.Get("PERIODIC_DATA_FETCH_INTERVAL", "300"),
		env.Get("CONFIG_FILE", "/config/config.json"),
		env.Get("GEO_USER", ""),
		env.Get("GEO_PASS", ""),
		env.Get("CALORIFIC_VALUE", "39.5"),
		env.Get("INFLUXDB_HOST", ""),
		env.Get("INFLUXDB_PORT", ""),
		env.Get("INFLUXDB_ORG", ""),
		env.Get("INFLUXDB_BUCKET", ""),
		env.Get("INFLUXDB_TOKEN", ""),
		env.Get("HTTP_HOST", "0.0.0.0"),
		env.Get("HTTP_PORT", "8080"),
	)
}
