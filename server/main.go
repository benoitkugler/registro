package main

import (
	"log"

	"registro/config"
)

func main() {
	config, err := config.NewConfig()
	if err != nil {
		log.Fatal(err)
	}
	_ = config
}
