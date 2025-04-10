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
      <v-row class="mt-2">
        <v-col>Prix des séjours</v-col>
        <v-col cols="4" class="text-right">
          {{ dossier.Bilan.Demande }}
        </v-col>
      </v-row>
      <v-row class="my-0">
        <v-col>Dont aides extérieures déduites</v-col>
        <v-col cols="4" class="text-right"> {{ dossier.Bilan.Aides }} </v-col>
      </v-row>
      <v-divider thickness="1"></v-divider>
      <v-row class="my-0">
        <v-col>Paiements</v-col>
        <v-col cols="4" class="text-right">
          {{ dossier.Bilan.Recu }}
        </v-col>
      </v-row>
      <v-divider thickness="1"></v-divider>
      <v-row class="my-0">
        <v-col>Montant restant à régler</v-col>
        <v-col cols="4" class="text-right">
          <b
            :class="
              dossier.Bilan.Statut == StatutPaiement.Complet
                ? 'text-green'
                : 'text-orange'
            "
          >
            {{ dossier.Bilan.Restant }}
          </b>
        </v-col>
      </v-row>
    </v-card-text>

    <v-dialog></v-dialog>
  </v-card>
</template>

<script setup lang="ts">
import { type Participant } from "@/clients/backoffice/logic/api";
import { StatutPaiement, type DossierExt } from "../logic/api";
const props = defineProps<{
  dossier: DossierExt;
}>();

const emit = defineEmits<{
  (e: "save", participants: Participant[]): void;
}>();
</script>
