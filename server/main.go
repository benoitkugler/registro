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
	fsAPI "registro/controllers/files"
	"registro/controllers/inscriptions"
	"registro/controllers/services"
	"registro/crypto"
	"registro/generators/pdfcreator"
	"registro/logic"
	cp "registro/sql/camps"
	"registro/sql/files"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	devPtr := flag.Bool("dev", false, "utilise un mode de développement (mail, localhost)")
	flag.Parse()
	isDev := *devPtr

	asso, keys, dbCreds, smtp, directories, immich := loadEnvs(isDev)

	fmt.Println("Loading env. -> OK.")
	fmt.Println("\tASSO:", asso.Title)
	fmt.Println("\tFILES_DIR:", directories.Files)
	fmt.Println("\tASSETS_DIR:", directories.Assets)
	fmt.Println("\tCACHE_DIR:", directories.Cache)
	fmt.Println("\tMEDIA_DIR:", directories.Media)

	err := pdfcreator.Init(directories.Cache, directories.Assets)
	check(err)
	fmt.Println("Setting up pdfcreator -> OK.")

	// TODO: setup Dons, OnlinePaiement APIS
	helloasso := config.Helloasso{}

	fmt.Println("Connecting to DB", dbCreds.Name, "at", dbCreds.Host, "...")
	db, err := dbCreds.ConnectPostgres()
	check(err)
	check(db.Ping())
	fmt.Println("Connecting: OK.")

	fs := files.NewFileSystem(directories.Files)

	e := echo.New()
	e.HideBanner = true
	e.HTTPErrorHandler = func(err error, c echo.Context) {
		err = echo.NewHTTPError(400, err.Error())
		e.DefaultHTTPErrorHandler(err, c)
	}

	encrypter := crypto.NewEncrypter(keys.EncryptKey)

	backofficeCt, err := backoffice.NewController(db, encrypter, keys.Backoffice, keys.FondSoutien, fs, smtp, asso, immich, helloasso)
	check(err)

	directeursCt, err := directeurs.NewController(db, keys.EncryptKey, keys.Directeurs, fs, smtp, asso, immich)
	check(err)

	espacepersoCt := espaceperso.NewController(db, encrypter, smtp, asso, fs, immich)

	inscriptionsCt := inscriptions.NewController(db, encrypter, smtp, asso)

	equipiersCt := equipiers.NewController(db, encrypter, fs, immich)

	servicesCt := services.NewController(db, encrypter, smtp, asso)

	filesCt := fsAPI.NewController(db, encrypter, fs, asso)

	if isDev {
		fmt.Println("Running in dev mode :")

		e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			AllowMethods:  append(middleware.DefaultCORSConfig.AllowMethods, http.MethodOptions),
			AllowHeaders:  []string{"Authorization", "Content-Type", "Access-Control-Allow-Origin"},
			ExposeHeaders: []string{"Content-Disposition"},
		}))
		fmt.Println("\tenabling CORS")

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
	setupRoutesServices(e, servicesCt, espacepersoCt)
	setupRoutesMisc(e, filesCt)
	setupClientApps(e, asso.ID)

	if directories.Media != "" {
		e.Group("/media", cacheStatic).Static("/*", directories.Media)
	}

	fmt.Println("Setup done. Starting...")

	adress := getAdress(isDev)
	e.Logger.Fatal(e.Start(adress))
}

func loadEnvs(devMode bool) (config.Asso, config.Keys, config.DB, config.SMTP, config.Directories, config.Immich) {
	asso, err := config.NewAsso()
	check(err)

	keys, err := config.NewKeys()
	check(err)

	db, err := config.NewDB()
	check(err)

	smtp, err := config.NewSMTP(!devMode)
	check(err)

	dirs, err := config.NewDirectories()
	check(err)

	immich, err := config.NewImmich()
	check(err)

	return asso, keys, db, smtp, dirs, immich
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

func setupClientApps(e *echo.Echo, asso string) {
	// path must be relative to static/<asso>
	serveStaticAsso := func(path string) echo.HandlerFunc {
		return func(c echo.Context) error { return c.File("static/" + asso + path) }
	}

	e.GET("/backoffice", serveStaticAsso("/backoffice/index.html"), middleware.Gzip(), noCache)
	e.GET("/backoffice/*", serveStaticAsso("/backoffice/index.html"), middleware.Gzip(), noCache)

	e.GET("/directeurs", serveStaticAsso("/directeurs/index.html"), middleware.Gzip(), noCache)
	e.GET("/directeurs/*", serveStaticAsso("/directeurs/index.html"), middleware.Gzip(), noCache)

	e.GET(directeurs.EndpointEquipier, serveStaticAsso("/equipier/index.html"), middleware.Gzip(), noCache, noIndex)
	e.GET(directeurs.EndpointEquipier+"/*", serveStaticAsso("/equipier/index.html"), middleware.Gzip(), noCache, noIndex)

	e.GET(logic.EndpointEspacePerso, serveStaticAsso("/espaceperso/index.html"), middleware.Gzip(), noCache, noIndex)
	e.GET(logic.EndpointEspacePerso+"/*", serveStaticAsso("/espaceperso/index.html"), middleware.Gzip(), noCache, noIndex)

	e.GET(inscriptions.EndpointInscription, serveStaticAsso("/inscription/index.html"), middleware.Gzip(), noCache)
	e.GET(inscriptions.EndpointInscription+"/*", serveStaticAsso("/inscription/index.html"), middleware.Gzip(), noCache)

	// services also contains the index page and the CGUs
	e.GET("/", serveStaticAsso("/services/index.html"), middleware.Gzip(), noCache)
	e.GET("/cgu", serveStaticAsso("/services/index.html"), middleware.Gzip(), noCache)
	e.GET("/cgu/*", serveStaticAsso("/services/index.html"), middleware.Gzip(), noCache)
	e.GET("/services", serveStaticAsso("/services/index.html"), middleware.Gzip(), noCache)
	e.GET("/services/*", serveStaticAsso("/services/index.html"), middleware.Gzip(), noCache)

	// global static files used by frontend apps
	e.Group("/static", middleware.Gzip(), cacheStatic).Static("/*", "static/"+asso)
}
