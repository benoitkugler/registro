<template>
  <v-card subtitle="Réglement" class="mt-2">
    <template #append>
      <v-btn size="small">
        <template #prepend>
          <v-icon color="green">mdi-plus</v-icon>
        </template>
        Ajouter une aide
      </v-btn>
    </template>

    <v-card-text>
      <v-row>
        <v-col>Prix des séjours</v-col>
        <v-col cols="4" class="text-right">
          {{ dossier.Bilan.Demande }}
        </v-col>
      </v-row>
      <v-row>
        <v-col>Dont aides extérieures</v-col>
        <v-col cols="4" class="text-right">
          {{ dossier.Bilan.Aides }}
        </v-col>
      </v-row>
      <v-row>
        <v-col>Paiements</v-col>
        <v-col cols="4" class="text-right">
          {{ dossier.Bilan.Recu }}
        </v-col>
      </v-row>
      <v-row>
        <v-col>Montant restant à régler</v-col>
        <v-col cols="4" class="text-right">
          {{ dossier.Bilan.Restant }}
        </v-col>
      </v-row>
    </v-card-text>
  </v-card>
</template>

<script setup lang="ts">
import { type Participant } from "@/clients/backoffice/logic/api";
import type { DossierExt } from "../logic/api";
const props = defineProps<{
  dossier: DossierExt;
}>();

const emit = defineEmits<{
  (e: "save", participants: Participant[]): void;
}>();
</script>
