<template>
  <v-card border="sm" rounded="md" variant="text" class="my-1">
    <v-card-text>
      <v-row>
        <v-col cols="7">
          <v-row>
            <v-col md="4" sm="6">
              <v-text-field
                variant="outlined"
                density="compact"
                v-model="participant.Nom"
                label="Nom"
                :rules="[FormRules.required('Merci de remplir votre nom.')]"
              ></v-text-field>
            </v-col>
            <v-col md="4" sm="6">
              <v-text-field
                variant="outlined"
                density="compact"
                v-model="participant.Prenom"
                label="Prénom"
                :rules="[FormRules.required('Merci de remplir votre prénom.')]"
              ></v-text-field>
            </v-col>
            <v-col md="4" sm="4">
              <SexeField
                v-model="participant.Sexe"
                :rules="[
                  FormRules.required(
                    'Nous avons besoin du sexe des participants pour constituer des groupes.'
                  ),
                ]"
              ></SexeField>
            </v-col>
          </v-row>
          <v-row>
            <v-col md="7" sm="8">
              <DateNaissanceField
                v-model="participant.DateNaissance"
                label="Date de naissance"
                :rule="checkDateNaissance"
              ></DateNaissanceField>
            </v-col>
            <v-col md="5">
              <NationaliteField
                v-model="participant.Nationnalite"
              ></NationaliteField>
            </v-col>
          </v-row>
        </v-col>
        <v-divider vertical thickness="1"></v-divider>
        <v-col>
          <v-row>
            <v-col cols="12">
              <v-select
                v-model="participant.IdCamp"
                variant="outlined"
                density="compact"
                label="Choix du séjour"
                :items="
                  props.camps.map((c) => ({
                    value: c.Id,
                    title: Camps.label(c),
                  }))
                "
              ></v-select>
            </v-col>
          </v-row>
          <v-row>
            <v-col>
              <CampCard
                v-if="selectedCamp !== undefined"
                :camp="selectedCamp"
              ></CampCard>
            </v-col>
          </v-row>
        </v-col>
      </v-row>
    </v-card-text>
    <v-card-actions>
      <v-btn color="red" @click="emit('delete')">
        <template v-slot:prepend>
          <v-icon>mdi-delete</v-icon>
          Enlever ce participant
        </template>
        <v-spacer></v-spacer>
      </v-btn>
    </v-card-actions>
  </v-card>
</template>

<script lang="ts" setup>
import { computed, ref } from "vue";
import type { CampExt, Date_, Participant } from "../logic/api";
import { Camps, FormRules } from "@/utils";
import { isDateZero } from "@/components/date";
import { formatDate } from "@/components/format";
import CampCard from "./CampCard.vue";

const props = defineProps<{
  camps: CampExt[];
}>();

const emit = defineEmits<{
  (e: "delete"): void;
}>();

const participant = defineModel<Participant>({ required: true });

function checkDateNaissance(d: Date_) {
  if (isDateZero(d)) {
    return "Nous avons besoin de l'âge des participants pour constituer des groupes.";
  }
  return true;
}

const selectedCamp = computed(() =>
  props.camps.find((c) => c.Id == participant.value.IdCamp)
);
</script>
