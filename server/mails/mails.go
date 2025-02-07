package mails

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"mime"
	"net/smtp"
	"path/filepath"
	"strings"
	"time"

	"registro/config"

	"github.com/jordan-wright/email"
	"github.com/k3a/html2text"
)

var debugEmails = [...]string{
	"x.ben.x@free.fr", "bench26@gmail.com", "benoit-marc.kugler@ac-creteil.fr",
}

// Utilise l'adresse fournie par les crédences
var DefaultReplyTo ReplyTo = defaultReplyTo{}

type PieceJointe struct {
	Content  []byte
	Filename string
}

func (pj PieceJointe) attach(e *email.Email) error {
	ty := mime.TypeByExtension(filepath.Ext(pj.Filename))
	_, err := e.Attach(bytes.NewReader(pj.Content), pj.Filename, ty)
	return err
}

type ReplyTo interface {
	selectReplyTo(settings config.MailsSettings) string
}

type defaultReplyTo struct{}

func (d defaultReplyTo) selectReplyTo(settings config.MailsSettings) string { return settings.ReplyTo }

type CustomReplyTo string

func (d CustomReplyTo) selectReplyTo(config.MailsSettings) string { return string(d) }

func getFromAuth(creds config.SMTP) (string, smtp.Auth) {
	from := fmt.Sprintf("%s:%d", creds.Host, creds.Port)
	auth := smtp.PlainAuth("", creds.User, creds.Password, creds.Host)
	return from, auth
}

func getTo(prod bool, to string) string {
	if !prod { // sécurité
		oldTo := to
		to = debugEmails[rand.Intn(len(debugEmails))]
		log.Printf("Dev mode : changement de destinataire : %s ---> %s", oldTo, to)
	}
	return strings.TrimSpace(to)
}

func newMail(to []string, subject, htmlBody string, ccs []string, replyTo ReplyTo,
	pjs []PieceJointe, creds config.SMTP, settings config.MailsSettings,
) (*email.Email, error) {
	e := email.NewEmail()
	if replyTo != nil {
		e.ReplyTo = []string{replyTo.selectReplyTo(settings)}
	}

	e.To = make([]string, len(to))
	for index, m := range to {
		e.To[index] = getTo(creds.Prod, m)
		if e.To[index] == "" {
			return nil, errors.New("adresse mail non fournie")
		}
	}
	e.Cc = make([]string, len(ccs))
	for index, cc := range ccs {
		e.Cc[index] = getTo(creds.Prod, cc)
	}
	e.From = fmt.Sprintf("%s <%s>", settings.AssoName, creds.User)
	e.Subject = fmt.Sprintf("[%s] %s", settings.AssoName, subject)
	e.HTML = []byte(htmlBody)

	if settings.Sauvegarde != "" {
		e.Bcc = []string{settings.Sauvegarde}
	}
	if settings.Unsubscribe != "" {
		e.Headers.Set("List-Unsubscribe", fmt.Sprintf("<mailto:%s>", settings.Unsubscribe))
	}

	text := html2text.HTML2Text(htmlBody)
	e.Text = []byte(text)

	for _, pj := range pjs { // ajout des pièces jointes
		if err := pj.attach(e); err != nil {
			return nil, err
		}
	}
	return e, nil
}

type Mailer interface {
	// MailsSettings.AssoName is added as prefix to the subject
	SendMail(to, subject, htmlBody string, ccs []string, replyTo ReplyTo) (err error)
}

type basicMailer struct {
	smtp     config.SMTP
	settings config.MailsSettings
}

func NewMailer(smtp config.SMTP, settings config.MailsSettings) Mailer {
	return basicMailer{smtp, settings}
}

// SendMail envoie un mail directement.
func (b basicMailer) SendMail(to, subject, htmlBody string, ccs []string, replyTo ReplyTo) (err error) {
	e, err := newMail([]string{to}, subject, htmlBody, ccs, replyTo, nil, b.smtp, b.settings)
	if err != nil {
		return err
	}

	from, auth := getFromAuth(b.smtp)
	if err = e.Send(from, auth); err != nil {
		return fmt.Errorf("impossible d'envoyer le mail : %s", err)
	}
	return nil
}

// Pool optimise l'envoie de plusieurs mails consécutifs
type Pool struct {
	pool *email.Pool

	pjs      []PieceJointe
	creds    config.SMTP
	settings config.MailsSettings
}

// NewPool crée une interface pour envoyer plusieurs fois un mail avec
// les mêmes pièces jointes.
// Utiliser `SendMail` (plusieurs fois) puis `Close` (une fois)
func NewPool(credences config.SMTP, settings config.MailsSettings, pjs []PieceJointe) (Pool, error) {
	from, auth := getFromAuth(credences)
	p, err := email.NewPool(from, 1, auth)
	if err != nil {
		return Pool{}, err
	}
	return Pool{p, pjs, credences, settings}, err
}

func (p Pool) SendMail(to, subject, htmlBody string, ccs []string, replyTo ReplyTo) error {
	mail, err := newMail([]string{to}, subject, htmlBody, ccs, replyTo, p.pjs, p.creds, p.settings)
	if err != nil {
		return err
	}
	if err := p.pool.Send(mail, 10*time.Second); err != nil {
		return fmt.Errorf("impossible d'envoyer le mail : %s", err)
	}
	return nil
}

func (p Pool) Close() { p.pool.Close() }
