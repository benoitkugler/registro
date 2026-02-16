package directeurs

import (
	"database/sql"
	"fmt"
	"log"
	"mime"
	"net/url"
	"path/filepath"
	"strings"

	filesAPI "registro/controllers/files"
	"registro/crypto"
	"registro/generators/pdfcreator"
	"registro/logic"
	cps "registro/sql/camps"
	fs "registro/sql/files"
	pr "registro/sql/personnes"
	"registro/utils"

	"github.com/labstack/echo/v4"
	"github.com/microcosm-cc/bluemonday"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

const (
	EndpointLettreImages = "/service/lettre-images"
	queryParamName       = "key"
)

var (
	htmlPolicy *bluemonday.Policy

	tags = []string{
		"sub", "sup", "b", "i", "u", "s", "h1", "h2", "h3", "h4", "h5", "h6",
		"strike", "a", "strong", "ul", "ol", "li", "br",
		"span", "em", "p", "blockquote", "hr", "img",
	}

	attributes = []string{"style", "title", "src", "width", "height", "href", "target"}

	styles = []string{
		"color", "background-color", "font-weight", "text-align", "font-size", "float", "margin",
		"text-decoration", "margin-left", "margin-right", "display", "border-style", "border-width",
	}

	fonts = []string{"arial"}
)

// mise en place des régles de filtrages
func init() {
	htmlPolicy = bluemonday.NewPolicy()
	htmlPolicy.AllowElements(tags...)
	htmlPolicy.AllowAttrs(attributes...).Globally()
	htmlPolicy.AllowStyles(styles...).Globally()
	// URLs must be parseable by net/url.Parse()
	htmlPolicy.RequireParseableURLs(true)
	// Most common URL schemes only
	htmlPolicy.AllowURLSchemes("mailto", "http", "https")
	htmlPolicy.AllowDataURIImages()
	htmlPolicy.AllowStyles("font-family").MatchingEnum(fonts...).Globally()
}

// Location of the image as expected by Tinymce
type LettreImageUploadOut struct {
	Location string `json:"location"`
}

// LettreImageUpload enregistre une image de la lettre au directeur
// et renvoie un chemin d'accès (au format attendu par Tinymce)
func (ct Controller) LettreImageUpload(c echo.Context) error {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return err
	}
	content, filname, err := filesAPI.ReadUpload(fileHeader)
	if err != nil {
		return err
	}
	location, err := ct.uploadLettreImage(c.Request().Host, content, filname)
	if err != nil {
		return err
	}
	out := LettreImageUploadOut{Location: location}
	return c.JSON(200, out)
}

// return an access URL
func (ct Controller) uploadLettreImage(host string, content []byte, filename string) (string, error) {
	out, err := cps.LettreImage{Filename: filename, Content: content}.Insert(ct.db)
	if err != nil {
		return "", utils.SQLError(err)
	}

	key := crypto.EncryptID(ct.key, out.Id)
	url := utils.BuildUrl(host, EndpointLettreImages, utils.QP(queryParamName, key))

	return url, err
}

// LettreImageGet renvoie l'image demandée, identifiée par un lien crypté.
func (ct Controller) LettreImageGet(c echo.Context) error {
	key := c.QueryParam(queryParamName)
	id, err := crypto.DecryptID[cps.IdLettreImage](ct.key, key)
	if err != nil {
		return err
	}
	image, err := cps.SelectLettreImage(ct.db, id)
	if err != nil {
		return utils.SQLError(err)
	}
	contentType := mime.TypeByExtension(filepath.Ext(image.Filename))
	return c.Blob(200, contentType, image.Content)
}

type LettreOut struct {
	Lettre cps.Lettredirecteur
	File   logic.PublicFile
}

func (ct *Controller) LettreGet(c echo.Context) error {
	user := JWTUser(c)
	out, err := ct.getLettre(user)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

func (ct *Controller) getLettre(user cps.IdCamp) (LettreOut, error) {
	lettre, _, err := cps.SelectLettredirecteurByIdCamp(ct.db, user)
	if err != nil {
		return LettreOut{}, utils.SQLError(err)
	}
	file, _, err := findLettre(ct.db, user)
	if err != nil {
		return LettreOut{}, utils.SQLError(err)
	}
	return LettreOut{lettre, logic.NewPublicFile(ct.key, file)}, nil
}

// LettreUpdate save the data and generate a PDF file from it.
func (ct *Controller) LettreUpdate(c echo.Context) error {
	user := JWTUser(c)

	var args cps.Lettredirecteur
	if err := c.Bind(&args); err != nil {
		return err
	}
	out, err := ct.updateLettreDirecteur(user, args)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

// select the [File] with IsLettre, or return false
// do not wraps errors
func findLettre(db cps.DB, idCamp cps.IdCamp) (fs.File, bool, error) {
	campFiles, err := fs.SelectFileCampsByIdCamps(db, idCamp)
	if err != nil {
		return fs.File{}, false, err
	}
	for _, link := range campFiles {
		if link.IsLettre {
			out, err := fs.SelectFile(db, link.IdFile)
			if err != nil {
				return fs.File{}, false, err
			}
			return out, true, nil
		}
	}
	return fs.File{}, false, nil
}

// updateLettreDirecteur enregistre le html et génère et enregistre le document associé.
// S'il existe, le document est remplacé.
//
// La mise à jour de la lettre se fait en 6 étapes :
//  0. mise à jour des paramètres (html + options)
//  1. la génération du PDF, via go-weasyprint
//  2. la lecture ou la création des méta-données du document final
//  3. l'enregistrement du HTML et l'enregistrement du PDF comme contenu du document
//  4. le partage du document final
//  5. le nettoyage (en arrière plan) des images non utilisées
func (ct *Controller) updateLettreDirecteur(idCamp cps.IdCamp, lettre cps.Lettredirecteur) (LettreOut, error) {
	// Etape 0: on prépare les paramètres de la lettre
	camp, err := cps.SelectCamp(ct.db, idCamp)
	if err != nil {
		return LettreOut{}, utils.SQLError(err)
	}

	lettre.IdCamp = idCamp
	lettre.Html = htmlPolicy.Sanitize(lettre.Html) // sécurité

	var directeur pr.Personne
	if !lettre.UseCoordCentre { // select Directeur
		var hasDirecteur bool
		directeur, hasDirecteur, err = ct.findDirecteur(idCamp)
		if err != nil {
			return LettreOut{}, err
		}
		if !hasDirecteur {
			return LettreOut{}, errNoDir
		}
	}

	// Etape 1: HTML -> PDF
	pdfContent, err := pdfcreator.CreateLettreDirecteur(ct.asso, lettre, directeur.Identite)
	if err != nil {
		return LettreOut{}, err
	}

	var file logic.PublicFile
	err = utils.InTx(ct.db, func(tx *sql.Tx) error {
		// Etape 2
		fileLettre, hasFile, err := findLettre(tx, idCamp)
		if err != nil {
			return err
		}
		if !hasFile { // pas encore de doc associé, on crée
			fileLettre, err = fs.File{}.Insert(tx)
			if err != nil {
				return err
			}
			err = fs.FileCamp{IdFile: fileLettre.Id, IdCamp: idCamp, IsLettre: true}.Insert(tx)
			if err != nil {
				return err
			}
		}

		// Etape 3.a
		// on efface un doublon potentiel ...
		err = lettre.Delete(tx)
		if err != nil {
			return err
		}
		// ... et on insère
		err = lettre.Insert(tx)
		if err != nil {
			return err
		}

		// Etape 3.b
		pdfName := fmt.Sprintf("Lettre du directeur %s.pdf", camp.Label())
		fileLettre, err = fs.UploadFile(ct.files, tx, fileLettre.Id, pdfContent, pdfName)
		if err != nil {
			return err
		}

		// Etape 4
		file = logic.NewPublicFile(ct.key, fileLettre)

		return nil
	})
	if err != nil {
		return LettreOut{}, err
	}

	// Etape 5 : on peut lancer le nettoyage en tache de fond
	go ct.garbageCollectImages()

	return LettreOut{lettre, file}, err
}

// return the key=<...> query parameters of each img tag
func extractImagesKeys(root *html.Node) []string {
	var (
		f   func(n *html.Node)
		out []string
	)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.DataAtom == atom.Img {
			for _, attr := range n.Attr {
				if attr.Key == "src" { // found an image
					parsed, err := url.Parse(attr.Val)
					if err != nil {
						log.Println("error parsing image url", err)
						continue
					}
					key := parsed.Query().Get(queryParamName)
					out = append(out, key)
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(root)
	return out
}

func (ct *Controller) garbageCollectImages() error {
	// comme ume même image peut être partagée dans plusieurs lettres
	// on doit considérer tous les camps en même temps
	lettres, err := cps.SelectAllLettredirecteurs(ct.db)
	if err != nil {
		return utils.SQLError(err)
	}
	used := utils.NewSet[cps.IdLettreImage]()
	for _, lettre := range lettres {
		root, err := html.Parse(strings.NewReader(lettre.Html))
		if err != nil {
			return fmt.Errorf("error extracting images: %s", err)
		}
		keys := extractImagesKeys(root)
		for _, key := range keys {
			id, err := crypto.DecryptID[cps.IdLettreImage](ct.key, key)
			if err != nil {
				return err
			}
			used.Add(id)
		}
	}
	err = cps.DeleteLettreImagesOthers(ct.db, used.Keys())
	if err != nil {
		return fmt.Errorf("error removing unsued images: %s", err)
	}
	return nil
}
