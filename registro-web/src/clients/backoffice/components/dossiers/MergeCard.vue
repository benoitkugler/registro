<template>
  <v-card
    title="Fusionner le dossier vers"
    subtitle="Les participants, paiements et messages seront copiés vers le dossier cible."
  >
    <v-card-text>
      <v-row>
        <v-col cols="12">
          <SelectDossier v-model="to"></SelectDossier>
        </v-col>
        <v-col cols="12">
          <v-checkbox
            density="compact"
            label="Notifier par email"
            hint="Le reponsable du dossier absorbé sera averti du changement d'espace personnel."
            persistent-hint
            v-model="notifie"
          ></v-checkbox>
        </v-col>
      </v-row>
    </v-card-text>
    <v-card-actions>
      <v-spacer></v-spacer>
      <v-btn
        :disabled="to == 0 || to == props.from"
        @click="
          emit('merge', {
            From: props.from,
            To: to,
            Notifie: notifie,
          })
        "
        >Fusionner</v-btn
      >
    </v-card-actions>
  </v-card>
</template>

<script setup lang="ts">
import { type DossiersMergeIn, type IdDossier } from "../../logic/api";
import { ref } from "vue";

const props = defineProps<{
  from: IdDossier;
}>();

const emit = defineEmits<{
  (e: "merge", args: DossiersMergeIn): void;
}>();

const to = ref<IdDossier>(0 as IdDossier);
const notifie = ref(true);
</script>
