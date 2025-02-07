package personnes

import "time"

func (r Etatcivil) Age() int { return r.DateNaissance.Age(time.Now()) }
