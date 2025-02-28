<template>
  <v-row no-gutters>
    <v-col :cols="props.readonlyCurrency ? 12 : 6">
      <v-text-field
        class="mr-1"
        variant="outlined"
        density="compact"
        :hide-details="props.hideDetails"
        :label="props.label"
        type="number"
        :disabled="props.disabled"
        :model-value="modelValue.Cent / 100"
        @update:model-value="
          (v) =>
            (modelValue = {
              Cent: round(Number(v) * 100),
              Currency: modelValue.Currency,
            })
        "
        :suffix="
          props.readonlyCurrency
            ? CurrencyLabels[modelValue.Currency]
            : undefined
        "
      >
      </v-text-field>
    </v-col>
    <v-col cols="6" v-if="!props.readonlyCurrency">
      <v-select
        label="Monnaie"
        variant="outlined"
        density="compact"
        :disabled="props.disabled"
        :items="currencyItems"
        :model-value="modelValue.Currency"
        :hide-details="props.hideDetails"
        @update:model-value="
          (v) => (modelValue = { Cent: modelValue.Cent, Currency: v })
        "
      ></v-select>
    </v-col>
  </v-row>
</template>

<script setup lang="ts">
import { CurrencyLabels, type Montant } from "@/clients/backoffice/logic/api";
import { round, selectItems } from "@/utils";
const props = defineProps<{
  label: string;
  hideDetails?: boolean;
  readonlyCurrency?: boolean;
  disabled?: boolean;
}>();

const modelValue = defineModel<Montant>({ required: true });

const currencyItems = selectItems(CurrencyLabels);
</script>

<style scoped></style>
