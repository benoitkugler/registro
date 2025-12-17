import { Formatters } from "@/utils";
import {
  StatutParticipant,
  type Date_,
  type Groupe,
  type Groupes,
  type IdGroupe,
  type ParticipantExt,
} from "../../logic/api";
import { addDays } from "@/components/date";
import type { Int } from "@/urls";

function sorted(groupes: Groupes) {
  const l = Object.values(groupes || {});
  l.sort((a, b) => new Date(a.Fin).valueOf() - new Date(b.Fin).valueOf());
  return l;
}
function trouveGroupe(sortedGroupes: Groupe[], dateNaissance: Date_) {
  // see server code : Groupes.TrouveGroupe

  const d = new Date(dateNaissance).valueOf();

  for (const groupe of sortedGroupes) {
    const fin = new Date(groupe.Fin).valueOf();
    if (d <= fin) {
      return groupe.Id;
    }
  }

  return null;
}

export function groupesPlages(groupes: Groupes) {
  const out: Record<IdGroupe, string> = {};

  const l = sorted(groupes);
  if (l.length == 0) return out;

  const plages: string[] = Array.from({ length: l.length });

  // le premier groupe commence à -infty
  plages[0] = `né avant le ${Formatters.dateNaissance(l[0].Fin)}`;

  for (let index = 1; index < l.length; index++) {
    const element = l[index];
    const previous = l[index - 1];
    const start = addDays(previous.Fin, 1 as Int);
    plages[index] = `né entre le ${Formatters.dateNaissance(
      start
    )} et le ${Formatters.dateNaissance(element.Fin)}`;
  }

  l.forEach((groupe, index) => (out[groupe.Id] = plages[index]));
  return out;
}

export function groupesSizes(groupes: Groupes, participants: ParticipantExt[]) {
  const l = sorted(groupes);
  const sizes: Record<IdGroupe, number> = {};

  let isMissing = false;
  for (const participant of participants) {
    if (participant.Participant.Statut != StatutParticipant.Inscrit) continue;

    const idGroupe = trouveGroupe(l, participant.Personne.DateNaissance);
    if (idGroupe === null) {
      isMissing = true;
    } else {
      sizes[idGroupe] = (sizes[idGroupe] || 0) + 1;
    }
  }

  return { sizes, isMissing };
}
