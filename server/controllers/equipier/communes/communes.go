package communes

import (
	"strconv"
	"strings"

	"fmt"
)

func IsCorse(departement string) (is2A, is2B bool) {
	switch departement {
	case departementsFrancais[100]:
		return true, false
	case departementsFrancais[101]:
		return false, true
	}
	return false, false
}

// CommuneByCode returns the departement and commune defined by the given
// INSEE 5 letters code
func CommuneByCode(insee string) (string, string, error) {
	departementNb, communeNb, err := ParseCommuneCode(insee)
	if err != nil {
		return "", "", err
	}
	departement := departementsFrancais[departementNb] // sanitized in ParseCommuneCode
	// properly handle outre mer
	if departementNb == 97 {
		nextDigit := communeNb / 100
		if nextDigit <= 6 {
			departement = departementsFrancais[110+nextDigit]
		}
	}

	departementNb -= 1
	communeNb -= 1
	l := communesFrancaises[departementNb] // sanitized in ParseCommuneCode
	if communeNb >= len(l) {
		return departement, "", nil
	}
	return departement, l[communeNb], nil
}

// ParseCommuneCode parse the INSEE commune code, handling special cases
// such as Corse.
func ParseCommuneCode(communeCode string) (departement, commune int, _ error) {
	// handle Corse
	if strings.HasPrefix(communeCode, "2A") {
		communeCode = "100" + communeCode[2:]
	} else if strings.HasPrefix(communeCode, "2B") {
		communeCode = "101" + communeCode[2:]
	}

	code, err := strconv.Atoi(communeCode)
	if err != nil {
		return 0, 0, err
	}
	// <DD><CCC>
	departement = code / 1000
	commune = code % 1000

	if !(1 <= departement && departement <= 101) {
		return 0, 0, fmt.Errorf("unexpected departement in %s", communeCode)
	}
	if !(1 <= commune && commune <= 999) {
		return 0, 0, fmt.Errorf("unexpected commune in %s", communeCode)
	}

	return departement, commune, nil
}

// 1-based
var departementsFrancais = [...]string{
	1:  "Ain",
	2:  "Aisne",
	3:  "Allier",
	4:  "Alpes-de-Haute-Provence",
	5:  "Hautes-Alpes",
	6:  "Alpes-Maritimes",
	7:  "Ardèche",
	8:  "Ardennes",
	9:  "Ariège",
	10: "Aube",
	11: "Aude",
	12: "Aveyron",
	13: "Bouches-du-Rhône",
	14: "Calvados",
	15: "Cantal",
	16: "Charente",
	17: "Charente-Maritime",
	18: "Cher",
	19: "Corrèze",
	21: "Côte-d'Or",
	22: "Côtes-d'Armor",
	23: "Creuse",
	24: "Dordogne",
	25: "Doubs",
	26: "Drôme",
	27: "Eure",
	28: "Eure-et-Loir",
	29: "Finistère",
	30: "Gard",
	31: "Haute-Garonne",
	32: "Gers",
	33: "Gironde",
	34: "Hérault",
	35: "Ille-et-Vilaine",
	36: "Indre",
	37: "Indre-et-Loire",
	38: "Isère",
	39: "Jura",
	40: "Landes",
	41: "Loir-et-Cher",
	42: "Loire",
	43: "Haute-Loire",
	44: "Loire-Atlantique",
	45: "Loiret",
	46: "Lot",
	47: "Lot-et-Garonne",
	48: "Lozère",
	49: "Maine-et-Loire",
	50: "Manche",
	51: "Marne",
	52: "Haute-Marne",
	53: "Mayenne",
	54: "Meurthe-et-Moselle",
	55: "Meuse",
	56: "Morbihan",
	57: "Moselle",
	58: "Nièvre",
	59: "Nord",
	60: "Oise",
	61: "Orne",
	62: "Pas-de-Calais",
	63: "Puy-de-Dôme",
	64: "Pyrénées-Atlantiques",
	65: "Hautes-Pyrénées",
	66: "Pyrénées-Orientales",
	67: "Bas-Rhin",
	68: "Haut-Rhin",
	69: "Rhône",
	70: "Haute-Saône",
	71: "Saône-et-Loire",
	72: "Sarthe",
	73: "Savoie",
	74: "Haute-Savoie",
	75: "Paris",
	76: "Seine-Maritime",
	77: "Seine-et-Marne",
	78: "Yvelines",
	79: "Deux-Sèvres",
	80: "Somme",
	81: "Tarn",
	82: "Tarn-et-Garonne",
	83: "Var",
	84: "Vaucluse",
	85: "Vendée",
	86: "Vienne",
	87: "Haute-Vienne",
	88: "Vosges",
	89: "Yonne",
	90: "Territoire de Belfort",
	91: "Essonne",
	92: "Hauts-de-Seine",
	93: "Seine-Saint-Denis",
	94: "Val-de-Marne",
	95: "Val-d'Oise",

	99: "Hors France",

	100: "Corse-du-Sud", // internal convention for Corse
	101: "Haute-Corse",  // internal convention for Corse

	111: "Guadeloupe", // internal convention for OutreMer
	112: "Martinique", // internal convention for OutreMer
	113: "Guyane",     // internal convention for OutreMer
	114: "La Réunion", // internal convention for OutreMer
	116: "Mayotte",    // internal convention for OutreMer
}
