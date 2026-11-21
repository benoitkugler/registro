<template>
  <v-card>
    <v-card-text>
      <v-row>
        <v-col>
          Le prix varie selon l'âge du participant au premier jour du séjour. À
          partir de la dernière tranche, le prix de base du séjour s'applique.
        </v-col>
      </v-row>
      <v-row v-for="(tranche, index) in tranches" :key="index" class="my-1">
        <v-col cols="4" align-self="center">
          {{ trancheLabel(index) }}
        </v-col>
        <v-col>
          <IntField
            label="Âge limite"
            :min="minAge(index)"
            :max="maxAge(index)"
            hide-details
            v-model="tranche.AgeMax"
          ></IntField>
        </v-col>
        <v-col>
          <MontantField
            label="Prix"
            :model-value="{
              Cent: tranche.Prix,
              Currency: props.camp.Prix.Currency,
            }"
            @update:model-value="(v) => (tranche.Prix = v.Cent)"
            readonly-currency
            hide-details
          ></MontantField>
        </v-col>
        <v-col cols="auto" align-self="center">
          <v-btn
            icon="mdi-delete-outline"
            variant="text"
            size="small"
            density="comfortable"
            :disabled="tranches.length <= 1"
            aria-label="Supprimer la tranche"
            @click="removeTranche(index)"
          ></v-btn>
        </v-col>
      </v-row>
      <v-row>
        <v-col cols="4" align-self="center">
          {{ trancheLabel(tranches.length) }}
        </v-col>
        <v-col></v-col>
        <v-col>
          {{ Formatters.montant(props.camp.Prix) }}
        </v-col>
      </v-row>
      <v-row class="mt-1">
        <v-col cols="auto">
          <v-btn
            prepend-icon="mdi-plus"
            variant="tonal"
            size="small"
            @click="addTranche"
            >Ajouter une tranche</v-btn
          >
        </v-col>
      </v-row>
    </v-card-text>
  </v-card>
</template>

<script setup lang="ts">
import { computed } from "vue";
import {
  type Camp,
  type Int,
  type PrixAgeTranche,
  type PrixParAge,
} from "@/clients/backoffice/logic/api";
import { Formatters } from "@/utils";

const props = defineProps<{
  camp: Camp;
}>();

const modelValue = defineModel<PrixParAge>({ required: true });

const tranches = computed(() => modelValue.value ?? []);

function ensureGrid(): PrixAgeTranche[] {
  if (!modelValue.value) modelValue.value = [];
  return modelValue.value;
}

function minAge(index: number): Int {
  if (index === 0) return 1 as Int;
  return (tranches.value[index - 1].AgeMax + 1) as Int;
}

function maxAge(index: number): Int | undefined {
  const next = tranches.value[index + 1];
  return next ? ((next.AgeMax - 1) as Int) : undefined;
}

function trancheLabel(index: number) {
  return Formatters.ageTranche(tranches.value, index);
}

function addTranche() {
  const grid = ensureGrid();
  const maxAge = Math.max(0, ...grid.map((t) => t.AgeMax));
  const previous = grid[grid.length - 1];
  grid.push({
    AgeMax: (maxAge + 1) as Int,
    Prix: previous ? previous.Prix : props.camp.Prix.Cent,
  });
}

function removeTranche(index: number) {
  if (tranches.value.length <= 1) return;
  ensureGrid().splice(index, 1);
}
</script>
