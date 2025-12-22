package rawdata

const (
	SignatureMail = "Pour le centre d'inscriptions, <br /> Marie-Pierre BUFFET"
)

var (
	CoordonnesCentre = Centre{
		Nom:  "ACVE - Centre d'inscriptions",
		Tel:  "04 75 22 03 95",
		Mail: "inscriptions@acve.asso.fr",
	}

	Asso = InfosAsso{
		Title:   "<b>A</b>ssociation <b>C</b>hrétienne de <b>V</b>acances et de <b>L</b>oisirs",
		Infos:   "Siège social : <i>La Maison du Rocher - 26150 Chamaloc</i> - tél. <i>04 75 22 13 88</i> - <i>www.acve.asso.fr</i> - email: <i>contact@acve.asso.fr</i>",
		Details: "Association loi 1901 - N° Siret: 781 875 851 00037 - code APE: 552EB - Agréments :  Centre de Vacances 26 069 1003 - Jeunesse et Sport : 026ORG0163",
	}

	ExpediteurDefault = []string{CoordonnesCentre.Nom, CoordonnesCentre.Tel, CoordonnesCentre.Mail}
)

type Centre struct {
	Nom, Tel, Mail string
}

type InfosAsso struct {
	Title, Infos, Details string
}
