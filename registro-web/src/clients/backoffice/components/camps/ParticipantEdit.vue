<template>
  <v-card title="Editer le participant">
    <v-card-text>
      <v-form>
        <v-row v-if="!props.hidePersonneDossier">
          <v-col>
            <SelectPersonne
              label="Personne"
              v-model="inner.IdPersonne"
              :initial-personne="Personnes.label(props.personne)"
              :api="{
                SelectPersonne: controller.SelectPersonne.bind(controller),
              }"
            ></SelectPersonne>
          </v-col>
          <v-col>
            <SelectDossier v-model="inner.IdDossier"></SelectDossier>
          </v-col>
        </v-row>
        <v-row>
          <v-col>
            <StatutParticipantField
              v-model="inner.Statut"
              hide-details
              :readonly="props.readonlyStatut"
            ></StatutParticipantField>
          </v-col>
        </v-row>
      </v-form>
    </v-card-text>
    <v-card-actions>
      <v-spacer></v-spacer>
      <v-btn :disabled="!areFieldsValid" @click="emit('save', inner)"
        >Enregistrer</v-btn
      >
    </v-card-actions>
  </v-card>
</template>

<script setup lang="ts">
import { ref, computed } from "vue";
import {
  type Participant,
  type Personne,
} from "@/clients/backoffice/logic/api";
import { copy, Personnes } from "@/utils";
import { controller } from "../../logic/logic";
const props = defineProps<{
  participant: Participant;
  personne: Personne;
  hidePersonneDossier?: boolean;
  readonlyStatut?: boolean;
}>();
const emit = defineEmits<{
  (e: "save", camp: Participant): void;
}>();
const inner = ref(copy(props.participant));

const areFieldsValid = computed(
  () => !!(inner.value.IdPersonne && inner.value.IdDossier),
);
</script>
