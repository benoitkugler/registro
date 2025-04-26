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
	"registro/controllers/files"
	"registro/controllers/inscriptions"
	"registro/controllers/logic"
	"registro/crypto"
	cp "registro/sql/camps"
	fs "registro/sql/files"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	devPtr := flag.Bool("dev", false, "utilise un mode de développement (mail, localhost)")
	flag.Parse()
	isDev := *devPtr

	asso, keys, dbCreds, fs, smtp, joomeo := loadEnvs(isDev)
	encrypter := crypto.NewEncrypter(keys.EncryptKey)
	fmt.Println("Loading env. -> OK.")

	// TODO: setup APIS
	helloasso := config.Helloasso{}

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

	backofficeCt, err := backoffice.NewController(db, encrypter, keys.Backoffice, fs, smtp, asso, joomeo, helloasso)
	check(err)

	directeursCt, err := directeurs.NewController(db, keys.EncryptKey, keys.Directeurs, fs, smtp, asso, joomeo)
	check(err)

	espacepersoCt := espaceperso.NewController(db, encrypter, smtp, asso, fs, joomeo)

	inscriptionsCt := inscriptions.NewController(db, encrypter, smtp, asso)

	equipiersCt := equipiers.NewController(db, encrypter, fs, joomeo)

	filesCt := files.NewController(db, encrypter, fs)

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
	setupRoutesFiles(e, filesCt)
	setupRoutesServices(e, espacepersoCt)
	setupClientApps(e)

	fmt.Println("Setup done. Starting...")

	adress := getAdress(isDev)
	e.Logger.Fatal(e.Start(adress))
}

func loadEnvs(devMode bool) (config.Asso, config.Keys, config.DB, fs.FileSystem, config.SMTP, config.Joomeo) {
	asso, err := config.NewAsso()
	check(err)

	keys, err := config.NewKeys()
	check(err)

	db, err := config.NewDB()
	check(err)

	root := os.Getenv("FILES_ROOT")
	if root == "" {
		log.Fatal("missing env. FILES_ROOT (files directory)")
	}
	// check that the dir exists
	dir, err := os.Stat(root)
	check(err)
	if !dir.IsDir() {
		log.Fatal("invalid FILES_ROOT", root)
	}
	fmt.Println("using FILES_ROOT:", root)
	fileSystem := fs.NewFileSystem(root)

	smtp, err := config.NewSMTP(!devMode)
	check(err)

	joomeo, err := config.NewJoomeo()
	check(err)

	return asso, keys, db, fileSystem, smtp, joomeo
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

// désactive le référencement
func noIndex(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Set("X-Robots-Tag", "noindex")
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
	e.GET("/backoffice/*", serve("static/backoffice/index.html"), middleware.Gzip(), noCache)

	e.GET("/directeurs", serve("static/directeurs/index.html"), middleware.Gzip(), noCache)
	e.GET("/directeurs/*", serve("static/directeurs/index.html"), middleware.Gzip(), noCache)

	e.GET(directeurs.EndpointEquipier, serve("static/equipier/index.html"), middleware.Gzip(), noCache, noIndex)
	e.GET(directeurs.EndpointEquipier+"/*", serve("static/equipier/index.html"), middleware.Gzip(), noCache, noIndex)

	e.GET(logic.EndpointEspacePerso, serve("static/espaceperso/index.html"), middleware.Gzip(), noCache, noIndex)
	e.GET(logic.EndpointEspacePerso+"/*", serve("static/espaceperso/index.html"), middleware.Gzip(), noCache, noIndex)

	e.GET(inscriptions.EndpointInscription, serve("static/inscription/index.html"), middleware.Gzip(), noCache)
	e.GET(inscriptions.EndpointInscription+"/*", serve("static/inscription/index.html"), middleware.Gzip(), noCache)

	e.GET("/services", serve("static/services/index.html"), middleware.Gzip(), noCache)
	e.GET("/services/*", serve("static/services/index.html"), middleware.Gzip(), noCache)

	// global static files used by frontend apps
	e.Group("/static", middleware.Gzip(), cacheStatic).Static("/*", "static")
}

func setupRoutesFiles(e *echo.Echo, filesCt *files.Controller) {
	// every endpoint expected a key=<idCrypted> query param
	e.GET("/api/v1/documents", filesCt.Get)
	e.GET("/api/v1/documents/miniature", filesCt.GetMiniature)
}
