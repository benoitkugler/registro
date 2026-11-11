<template>
  <v-card subtitle="Configuration">
    <template v-slot:append>
      <v-text-field
        variant="outlined"
        density="compact"
        label="Année"
        v-model.number="inner.Year"
        type="number"
        width="150"
      ></v-text-field>
    </template>
    <v-card-text>
      <v-row>
        <v-col align-self="center" cols="2">
          <v-card-subtitle class="pa-0 ma-0">
            Prix d'un repas <br />
            <small> (pour 1 personne)</small>
          </v-card-subtitle>
        </v-col>
        <v-col align-self="center">
          <EurosField
            v-model="inner.MealsCost[0]"
            label="Petit déjeuner"
          ></EurosField>
        </v-col>
        <v-col align-self="center">
          <EurosField v-model="inner.MealsCost[1]" label="Midi"></EurosField>
        </v-col>
        <v-col align-self="center">
          <EurosField v-model="inner.MealsCost[2]" label="Goûter"></EurosField>
        </v-col>
        <v-col align-self="center">
          <EurosField v-model="inner.MealsCost[3]" label="Soir"></EurosField>
        </v-col>
        <v-col align-self="center">
          <EurosField
            v-model="inner.MealsCost[4]"
            label="Cinquième"
          ></EurosField>
        </v-col>
      </v-row>
      <v-row>
        <v-col>
          <v-text-field
            variant="underlined"
            density="compact"
            type="number"
            min="0"
            v-model.number="inner.SmallMealThreshold"
            label="Seuil 'Petit camp'"
            hint="Applique une majoration si le nombre d'inscrits est strictement inférieur à ce seuil."
            persistent-hint
          >
          </v-text-field>
        </v-col>
        <v-col>
          <v-text-field
            variant="underlined"
            density="compact"
            type="number"
            min="0"
            v-model.number="inner.SmallMealMajoration"
            suffix="%"
            label="Majoration 'Petit camp'"
            hint="Pourcentage ajouté au coût des repas, pour un petit camp."
            persistent-hint
          >
          </v-text-field>
        </v-col>
      </v-row>

      <v-row>
        <v-col>
          <EurosField
            v-model="inner.ChauffageMDR"
            label="Chauffage MdR"
            hint="Coût du chauffage pour une journée (actif de Novembre à Avril)"
          ></EurosField>
        </v-col>
        <v-col>
          <EurosField
            v-model="inner.EntretienMDR"
            label="Entretien MdR"
            hint="Frais d'entretien pour une journée et un participant"
          ></EurosField>
        </v-col>
        <v-col>
          <EurosField
            v-model="inner.MaterielMDR"
            label="Matériel MdR"
            hint="Coût du matériel pour une journée et un participant"
          ></EurosField>
        </v-col>
      </v-row>
      <v-row>
        <v-col>
          <EurosField
            v-model="inner.AssoCars[0]"
            label="Frais Bus"
            hint="Coût d'utilisation, au kilomètre, du bus de l'association."
          ></EurosField>
        </v-col>
        <v-col>
          <EurosField
            v-model="inner.AssoCars[1]"
            label="Frais Minibus"
            hint="Coût d'utilisation, au kilomètre, d'un minibus de l'association."
          ></EurosField>
        </v-col>
        <v-col>
          <EurosField
            v-model="inner.AssoCars[2]"
            label="Frais Véhicule léger"
            hint="Coût d'utilisation, au kilomètre, d'un véhicule léger de l'association."
          ></EurosField>
        </v-col>
      </v-row>
    </v-card-text>
    <v-card-actions>
      <v-spacer></v-spacer>
      <v-btn color="green" @click="emit('update', inner)"> Enregistrer </v-btn>
    </v-card-actions>
  </v-card>
</template>

<script lang="ts" setup>
import { ref } from "vue";
import type { ComptaConfig } from "../../logic/api";
import { copy } from "@/utils";
import EurosField from "./EurosField.vue";

const props = defineProps<{
  config: ComptaConfig;
}>();

const emit = defineEmits<{
  (e: "update", cfg: ComptaConfig): void;
}>();

const inner = ref(copy(props.config));
</script>
