package backoffice

import (
	"registro/sql/events"

	"github.com/labstack/echo/v4"
)

type QueryAttente uint8

const (
	EmptyQA         QueryAttente = iota // Indifférent
	AvecAttente                         // Avec liste d'attente
	AvecAttenteOnly                     // Seulement avec liste d'attente
	AvecInscrits                        // Avec inscrits
)

type QueryReglement uint8

const (
	EmptyQR QueryReglement = iota // Indifférent
	Zero                          // Non commencé
	Partiel                       // En cours
	Total                         // Complété
)

// The zero value defaults to returning everything
type DossierQuery struct {
	Pattern   string // Responsable et participants
	IdCamp    events.OptIdCamp
	Attente   QueryAttente
	Reglement QueryReglement
}

// DossiersSearch returns a list of [Dossier] headers
// matching the given query, sorted by activity time (defined by the messages)
func (ct *Controller) DossiersSearch(c echo.Context) error {
	var out int
	return c.JSON(200, out)
}
