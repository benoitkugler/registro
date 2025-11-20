package services

import (
	"database/sql"
	"fmt"
	"html/template"

	"registro/config"
	"registro/crypto"
	"registro/logic"
	"registro/mails"

	"github.com/labstack/echo/v4"
)

type Controller struct {
	db   *sql.DB
	key  crypto.Encrypter
	smtp config.SMTP
	asso config.Asso
}

func NewController(db *sql.DB, key crypto.Encrypter, smtp config.SMTP, asso config.Asso) *Controller {
	return &Controller{db, key, smtp, asso}
}

type SearchMailOut struct {
	Found int
}

func (ct *Controller) SearchMail(c echo.Context) error {
	mail := c.QueryParam("mail")

	out, err := ct.searchMailAndSend(c.Request().Host, mail)
	if err != nil {
		return err
	}

	return c.JSON(200, out)
}

func (ct *Controller) searchMailAndSend(host string, mail string) (SearchMailOut, error) {
	dossiers, _, err := logic.LoadByMail(ct.db, mail)
	if err != nil {
		return SearchMailOut{}, err
	}

	var out []mails.ResumeDossier
	for _, dossier := range dossiers.Dossiers {
		loader := dossiers.For(dossier.Id)
		url := logic.EspacePersoURL(ct.key, host, dossier.Id)
		out = append(out, mails.ResumeDossier{
			Responsable: loader.Responsable().NOMPrenom(),
			CampsMap:    loader.Camps(),
			URL:         template.HTML(url),
		})
	}

	if len(out) != 0 {
		body, err := mails.RenvoieEspacePersoURL(ct.asso, mail, out)
		if err != nil {
			return SearchMailOut{}, fmt.Errorf("Votre adresse est présente, mais l'envoi du lien par mail a échoué (%s).", err)
		}
		err = mails.NewMailer(ct.smtp, ct.asso.MailsSettings).SendMail(mail, "Espace de suivi", body, nil, nil)
		if err != nil {
			return SearchMailOut{}, fmt.Errorf("Votre adresse est présente, mais l'envoi du lien par mail a échoué (%s).", err)
		}
	}

	return SearchMailOut{len(out)}, nil
}
