package main

import (
	"fmt"
	"log"

	"registro/config"
)

func main() {
	asso, err := config.NewAsso()
	if err != nil {
		log.Fatal(err)
	}

	db, err := config.NewDB()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(asso, db)
}
