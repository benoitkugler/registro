<template>
  <v-card title="Options de l'entête">
    <v-card-text>
      <v-row>
        <v-col>
          <v-switch
            color="primary"
            density="compact"
            v-model="inner.UseCoordCentre"
            hint="Activer pour remplacer les coordonnées du directeur par ceux du centre, dans l'entête de la lettre."
            persistent-hint
            label="Coordonnées du centre"
          ></v-switch>
        </v-col>
        <v-col cols="5">
          <v-switch
            color="primary"
            density="compact"
            v-show="!inner.UseCoordCentre"
            v-model="inner.ShowAdressePostale"
            hint="Afficher l'adresse postale du directeur."
            persistent-hint
            label="Ajouter l'adresse postale"
          ></v-switch>
        </v-col>
      </v-row>
      <v-row class="mt-4">
        <v-col cols="6">
          <ColorField
            label="Couleur du texte de l'entête"
            v-model="inner.ColorCoord"
            :swatches="swatchColors"
          ></ColorField>
        </v-col>
      </v-row>
    </v-card-text>
    <v-card-actions>
      <v-spacer></v-spacer>
      <v-btn color="success" @click="emit('save', inner)">Enregistrer</v-btn>
    </v-card-actions>
  </v-card>
</template>

<script lang="ts" setup>
import { ref } from "vue";
import { copy } from "@/utils";
import type { LettreOptions } from "../../logic/logic";

const props = defineProps<{
  options: LettreOptions;
}>();

const emit = defineEmits<{
  (e: "save", options: LettreOptions): void;
}>();

const inner = ref(copy(props.options));

/** Colors proposed by tinymce, should be kept in sync */
const swatchColors = [
  "#BFEDD2",
  "#FBEEB8",
  "#F8CAC6",
  "#ECCAFA",
  "#C2E0F4",
  "#2DC26B",
  "#F1C40F",
  "#E03E2D",
  "#B96AD9",
  "#3598DB",
  "#169179",
  "#E67E23",
  "#BA372A",
  "#843FA1",
  "#236FA1",
  "#ECF0F1",
  "#CED4D9",
  "#95A5A6",
  "#7E8C8D",
  "#34495E",
  "#000000",
  "#ffffff",
];
</script>
