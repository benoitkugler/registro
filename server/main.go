package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"registro/config"
	"registro/controllers/backoffice"
	"registro/controllers/espaceperso"
	"registro/controllers/inscriptions"
	"registro/crypto"
	"registro/sql/files"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	devPtr := flag.Bool("dev", false, "utilise les crédences de développement")
	flag.Parse()
	isDev := *devPtr

	asso, encrypter, dbCreds, fs, smtp := loadEnvs(isDev)
	fmt.Println("Loading env. -> OK.")
	// TODO: setup APIS
	joomeo, helloasso := config.Joomeo{}, config.Helloasso{}

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

	backofficeCt, err := backoffice.NewController(db, encrypter, fs, smtp, joomeo, helloasso)
	check(err)

	inscriptionsCt := inscriptions.NewController(db, encrypter, smtp, asso)

	espacepersoCt := espaceperso.NewController(db, encrypter, smtp, asso)

	if isDev {
		fmt.Println("Running in dev mode")

		e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			AllowMethods:  append(middleware.DefaultCORSConfig.AllowMethods, http.MethodOptions),
			AllowHeaders:  []string{"Authorization", "Content-Type", "Access-Control-Allow-Origin"},
			ExposeHeaders: []string{"Content-Disposition"},
		}))
		fmt.Println("\tenabling CORS")

		token, err := backofficeCt.NewToken(false)
		check(err)
		fmt.Println("\tcentral dev token:", token)
	}

	setupRoutesBackoffice(e, backofficeCt)
	setupRoutesInscriptions(e, inscriptionsCt)
	setupRoutesEspaceperso(e, espacepersoCt)

	adress := getAdress(isDev)

	fmt.Println("Setup done.")

	e.Logger.Fatal(e.Start(adress))
}

func loadEnvs(devMode bool) (config.Asso, crypto.Encrypter, config.DB, files.FileSystem, config.SMTP) {
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

	smtp, err := config.NewSMTP(!devMode)
	check(err)

	return asso, encrypter, db, fileSystem, smtp
}

func getAdress(devMode bool) string {
	var adress string
	if devMode {
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
