package directeurs

import (
	"errors"
	"slices"

	cps "registro/sql/camps"
	"registro/utils"

	"github.com/labstack/echo/v4"
)

type FormsOut struct {
	Forms   []cps.Form
	Answers map[cps.IdForm]cps.ParticipantForms
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
	answers, err := cps.SelectParticipantFormsByIdForms(ct.db, forms.IDs()...)
	if err != nil {
		return FormsOut{}, utils.SQLError(err)
	}

	list := utils.MapValues(forms)
	slices.SortFunc(list, func(a, b cps.Form) int { return int(a.Id - b.Id) })

	return FormsOut{Forms: list, Answers: answers.ByIdForm()}, nil
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

func (ct *Controller) updateForm(args cps.Form, user cps.IdCamp) error {
	form, err := cps.SelectForm(ct.db, args.Id)
	if err != nil {
		return utils.SQLError(err)
	}
	if form.IdCamp != user {
		return errors.New("internal error: wrong IdCamp")
	}
	// restrict if there already are answers
	answers, err := cps.SelectParticipantFormsByIdForms(ct.db, form.Id)
	if err != nil {
		return utils.SQLError(err)
	}
	if len(answers) != 0 {
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
