<template>
  <v-card title="Modifier le dossier">
    <v-card-text>
      <v-row>
        <v-col>
          <SelectPersonne
            label="Responsable"
            :initial-personne="props.responsable"
            :model-value="innerDossier.IdResponsable"
            @update:model-value="
              (v) => (v != null ? (innerDossier.IdResponsable = v) : undefined)
            "
          ></SelectPersonne>
        </v-col>
      </v-row>
      <v-row>
        <v-col>
          <StringList
            label="Adresses mail en copie"
            v-model="innerDossier.CopiesMails"
            :rule="FormRules.validMails()"
          ></StringList
        ></v-col>
      </v-row>
      <v-row>
        <v-col>
          <v-checkbox
            hide-details
            v-model="innerDossier.PartageAdressesOK"
            label="Partage des coordonnées"
          >
          </v-checkbox
        ></v-col>
      </v-row>
      <v-row>
        <v-col>
          <v-checkbox
            hide-details
            v-model="innerDossier.DemandeFondSoutien"
            label="Demande le fond de soutien"
          >
          </v-checkbox>
        </v-col>
      </v-row>
    </v-card-text>
    <v-card-actions>
      <v-spacer></v-spacer>
      <v-btn @click="emit('save', innerDossier)">Enregistrer</v-btn>
    </v-card-actions>
  </v-card>
</template>

<script setup lang="ts">
import {
  type Dossier,
  type ParticipantExt,
} from "@/clients/backoffice/logic/api";
import { copy, FormRules } from "@/utils";
import { ref } from "vue";

const props = defineProps<{
  responsable: string;
  dossier: Dossier;
}>();

const emit = defineEmits<{
  (e: "save", dossier: Dossier): void;
}>();

const innerDossier = ref(copy(props.dossier));
</script>
