<template>
  <v-card-text>
    <template v-if="availableCurrencies.size">
      <v-row>
        <v-col>
          Vous pouvez régler tout ou partie du montant demandé. Le paiement
          minimum est de
          <b>{{
            Formatters.montant({
              Cent: selectedOption.Min,
              Currency: selectedOption.Currency,
            })
          }}</b
          >.
        </v-col>
      </v-row>
      <v-alert
        v-if="selectedOption.Min == selectedOption.Max"
        type="info"
        class="my-3"
      >
        Le montant demandé est inférieur au minimum : seul le montant exact est
        accepté.
      </v-alert>
      <v-row>
        <v-col cols="4" v-if="currencyItems.length > 1">
          <v-select
            v-model="amount.Currency"
            :items="currencyItems"
            label="Devise"
            density="compact"
            variant="outlined"
          ></v-select>
        </v-col>
        <v-col>
          <MontantField
            v-model="amount"
            :min="selectedOption.Min"
            :max="selectedOption.Max"
            label="Montant à payer"
            readonly-currency
          ></MontantField>
        </v-col>
      </v-row>
    </template>
    <p v-else>Aucun paiement en ligne n'est nécessaire.</p>

    <v-card-actions>
      <v-spacer></v-spacer>
      <v-btn
        color="primary"
        class="mt-4"
        :loading="pending"
        :disabled="!canPay"
        @click="emit('start', amount)"
      >
        Payer {{ Formatters.montant(amount) }}
      </v-btn>
    </v-card-actions>
  </v-card-text>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import {
  Currency,
  CurrencyLabels,
  type Data,
  type Montant,
} from "../logic/api";
import { Formatters, selectItems } from "@/utils";
const props = defineProps<{
  paymentOptions: Data["PaiementOptions"];
  pending: boolean;
}>();

const emit = defineEmits<{
  (e: "start", montant: Montant): void;
}>();

const availableCurrencies = computed(
  () =>
    new Set(
      Object.keys(props.paymentOptions || {}).map((s) => Number(s) as Currency),
    ),
);
const currencyItems = computed(() =>
  selectItems(CurrencyLabels).filter((c) =>
    availableCurrencies.value.has(c.value),
  ),
);
const selectedOption = computed(
  () => (props.paymentOptions || {})[amount.value.Currency]!,
);

const amount = ref<Montant>({
  Cent: (props.paymentOptions || {})[currencyItems.value[0].value]!.Max,
  Currency: currencyItems.value[0].value,
});

const canPay = computed(() => {
  const option = selectedOption.value;
  return (
    !props.pending &&
    amount.value.Cent >= option.Min &&
    amount.value.Cent <= option.Max
  );
});
</script>
