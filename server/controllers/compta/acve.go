package compta

import ca "registro/sql/compta"

// custom logic for ACVE

type ResolvedBus struct {
	ca.Bus
	Kms       int
	Tarif     float32
	TotalCost float32
}

// type ResolvedPhotocopies struct {
// 	Start, End, Number int
// 	Tarif              Cent
// 	TotalCost          float32
// }

type ResolvedMateriel struct {
	Days                                       int
	NbPeople, NbInscrits                       int
	DaysPeopleEntretien, DaysPeoplePedagogique int
	TarifEntretien, TarifPedagogique           float32
	CostEntretien, CostPedagogique             float32
	Total                                      float32

	totalEntretien, totalPedagogique Cent
}

// BusExpense is a summary of a [ComptaRecord],
// keeping only the ventilation affected to cars
type BusExpense struct {
	Id    ca.IdComptaSubrecord
	Title string
	Total float32
}

type ResolvedFraisMDR struct {
	Bus            []ResolvedBus
	ExpensesForBus []BusExpense
	TotalBus       float32 // taking into account the potential expenses recorded
	totalBus       Cent

	// Photocopies ResolvedPhotocopies
	HasMateriel bool // true for camp at MDR
	Materiel    ResolvedMateriel

	Total float32
}

func (ld *comptaLoader) resolveFraismdr() (out ResolvedFraisMDR) {
	return resolve(ld.compta.Fraismdr, ld.config, ld.subrecords, ld.forBus, ld.camp.Duree, ld.inscrits(), len(ld.equipiers), ld.atMDR())
}

func resolve(frais ca.FraismdrACVE, config ca.ComptaConfig, subrecords ca.ComptaSubrecords, links ca.ComptaRecordBuss, days, inscrits, equipiers int, atMDR bool) (out ResolvedFraisMDR) {
	// 	out.Bus = make([]ResolvedBus, len(frais.Bus))
	// 	for i, bus := range frais.Bus {
	// 		tarif := config.CarsMDR[bus.Kind]
	// 		kms := bus.KmEnd - bus.KmStart
	// 		total := tarif * Cent(kms)
	// 		out.Bus[i] = ResolvedBus{
	// 			Bus:       bus,
	// 			Kms:       kms,
	// 			Tarif:     tarif.euros(),
	// 			TotalCost: total.euros(),
	// 		}
	// 		out.totalBus += total
	// 	}

	// 	// select the expenses devoted to bus
	// 	for _, link := range links {
	// 		rec := subrecords[link.IdSubrecord]
	// 		exp := rec.carExpenses()
	// 		out.ExpensesForBus = append(out.ExpensesForBus, BusExpense{rec.Id, rec.Title, exp.euros()})
	// 		out.totalBus -= exp
	// 	}

	// 	out.TotalBus = out.totalBus.euros()

	// 	// photocopNb := in.PhotocopiesEnd - in.PhotocopiesStart
	// 	// out.Photocopies = ResolvedPhotocopies{
	// 	// 	Start: in.PhotocopiesStart, End: in.PhotocopiesEnd,
	// 	// 	Number: photocopNb,
	// 	// }

	// 	all := inscrits + equipiers
	// 	costEntretien := Cent(days*all) * config.EntretienMDR
	// 	costMateriel := Cent(days*inscrits) * config.MaterielMDR
	// 	out.Materiel = ResolvedMateriel{
	// 		Days:                  days,
	// 		NbInscrits:            inscrits,
	// 		NbPeople:              all,
	// 		DaysPeopleEntretien:   days * all,
	// 		DaysPeoplePedagogique: days * inscrits,
	// 		TarifEntretien:        config.EntretienMDR.euros(),
	// 		TarifPedagogique:      config.MaterielMDR.euros(),
	// 		CostEntretien:         costEntretien.euros(),
	// 		CostPedagogique:       costMateriel.euros(),
	// 		totalEntretien:        costEntretien,
	// 		totalPedagogique:      costMateriel,
	// 		Total:                 (costEntretien + costMateriel).euros(),
	// 	}
	// 	out.HasMateriel = atMDR

	// 	total := out.totalBus
	// 	if out.HasMateriel {
	// 		total += costEntretien + costMateriel
	// 	}
	// 	out.Total = total.euros()

	return out
}
