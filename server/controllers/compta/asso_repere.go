package compta

import ca "registro/sql/compta"

// do not display a code
var ventilationCodeRepere = [ca.VentilationCount]string{}

func (ct *Controller) ventilationCodes() [ca.VentilationCount]string {
	if ct.asso.ID == "acve" {
		return ventilationCodeACVE
	}
	return ventilationCodeRepere
}
