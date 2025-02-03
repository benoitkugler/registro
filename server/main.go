package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"net/http"
	
	"registro/config"
	"registro/controllers/central"
	"registro/crypto"
	"registro/sql/files"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	devPtr := flag.Bool("dev", false, "utilise les crédences de développement")
	flag.Parse()
	isDev := *devPtr

	_, encrypter, dbCreds, fs := loadEnvs()

	db, err := dbCreds.ConnectPostgres()
	check(err)
	check(db.Ping())
	fmt.Println("Ping DB -> OK.")

	e := echo.New()
	e.HideBanner = true
	e.HTTPErrorHandler = func(err error, c echo.Context) {
		err = echo.NewHTTPError(400, err.Error())
		e.DefaultHTTPErrorHandler(err, c)
	}

	// TODO: setup APIS
	ct, err := central.NewController(db, encrypter, fs, config.SMTP{}, config.Joomeo{}, config.Helloasso{})
	check(err)
	
	if isDev {
		fmt.Println("Running in dev mode")
		
		e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			AllowMethods:  append(middleware.DefaultCORSConfig.AllowMethods, http.MethodOptions),
			AllowHeaders:  []string{"Authorization", "Content-Type", "Access-Control-Allow-Origin"},
			ExposeHeaders: []string{"Content-Disposition"},
		}))
		fmt.Println("\tenabling CORS")
	
		token, err := ct.NewToken(false)
		check(err)
		fmt.Println("\tcentral dev token:", token)
	}


	setupRoutesCentral(e, ct)

	adress := getAdress(isDev)

	fmt.Println("Setup done.")

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
