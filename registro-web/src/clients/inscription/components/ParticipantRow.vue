<template>
  <v-card border="sm" rounded="md" variant="text" class="my-1">
    <v-card-text>
      <v-row>
        <v-col sm="4" cols="12" align-self="center">
          <v-row>
            <v-col cols="12">
              <v-text-field
                autofocus
                variant="outlined"
                density="compact"
                v-model="participant.Nom"
                label="Nom"
                :rules="[FormRules.required('Merci de remplir votre nom.')]"
              ></v-text-field>
            </v-col>
            <v-col cols="12">
              <v-text-field
                variant="outlined"
                density="compact"
                v-model="participant.Prenom"
                label="Prénom"
                :rules="[FormRules.required('Merci de remplir votre prénom.')]"
              ></v-text-field>
            </v-col>
          </v-row>
          <v-row>
            <v-col sm="5" cols="12">
              <SexeField
                v-model="participant.Sexe"
                :rules="[FormRules.required('Le sexe est requis.')]"
              ></SexeField>
            </v-col>
            <v-col sm="7" cols="12">
              <DateNaissanceField
                v-model="participant.DateNaissance"
                :rule="checkDateNaissance"
              ></DateNaissanceField>
            </v-col>
          </v-row>
          <v-row v-if="props.settings.AskNationnalite">
            <v-col>
              <NationaliteField
                v-model="participant.Nationnalite"
                hint="Cochez si vous avez la nationalité suisse."
              ></NationaliteField>
            </v-col>
          </v-row>
        </v-col>
        <v-divider vertical thickness="1"></v-divider>
        <v-col sm="8" cols="12" align-self="center">
          <v-row>
            <v-col cols="12">
              <v-select
                :model-value="
                  participant.IdCamp >= 1 ? participant.IdCamp : null
                "
                @update:model-value="(v) => (participant.IdCamp = v)"
                variant="outlined"
                density="compact"
                label="Choix du séjour"
                :items="
                  props.camps.map((c) => ({
                    value: c.Id,
                    title: Camps.label(c),
                  }))
                "
                :rules="[FormRules.required('Merci de choisir un séjour.')]"
              ></v-select>
            </v-col>
          </v-row>

          <v-alert v-if="selectedCamp === undefined" class="text-center">
            <i>Veuillez sélectionner un séjour...</i>
          </v-alert>
          <CampCard v-else :camp="selectedCamp"></CampCard>

          <v-alert
            type="warning"
            v-if="avertissementAge && !avertissementAge.Valid"
          >
            {{ Personnes.prenomN(participant) }} aura
            <b>{{ avertissementAge.Age }} ans</b>
            {{ avertissementAge.Jeune ? "à la fin" : "au début" }}
            du séjour.
          </v-alert>
        </v-col>
      </v-row>
    </v-card-text>
    <v-card-actions>
      <v-spacer></v-spacer>
      <v-btn
        color="red"
        @click="emit('delete')"
        variant="outlined"
        prepend-icon="mdi-delete"
      >
        Enlever ce participant
      </v-btn>
    </v-card-actions>
  </v-card>
</template>

<script lang="ts" setup>
import { computed, ref, watch } from "vue";
import type {
  CampExt,
  Date_,
  Participant,
  ConfigInscription,
  StatutParticipantOut,
} from "../logic/api";
import { Camps, FormRules, Personnes } from "@/utils";
import { isDateZero } from "@/components/date";
import CampCard from "./CampCard.vue";
import { controller } from "../logic/logic";
import { isDate } from "util/types";

const props = defineProps<{
  camps: CampExt[];
  settings: ConfigInscription;
}>();

const emit = defineEmits<{
  (e: "delete"): void;
}>();

const participant = defineModel<Participant>({ required: true });

function checkDateNaissance(d: Date_) {
  if (isDateZero(d)) {
    return "La date de naissance est requise.";
  }
  return true;
}

const selectedCamp = computed(() =>
  props.camps.find((c) => c.Id == participant.value.IdCamp)
);

watch(
  () => [participant.value.DateNaissance, participant.value.IdCamp],
  refreshCheck
);

// renvoie l'âge en début de camp s'il est invalide, null sinon
const avertissementAge = ref<StatutParticipantOut | null>(null);
async function refreshCheck() {
  console.log(
    isDateZero(participant.value.DateNaissance),
    participant.value.DateNaissance
  );

  if (
    isDateZero(participant.value.DateNaissance) ||
    participant.value.IdCamp == 0
  )
    return;
  const res = await controller.CheckParticipant(participant.value);
  if (res === undefined) return; // arg..
  avertissementAge.value = res;
}
</script>
