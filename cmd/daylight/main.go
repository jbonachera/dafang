package main

import (
	"log"
	"time"

	"github.com/jbonachera/dafang/daylight"
)

func main() {
	sensor, err := daylight.NewReporter()
	if err != nil {
		log.Fatal(err)
	}
	for range time.Tick(1 * time.Second) {
		log.Printf("Light is %v", sensor.Percent())
	}
}
