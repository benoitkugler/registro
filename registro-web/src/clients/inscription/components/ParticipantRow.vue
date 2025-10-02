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

          <v-alert type="warning" v-if="avertissementAge != null" class="mb-2">
            {{ participant.Prenom }} aura <b>{{ avertissementAge }} ans</b> au
            début du séjour.
          </v-alert>

          <v-alert v-if="selectedCamp === undefined" class="text-center">
            <i>Veuillez sélectionner un séjour...</i>
          </v-alert>
          <CampCard v-else :camp="selectedCamp"></CampCard>
        </v-col>
      </v-row>
    </v-card-text>
    <v-card-actions>
      <v-spacer></v-spacer>
      <v-btn color="red" @click="emit('delete')" variant="outlined">
        <template #prepend>
          <v-icon>mdi-delete</v-icon>
          Enlever ce participant
        </template>
      </v-btn>
    </v-card-actions>
  </v-card>
</template>

<script lang="ts" setup>
import { computed, ref } from "vue";
import type { CampExt, Date_, Participant, Settings } from "../logic/api";
import { Camps, FormRules } from "@/utils";
import { ageFrom, isDateZero } from "@/components/date";
import CampCard from "./CampCard.vue";

const props = defineProps<{
  camps: CampExt[];
  settings: Settings;
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

// renvoie l'âge en début de camp s'il est invalide, null sinon
const avertissementAge = computed(() => {
  const camp = selectedCamp.value;
  if (camp === undefined || isDateZero(participant.value.DateNaissance))
    return null;
  const ageDebut = ageFrom(
    participant.value.DateNaissance,
    new Date(camp.DateDebut)
  );
  if (ageDebut == null) return null;
  const isInvalid =
    (camp.AgeMin && ageDebut < camp.AgeMin) ||
    (camp.AgeMax && ageDebut > camp.AgeMax);
  return isInvalid ? ageDebut : null;
});
</script>
