<template>
  <v-row no-gutters>
    <v-col cols="6">
      <v-text-field
        variant="outlined"
        density="compact"
        :hide-details="props.hideDetails"
        :label="props.label"
        type="number"
        :model-value="modelValue.Cent / 100"
        @update:model-value="
          (v) =>
            (modelValue = {
              Cent: round(Number(v) * 100),
              Currency: modelValue.Currency,
            })
        "
      >
      </v-text-field>
    </v-col>
    <v-col cols="6">
      <v-select
        label="Monnaie"
        variant="outlined"
        density="compact"
        :items="currencyItems"
        :model-value="modelValue.Currency"
        @update:model-value="
          (v) => (modelValue = { Cent: modelValue.Cent, Currency: v })
        "
      ></v-select>
    </v-col>
  </v-row>
</template>

<script setup lang="ts">
import {
  Currency,
  CurrencyLabels,
  type Int,
  type Montant,
} from "@/logic/app/api";
import { round } from "@/logic/app/logic";
const props = defineProps<{
  label: string;
  hideDetails?: boolean;
}>();

const modelValue = defineModel<Montant>({ required: true });

const currencyItems = Object.entries(CurrencyLabels).map((pair) => ({
  value: Number(pair[0]) as Currency,
  title: pair[1],
}));
</script>

<style scoped></style>
