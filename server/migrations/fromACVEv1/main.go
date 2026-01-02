// Script de migration depuis la v1 ACVE.
//
// Ce script suppose que la base v2 a déjà été créée (sans données) et
//   - charge les données v1
//   - nettoie en conservant uniquement les années récentes
//   - convertit au format v2 et insère
package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"

	_ "github.com/lib/pq"

	"registro/config"
	rd "registro/migrations/fromACVEv1/rawdata"
	"registro/sql/dons"
	ds "registro/sql/dossiers"
	pr "registro/sql/personnes"
	"registro/sql/shared"
	"registro/utils"
)

// parameters from trimming

const (
	donFirstYear, donLastYear = 2025, 2026
)

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	dbV1, err := connectV1()
	check(err)
	defer dbV1.Close()

	dbV2, err := connectV2()
	check(err)
	defer dbV2.Close()

	fmt.Println("Connected.")

	fmt.Println("Loading v1 data...")
	originV1, err := loadV1(dbV1)
	check(err)
	fmt.Println("Done. Trimming...")
	trimmedV1 := originV1.trim()
	fmt.Println("Done. Inserting in v2...")
	err = insert(dbV2, originV1, trimmedV1)
	check(err)
	fmt.Println("Done.")
}

func euros(v rd.Euros) ds.Montant { return ds.NewEuros(float64(v)) }

func md(v rd.ModePaiment) ds.ModePaiement {
	switch v {
	default:
		panic("unsupported ModePaiement " + v)
	case rd.MPVirement:
		return ds.Virement
	case rd.MPCheque:
		return ds.Cheque
	case rd.MPEspece:
		return ds.Especes
	case rd.MPCarte:
		return ds.EnLigne
	case rd.MPAncv:
		return ds.Ancv
	case rd.MPHelloasso:
		return ds.Helloasso
	}
}

func sexe(v rd.Sexe) pr.Sexe {
	switch v {
	case rd.SHomme:
		return pr.Man
	case rd.SFemme:
		return pr.Woman
	default:
		return pr.NoSexe
	}
}

func diplome(v rd.Diplome) pr.Diplome {
	switch v {
	default:
		return pr.DAucun
	case rd.DBafa:
		return pr.DBafa
	case rd.DBafaStag:
		return pr.DBafaStag
	case rd.DBafd:
		return pr.DBafd
	case rd.DBafdStag:
		return pr.DBafdStag
	case rd.DCap:
		return pr.DCap
	case rd.DAssSociale:
		return pr.DAssSociale
	case rd.DEducSpe:
		return pr.DEducSpe
	case rd.DMonEduc:
		return pr.DMonEduc
	case rd.DInstit:
		return pr.DInstit
	case rd.DProf:
		return pr.DProf
	case rd.DAgreg:
		return pr.DAgreg
	case rd.DBjeps:
		return pr.DBjeps
	case rd.DDut:
		return pr.DDut
	case rd.DEje:
		return pr.DEje
	case rd.DDeug:
		return pr.DDeug
	case rd.DStaps:
		return pr.DStaps
	case rd.DBapaat:
		return pr.DBapaat
	case rd.DBeatep:
		return pr.DBeatep
	case rd.DZzautre:
		return pr.DZzautre
	}
}

func approfondissement(v rd.Approfondissement) pr.Approfondissement {
	switch v {
	default:
		return pr.AAucun
	case rd.AAutre:
		return pr.AAutre
	case rd.ASb:
		return pr.ASb
	case rd.ACanoe:
		return pr.ACanoe
	case rd.AVoile:
		return pr.AVoile
	case rd.AMoto:
		return pr.AMoto
	}
}

type baseV1 struct {
	personnes  rd.Personnes
	dons       rd.Dons
	donateurs  rd.DonDonateurs
	organismes rd.Organismes

	camps rd.Camps

	equipiers rd.Equipiers
}

func loadV1(db *sql.DB) (out baseV1, err error) {
	out.personnes, err = rd.SelectAllPersonnes(db)
	if err != nil {
		return baseV1{}, err
	}
	out.dons, err = rd.SelectAllDons(db)
	if err != nil {
		return baseV1{}, err
	}
	out.donateurs, err = rd.SelectAllDonDonateurs(db)
	if err != nil {
		return baseV1{}, err
	}
	out.organismes, err = rd.SelectAllOrganismes(db)
	if err != nil {
		return baseV1{}, err
	}
	out.camps, err = rd.SelectAllCamps(db)
	if err != nil {
		return baseV1{}, err
	}
	out.equipiers, err = rd.SelectAllEquipiers(db)
	if err != nil {
		return baseV1{}, err
	}

	return
}

// returns a trimmed, deep copy
func (b baseV1) trim() (out baseV1) {
	out.dons, out.donateurs = trimDons(b.dons, b.donateurs)
	out.organismes = trimOrganismes(out.dons, out.donateurs, b.organismes)
	out.personnes = trimPersonnes(b.personnes, out)

	return out
}

// only keep (current) and last year, also drop anonymous gifts
// properly trim link table
func trimDons(dons rd.Dons, donateurs rd.DonDonateurs) (rd.Dons, rd.DonDonateurs) {
	byDon := donateurs.ByIdDon()

	outDons := make(rd.Dons)
	for id, don := range dons {
		if year := don.DateReception.Time().Year(); !(donFirstYear <= year && year <= donLastYear) {
			continue
		} else if _, hasDonateur := byDon[id]; !hasDonateur {
			continue
		}
		outDons[id] = don
	}

	var outLinks rd.DonDonateurs
	for _, link := range donateurs {
		if _, has := outDons[link.IdDon]; !has {
			continue
		}
		outLinks = append(outLinks, link)
	}

	return outDons, outLinks
}

// only keep organismes required in [dons]
func trimOrganismes(dons rd.Dons, donateurs rd.DonDonateurs, organismes rd.Organismes) rd.Organismes {
	byDon := donateurs.ByIdDon()
	usedOrganismes := rd.NewSet()
	for idDon := range dons {
		donateur := byDon[idDon]
		if donateur.IdOrganisme.Valid {
			usedOrganismes.Add(donateur.IdOrganisme.Int64)
		}
	}
	out := make(rd.Organismes)
	for id, v := range organismes {
		if !usedOrganismes.Has(id) {
			continue
		}
		out[id] = v
	}
	return out
}

// only keep personnes used in other tables
func trimPersonnes(v1 rd.Personnes, trimmed baseV1) rd.Personnes {
	used := rd.NewSet()
	for _, link := range trimmed.donateurs {
		if link.IdPersonne.Valid {
			used.Add(link.IdPersonne.Int64)
		}
	}
	// TODO

	out := make(rd.Personnes)
	for id := range used {
		p := v1[id]
		if p.IsTemporaire {
			panic("temporary profile with ID: " + strconv.Itoa(int(id)))
		}
		out[id] = p
	}
	return out
}

// use the last equipier; returns a map keyed by IdPersonne
func (b baseV1) inferDiplomeAppro() map[int64]rd.Equipier {
	lastEquipier := map[int64]rd.Equipier{}
	for _, equipier := range b.equipiers {
		date := b.camps[equipier.IdCamp].DateDebut.Time()
		idPersonne := equipier.IdPersonne
		last, ok := lastEquipier[idPersonne]
		lastDate := b.camps[last.IdCamp].DateDebut.Time()
		if !ok || lastDate.Before(date) {
			lastEquipier[idPersonne] = equipier
		}
	}
	return lastEquipier
}

type (
	organismesM map[int64]dons.IdOrganisme
	personnesM  map[int64]pr.IdPersonne
)

func insert(dbV2 *sql.DB, origin, trimmed baseV1) error {
	diplomeAppros := origin.inferDiplomeAppro()

	pM := make(personnesM)
	oM := make(organismesM)
	return utils.InTx(dbV2, func(tx *sql.Tx) error {
		var err error

		// insert personnes
		fmt.Println("\t", len(trimmed.personnes), "personnes")
		for _, v1 := range trimmed.personnes {
			equipier := diplomeAppros[v1.Id]
			v2 := pr.Personne{
				Etatcivil: pr.Etatcivil{
					Nom:                  string(v1.Nom),
					Prenom:               string(v1.Prenom),
					Sexe:                 sexe(v1.Sexe),
					DateNaissance:        shared.NewDateFrom(v1.DateNaissance.Time()),
					VilleNaissance:       string(v1.VilleNaissance),
					DepartementNaissance: pr.Departement(v1.DepartementNaissance),
					Nationnalite:         pr.Nationnalite{},
					Tels:                 pr.Tels(v1.Tels),
					Mail:                 string(v1.Mail),
					Adresse:              string(v1.Adresse),
					CodePostal:           string(v1.CodePostal),
					Ville:                string(v1.Ville),
					Pays:                 pr.Pays(v1.Pays),
					NomJeuneFille:        string(v1.NomJeuneFille),
					Profession:           string(v1.Profession),
					Etudiant:             bool(v1.Etudiant),
					Fonctionnaire:        bool(v1.Fonctionnaire),
					Diplome:              diplome(equipier.Diplome),
					Approfondissement:    approfondissement(equipier.Appro),
				},
			}
			v2, err = v2.Insert(tx)
			if err != nil {
				return err
			}
			pM[v1.Id] = v2.Id
		}

		// insert organismes
		fmt.Println("\t", len(trimmed.organismes), "organismes")
		for _, v1 := range trimmed.organismes {
			var coord rd.Coordonnees
			if contact := v1.IdContactDon; contact.Valid {
				p := origin.personnes[contact.Int64]
				coord = p.Coordonnees()
			} else if v1.ContactPropre {
				coord = v1.Contact
			} else if contact := v1.IdContact; contact.Valid {
				p := origin.personnes[contact.Int64]
				coord = p.Coordonnees()
			}
			v2 := dons.Organisme{
				Nom:        string(v1.Nom),
				Mail:       string(coord.Mail),
				Adresse:    string(coord.Adresse),
				CodePostal: string(coord.CodePostal),
				Ville:      string(coord.Ville),
				Pays:       pr.Pays(coord.Pays),
			}
			v2, err = v2.Insert(tx)
			if err != nil {
				return err
			}
			oM[v1.Id] = v2.Id
		}

		// insert dons
		donateursByDon := origin.donateurs.ByIdDon()
		fmt.Println("\t", len(trimmed.dons), "dons")
		for _, v1 := range trimmed.dons {
			details := string(v1.Details)
			if v1.ModePaiement == rd.MPHelloasso {
				details += v1.Infos.IdPaiementHelloAsso
			}
			v2 := dons.Don{
				Montant:      euros(v1.Valeur),
				ModePaiement: md(v1.ModePaiement),
				Date:         shared.NewDateFrom(v1.DateReception.Time()),
				Affectation:  string(v1.Affectation),
				Details:      details,
				Remercie:     bool(v1.Remercie),
			}
			donateur := donateursByDon[v1.Id]
			if idV1 := donateur.IdPersonne; idV1.Valid {
				v2.IdPersonne = pM[idV1.Int64].Opt()
			} else if idV1 := donateur.IdOrganisme; idV1.Valid {
				v2.IdOrganisme = oM[idV1.Int64].Opt()
			}
			v2, err = v2.Insert(tx)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

func connectV1() (*sql.DB, error) {
	v1_HOST := os.Getenv("v1_HOST")
	v1_USER := os.Getenv("v1_USER")
	v1_PASSWORD := os.Getenv("v1_PASSWORD")
	v1_NAME := os.Getenv("v1_NAME")
	port := 5432
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s",
		v1_HOST, port, v1_USER, v1_PASSWORD, v1_NAME)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("connexion DB : %s", err)
	}
	return db, nil
}

func connectV2() (*sql.DB, error) {
	config, err := config.NewDB()
	if err != nil {
		return nil, err
	}
	return config.ConnectPostgres()
}
