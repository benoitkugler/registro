package directeurs

import (
	"errors"
	"slices"
	"strings"

	cps "registro/sql/camps"
	"registro/utils"

	"github.com/labstack/echo/v4"
)

type FormsOut struct {
	Forms        []cps.Form
	AnswersCount map[cps.IdForm]int
}

// FormsLoad returns the forms and answers for the given camp.
func (ct *Controller) FormsLoad(c echo.Context) error {
	user := JWTUser(c)
	out, err := ct.loadForms(user)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

func (ct *Controller) loadForms(user cps.IdCamp) (FormsOut, error) {
	forms, err := cps.SelectFormsByIdCamps(ct.db, user)
	if err != nil {
		return FormsOut{}, utils.SQLError(err)
	}
	list := utils.MapValues(forms)
	slices.SortFunc(list, func(a, b cps.Form) int { return int(a.Id - b.Id) })

	answers, err := cps.SelectParticipantFormsByIdForms(ct.db, forms.IDs()...)
	if err != nil {
		return FormsOut{}, utils.SQLError(err)
	}
	byForm := answers.ByIdForm()
	counts := map[cps.IdForm]int{}
	for id, reponses := range byForm {
		counts[id] = len(reponses)
	}

	return FormsOut{Forms: list, AnswersCount: counts}, nil
}

func (ct *Controller) FormsLoadReponses(c echo.Context) error {
	user := JWTUser(c)
	idForm, err := utils.QueryParamInt[cps.IdForm](c, "id")
	if err != nil {
		return err
	}
	out, err := ct.loadFormReponses(idForm, user)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

type FormReponsesInscrit struct {
	IdParticipant cps.IdParticipant
	Inscrit       string
	Reponses      []string // same length as [FormReponses].Champs
}
type FormReponses struct {
	Inscrits []FormReponsesInscrit
	Champs   []string
}

func (ct *Controller) loadFormReponses(idForm cps.IdForm, user cps.IdCamp) (FormReponses, error) {
	form, err := cps.SelectForm(ct.db, idForm)
	if err != nil {
		return FormReponses{}, utils.SQLError(err)
	}
	if form.IdCamp != user {
		return FormReponses{}, errors.New("internal error: access fordidden")
	}

	camp, err := cps.LoadCamp(ct.db, user)
	if err != nil {
		return FormReponses{}, err
	}
	tmp, err := cps.SelectParticipantFormsByIdForms(ct.db, idForm)
	if err != nil {
		return FormReponses{}, utils.SQLError(err)
	}
	byParticipant := tmp.ByIdParticipant()

	champs := make([]string, len(form.Champs))
	for i, c := range form.Champs {
		champs[i] = c.Titre
	}

	pps := camp.Participants(true)
	inscrits := make([]FormReponsesInscrit, len(pps))
	for i, pp := range pps {
		reponses, _ := byParticipant[pp.Participant.Id].NonEmpty()
		inscrits[i] = FormReponsesInscrit{
			IdParticipant: pp.Participant.Id,
			Inscrit:       pp.Personne.PrenomNOM(),
			Reponses:      reponses.ToString(form.Champs),
		}
	}

	slices.SortFunc(inscrits, func(a, b FormReponsesInscrit) int { return strings.Compare(a.Inscrit, b.Inscrit) })

	return FormReponses{inscrits, champs}, nil
}

// FormsCreate create a new, empty form, for the given camp
func (ct *Controller) FormsCreate(c echo.Context) error {
	user := JWTUser(c)
	out, err := ct.createForm(user)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

func (ct *Controller) createForm(user cps.IdCamp) (cps.Form, error) {
	out, err := cps.Form{IdCamp: user}.Insert(ct.db)
	if err != nil {
		return out, utils.SQLError(err)
	}
	return out, nil
}

func (ct *Controller) FormsUpdate(c echo.Context) error {
	user := JWTUser(c)
	var args cps.Form
	if err := c.Bind(&args); err != nil {
		return err
	}
	err := ct.updateForm(args, user)
	if err != nil {
		return err
	}
	return c.NoContent(200)
}

func hasSameShape(f1, f2 cps.Form) bool {
	if len(f1.Champs) != len(f2.Champs) {
		return false
	}
	for i := range f1.Champs {
		c1, c2 := f1.Champs[i].Question, f2.Champs[i].Question
		switch c1 := c1.(type) {
		case cps.ChampTexte:
			_, ok := c2.(cps.ChampTexte)
			if !ok {
				return false
			}
		case cps.ChampQCM:
			c2, ok := c2.(cps.ChampQCM)
			if !ok {
				return false
			}
			if len(c1.Propositions) != len(c2.Propositions) {
				return false
			}
		}
	}
	return true
}

func (ct *Controller) updateForm(args cps.Form, user cps.IdCamp) error {
	form, err := cps.SelectForm(ct.db, args.Id)
	if err != nil {
		return utils.SQLError(err)
	}
	if form.IdCamp != user {
		return errors.New("internal error: wrong IdCamp")
	}
	// restrict if there already are answers :
	// we allow updating title and description of every field,
	// but not the question and length
	answers, err := cps.SelectParticipantFormsByIdForms(ct.db, form.Id)
	if err != nil {
		return utils.SQLError(err)
	}
	if len(answers) != 0 && !hasSameShape(form, args) {
		return errors.New("Ce formulaire a déjà des réponses.")
	}
	form.Nom = args.Nom
	form.Introduction = args.Introduction
	form.Champs = args.Champs
	_, err = form.Update(ct.db)
	if err != nil {
		return utils.SQLError(err)
	}
	return nil
}

func (ct *Controller) FormsDelete(c echo.Context) error {
	user := JWTUser(c)
	id, err := utils.QueryParamInt[cps.IdForm](c, "id")
	if err != nil {
		return err
	}
	err = ct.deleteForm(id, user)
	if err != nil {
		return err
	}
	return c.NoContent(200)
}

func (ct *Controller) deleteForm(id cps.IdForm, user cps.IdCamp) error {
	form, err := cps.SelectForm(ct.db, id)
	if err != nil {
		return utils.SQLError(err)
	}
	if form.IdCamp != user {
		return errors.New("internal error: wrong IdCamp")
	}
	_, err = cps.DeleteFormById(ct.db, id) // answsers delete by cascade
	if err != nil {
		return utils.SQLError(err)
	}
	return nil
}
