package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"registro/config"
	"registro/controllers/backoffice"
	"registro/controllers/directeurs"
	equipiers "registro/controllers/equipier"
	"registro/controllers/espaceperso"
	"registro/controllers/inscriptions"
	"registro/controllers/logic"
	"registro/crypto"
	cp "registro/sql/camps"
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

	fmt.Println("Connecting to DB", dbCreds.Name, "at", dbCreds.Host, "...")
	db, err := dbCreds.ConnectPostgres()
	check(err)
	check(db.Ping())
	fmt.Println("Connecting DB -> OK.")

	e := echo.New()
	e.HideBanner = true
	e.HTTPErrorHandler = func(err error, c echo.Context) {
		err = echo.NewHTTPError(400, err.Error())
		e.DefaultHTTPErrorHandler(err, c)
	}

	backofficeCt, err := backoffice.NewController(db, encrypter, fs, smtp, asso, joomeo, helloasso)
	check(err)

	directeursCt, err := directeurs.NewController(db, encrypter, fs, smtp, asso, joomeo)
	check(err)

	espacepersoCt := espaceperso.NewController(db, encrypter, smtp, asso)

	inscriptionsCt := inscriptions.NewController(db, encrypter, smtp, asso)

	equipiersCt := equipiers.NewController(db, encrypter, joomeo)

	if isDev {
		fmt.Println("Running in dev mode :")

		e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			AllowMethods:  append(middleware.DefaultCORSConfig.AllowMethods, http.MethodOptions),
			AllowHeaders:  []string{"Authorization", "Content-Type", "Access-Control-Allow-Origin"},
			ExposeHeaders: []string{"Content-Disposition"},
		}))
		fmt.Println("\tenabling CORS")

		tokenB, err := backofficeCt.NewToken(false)
		check(err)
		fmt.Println("\tbackoffice dev token:", tokenB)

		// select and generate a directeur token
		camps, err := cp.SelectAllCamps(db)
		check(err)

		ids := camps.IDs()
		if len(ids) != 0 {
			camp := camps[ids[0]]
			tokenC, err := directeursCt.NewToken(camp.Id)
			check(err)
			b, _ := json.Marshal(logic.NewCampItem(camp))
			fmt.Println("\tdirecteurs dev env:")
			fmt.Printf("export const devCamp = %s;\nexport const devToken = %q\n", b, tokenC)
		}
	}

	setupRoutesBackoffice(e, backofficeCt)
	setupRoutesDirecteurs(e, directeursCt)
	setupRoutesEspaceperso(e, espacepersoCt)
	setupRoutesInscription(e, inscriptionsCt)
	setupRoutesEquipier(e, equipiersCt)

	setupClientApps(e)

	adress := getAdress(isDev)

	fmt.Println("Setup done.")

	e.Logger.Fatal(e.Start(adress))
}

func loadEnvs(devMode bool) (config.Asso, crypto.Encrypter, config.DB, files.FileSystem, config.SMTP) {
	asso, err := config.NewAsso()
	check(err)

	serverKey := os.Getenv("SERVER_KEY")
	if serverKey == "" {
		log.Fatal("missing env. SERVER_KEY (encryption key)")
	}
	encrypter := crypto.NewEncrypter(serverKey)

	db, err := config.NewDB()
	check(err)

	fs := os.Getenv("FILES_ROOT")
	if fs == "" {
		log.Fatal("missing env. FILES_ROOT (files directory)")
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

// Empêche le navigateur de mettre en cache
// pour avoir les dernières versions des fichiers statiques
// (essentiellement les builds .js)
func noCache(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Set("Cache-Control", "no-store")
		c.Response().Header().Set("Expires", "0")
		return next(c)
	}
}

// cacheStatic adopt a very aggressive caching policy, suitable
// for immutable content
func cacheStatic(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Set("Cache-Control", "max-age=31536000")
		return next(c)
	}
}

func setupClientApps(e *echo.Echo) {
	serve := func(path string) echo.HandlerFunc {
		return func(c echo.Context) error { return c.File(path) }
	}

	e.GET("/backoffice", serve("static/backoffice/index.html"), middleware.Gzip(), noCache)
	e.GET(inscriptions.EndpointInscription, serve("static/inscription/index.html"), middleware.Gzip(), noCache)
	e.GET(equipiers.EndpointEquipier, serve("static/equipier/index.html"), middleware.Gzip(), noCache)
	e.GET("/directeurs", serve("static/directeurs/index.html"), middleware.Gzip(), noCache)

	// global static files used by frontend apps
	e.Group("/static", middleware.Gzip(), cacheStatic).Static("/*", "static")
}
