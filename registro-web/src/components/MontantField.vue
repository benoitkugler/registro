<template>
  <v-row no-gutters>
    <v-col :cols="hideCurrencySelect ? 12 : 7">
      <v-text-field
        :autofocus="props.autofocus"
        class="mr-1"
        variant="outlined"
        density="compact"
        :hide-details="props.hideDetails"
        :hint="props.hint"
        :label="props.label"
        type="number"
        :disabled="props.disabled"
        :model-value="modelValue.Cent / 100"
        @update:model-value="
          (v) =>
            props.debounce
              ? throughDebounce(v)
              : (modelValue = {
                  Cent: round(Number(v) * 100),
                  Currency: modelValue.Currency,
                })
        "
        :suffix="
          hideCurrencySelect ? CurrencyLabels[modelValue.Currency] : undefined
        "
      >
      </v-text-field>
    </v-col>
    <v-col cols="5" v-if="!hideCurrencySelect">
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
import { computed } from "vue";
import { CurrencyLabels, type Montant } from "@/clients/backoffice/logic/api";
import { singleCurrency, currencyItems } from "@/settings";
import { round } from "@/utils";
const props = defineProps<{
  label: string;
  hideDetails?: boolean;
  readonlyCurrency?: boolean;
  disabled?: boolean;
  autofocus?: boolean;
  hint?: string;
  debounce?: boolean;
}>();

const modelValue = defineModel<Montant>({ required: true });

// A single supported currency (ACVE has only EUR) doesn't need a selector.
const hideCurrencySelect = computed(
  () => props.readonlyCurrency === true || singleCurrency,
);

// debounce feature
let timerId: ReturnType<typeof setTimeout>;
function throughDebounce(s: string) {
  const debounceDelay = 500;

  // cancel pending call
  clearTimeout(timerId);

  const v = round(Number(s) * 100);

  // delay new call 500ms
  timerId = setTimeout(
    () => (modelValue.value = { Cent: v, Currency: modelValue.value.Currency }),
    debounceDelay,
  );
}
</script>

<style scoped></style>
