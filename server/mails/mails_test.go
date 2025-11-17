package mails

import (
	"os"
	"testing"

	"registro/config"
	tu "registro/utils/testutils"
)

func loadEnv(t *testing.T) (config.Asso, config.SMTP) {
	tu.LoadEnv(t, "../env.sh")

	asso, err := config.NewAsso()
	tu.AssertNoErr(t, err)
	creds, err := config.NewSMTP(false)
	tu.AssertNoErr(t, err)
	return asso, creds
}

func TestMailer(t *testing.T) {
	html := `
	<!DOCTYPE html>
	<html xmlns="http://www.w3.org/1999/xhtml">

	<head>
		<meta name="viewport" content="width=device-width" />
		<meta charset="UTF-8">
		<meta http-uiv="Content-Type" content="text/html; charset=UTF-8" />
		<title>Courriel ACVE</title>
	</head>

	<body bgcolor="#c8db30">
		Salut !
	</body>
	</html>`

	cfg, creds := loadEnv(t)

	err := NewMailer(creds, cfg.MailsSettings).SendMail("", "Test", html, []string{""}, DefaultReplyTo)
	tu.AssertNoErr(t, err)
	err = NewMailer(creds, cfg.MailsSettings).SendMail("", "Test", html, nil, CustomReplyTo("test@registro.fr"))
	tu.AssertNoErr(t, err)
}

func TestPool(t *testing.T) {
	html := `
	<!DOCTYPE html>
	<html xmlns="http://www.w3.org/1999/xhtml">

	<head>
		<meta name="viewport" content="width=device-width" />
		<meta charset="UTF-8">
		<meta http-uiv="Content-Type" content="text/html; charset=UTF-8" />
		<title>Courriel ACVE</title>
		<style type="text/css">
			body {
				margin: 0 auto;
				padding: 0;
				min-width: 100%;
				font-family: sans-serif;
			}

			table {
				margin: 0 0 0 0;
			}

			ul {
				margin-top: 2px;
				margin-bottom: 2px;
			}
 
			.header {
				height: 40px;
				text-align: center;
				font-size: 16px;
				font-weight: bold;
			}

			.bonjour {
				padding-top: 20px;
				padding-left: 10px;
				padding-bottom: 5px;
				font-size: 16px;
			}

			.content {
				font-size: 16px;
				line-height: 30px;
			}

			.salutations {
				height: 70px;
				text-align: left;
				font-size: 16px;
			}

			.footer {
				text-align: left;
				height: 45px;
				font-size: 12px;
			}

			table.coordonnees {
				margin-left: 30px;
				color: #5b3c33;
			}
		</style>
	</head>

	<body bgcolor="#c8db30">
		<table bgcolor="#FFFFFF" width="100%" border="0" cellspacing="0" cellpadding="0">
			<tr class="header" bgcolor="#c8db30">
				<td style="text-align: center; padding: 5px;">
					ACVE - Inscription rapide
				</td>
			</tr>
			<tr>
				<td class="bonjour">

	Bonjour,

				</td>
			</tr>
			<tr class="content">
				<td style="padding:10px;">

	<p>
		Votre adresse mail <b>smsld@free.fr</b> est présente dans nos données. Vous pouvez remplir le formulaire d'inscription
		en choisissant le responsable légal parmi la liste ci-dessous.
		<ul>

			<li>
				<a href="http://free.fr">lkdkmslkd</a>
			</li>

			<li>
				<a href="">sdsd</a>
			</li>

			<li>
				<a href="">lkdkmssdsdlkd</a>
			</li>

		</ul>
	</p>

	<p>
		Les participants associés seront aussi importés.
	</p>

				</td>
			</tr>

			<tr class="salutations">
				<td style="padding: 10px;">
					<i>Ps : Ceci est un mail automatique, merci de ne pas y répondre.</i>
				</td>
			</tr>

			<tr class="footer" bgcolor="#b8dbf1">
				<td style="padding: 5px;">
					<b>A</b>ssociation <b>C</b>hrétienne de <b>V</b>acances et de <b>L</b>oisirs <br />
					Siège social : <i>La Maison du Rocher - 26150 Chamaloc</i> - tél. <i>04 75 22 13 88</i> - <i>www.acve.config.fr</i> - email: <i>contact@acve.config.fr</i>
				</td>
			</tr>
		</table>
	</body>

	</html>`

	content, err := os.ReadFile("test/sample.pdf")
	tu.AssertNoErr(t, err)
	pjs := []PieceJointe{{Content: content, Filename: "sample.pdf"}}

	cfg, creds := loadEnv(t)

	p, err := NewPool(creds, cfg.MailsSettings, pjs)
	tu.AssertNoErr(t, err)
	defer p.Close()

	for i := 0; i < 3; i++ {
		err = p.SendMail("", "[registro] test", html, nil, nil)
		tu.AssertNoErr(t, err)
	}
}
