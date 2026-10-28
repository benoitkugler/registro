package compta

func (records ComptaBankentrys) AllValidated() bool {
	for _, rec := range records {
		if !rec.IsValidated {
			return false
		}
	}
	return true
}

func (ms Montants) equals(other Montants) bool {
	if len(ms) != len(other) {
		return false
	}
	for i := range ms {
		if ms[i] != other[i] {
			return false
		}
	}
	return true
}

func (ms Montants) Total() Cent {
	var total Cent
	for _, m := range ms {
		total += m.Cent
	}
	return total
}

// Uniquify returns a new slice, merging duplicate ventilations.
func (ms Montants) Uniquify() Montants {
	tmp := map[Ventilation]Cent{}
	for _, m := range ms {
		tmp[m.Kind] += m.Cent
	}
	out := make(Montants, 0, len(tmp))
	for k, m := range tmp {
		out = append(out, MontantPart{m, k})
	}
	return out
}
