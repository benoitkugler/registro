<template>
  <v-card
    title="Valider l'inscription"
    subtitle="Un mail de confirmation va être envoyé."
  >
    <v-card-text>
      <v-row v-for="p in participants">
        <v-col cols="4" align-self="center">
          <v-list-item
            :title="Personnes.label(p.Personne)"
            :subtitle="Camps.label(p.Camp)"
          ></v-list-item>
        </v-col>
        <v-col cols="3" align-self="center" class="text-center">
          <v-chip
            v-if="
              props.statuts[p.Participant.Id].Statut != ListeAttente.Inscrit
            "
            color="warning"
            prepend-icon="mdi-alert"
          >
            {{ formatStatutCauses(props.statuts[p.Participant.Id]) }}
          </v-chip>
        </v-col>
        <v-col align-self="center">
          <ListeAttenteField
            v-model="inner[p.Participant.Id]"
            hide-details
            :readonly="!isEditable(props.statuts[p.Participant.Id])"
          ></ListeAttenteField>
        </v-col>
      </v-row>
    </v-card-text>
    <v-card-actions>
      <v-spacer></v-spacer>
      <v-btn @click="emit('valide', inner)">Valider</v-btn>
    </v-card-actions>
  </v-card>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import type {
  IdParticipant,
  Inscription,
  StatutCauses,
  IdCamp,
} from "../../clients/backoffice/logic/api";
import { ListeAttente } from "@/clients/directeurs/logic/api";
import { Camps, Personnes } from "@/utils";
import type { BypassRights } from "./types";

const props = defineProps<{
  inscription: Inscription;
  statuts: { [key in IdParticipant]: StatutCauses };
  rights: BypassRights;
  idCamp?: IdCamp; // only edit these participants
}>();

const emit = defineEmits<{
  (e: "valide", params: Statuts): void;
}>();

const participants = computed(() =>
  (props.inscription.Participants || []).filter((p) =>
    props.idCamp ? p.Camp.Id == props.idCamp : true
  )
);

// start with server hints, restricted if needed to participants
const inner = ref(
  Object.fromEntries(
    participants.value.map((p) => [
      p.Participant.Id,
      props.statuts[p.Participant.Id].Statut,
    ])
  ) as Statuts
);

type Statuts = { [key in IdParticipant]: ListeAttente };

function formatStatutCauses(c: StatutCauses) {
  if (!c.AgeMin) {
    return "Trop jeune";
  } else if (!c.AgeMax) {
    return "Trop âgé";
  } else if (!c.EquilibreGF) {
    return "Equilibre G./F.";
  } else if (!c.Place) {
    return "Camp complet";
  } else {
    return "";
  }
}

// use BypassRights to check if the status are editable
function isEditable(s: StatutCauses) {
  // there is only 3 values the server may return
  switch (s.Statut) {
    case ListeAttente.AttenteProfilInvalide:
      return props.rights.ageInvalide;
    case ListeAttente.AttenteCampComplet:
      return props.rights.campComplet;
    case ListeAttente.Inscrit:
      return true;
    default:
      return true; // should not happen
  }
}
</script>
