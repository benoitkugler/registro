package compta

import "slices"

func (records ComptaRecords) AllValidated() bool {
	for _, rec := range records {
		if !rec.IsValidated {
			return false
		}
	}
	return true
}

func (ms Montants) Equals(other Montants) bool {
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

func (m ComptaSubrecords) Sorted() []ComptaSubrecord {
	out := make([]ComptaSubrecord, 0, len(m))
	for _, v := range m {
		out = append(out, v)
	}
	slices.SortFunc(out, func(a, b ComptaSubrecord) int { return int(a.Id - b.Id) })
	return out
}

// Sorted returns the more recent configs first
func (configs ComptaConfigs) Sorted() []ComptaConfig {
	var sorted []ComptaConfig
	for _, config := range configs {
		sorted = append(sorted, config)
	}
	slices.SortFunc(sorted, func(i, j ComptaConfig) int { return j.Year - i.Year })
	return sorted
}
