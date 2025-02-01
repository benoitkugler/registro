package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"registro/config"
	"registro/controllers/central"
	"registro/crypto"
	"registro/sql/files"

	"github.com/labstack/echo/v4"
)

func main() {
	devPtr := flag.Bool("dev", false, "utilise les crédences de développement")
	flag.Parse()
	isDev := *devPtr

	fmt.Println("Running in mode dev:", isDev)

	_, encrypter, dbCreds, fs := loadEnvs()

	db, err := dbCreds.ConnectPostgres()
	check(err)
	check(db.Ping())
	fmt.Println("\tPing DB -> OK.")

	// TODO: setup APIS
	ct, err := central.NewController(db, encrypter, fs, config.SMTP{}, config.Joomeo{}, config.Helloasso{})
	check(err)

	e := echo.New()
	e.HideBanner = true
	e.HTTPErrorHandler = func(err error, c echo.Context) {
		err = echo.NewHTTPError(400, err.Error())
		e.DefaultHTTPErrorHandler(err, c)
	}

	setupRoutesCentral(e, ct)

	adress := getAdress(isDev)

	fmt.Println("\tSetup done.")

	e.Logger.Fatal(e.Start(adress))
}

func loadEnvs() (config.Asso, crypto.Encrypter, config.DB, files.FileSystem) {
	asso, err := config.NewAsso()
	check(err)

	serverKey := os.Getenv("SERVER_KEY")
	if serverKey == "" {
		log.Fatal("missing encryption key")
	}
	encrypter := crypto.NewEncrypter(serverKey)

	db, err := config.NewDB()
	check(err)

	fs := os.Getenv("FILES_ROOT")
	if fs == "" {
		log.Fatal("missing encryption key")
	}
	fileSystem := files.NewFileSystem(fs)

	return asso, encrypter, db, fileSystem
}

func getAdress(dev bool) string {
	var adress string
	if dev {
		adress = "localhost:1323"
	} else {
		// TODO: adapt to infomaniak
		// alwaysdata use IP and PORT env var
		host := os.Getenv("IP")
		port, err := strconv.Atoi(os.Getenv("PORT"))
		if err != nil {
			log.Fatal("No PORT found", err)
		}
		if strings.Count(host, ":") >= 2 { // ipV6 -> besoin de crochet
			host = "[" + host + "]"
		}
		adress = fmt.Sprintf("%s:%d", host, port)
	}
	return adress
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
