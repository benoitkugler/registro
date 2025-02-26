<template>
  <EventItem color="#ad9726" icon="mdi-currency-eur">
    Paiement de {{ props.paiement.Payeur }} d'un montant de
    <b>
      {{ props.paiement.Montant.Cent / 100 }}
      {{ CurrencyLabels[props.paiement.Montant.Currency] }}
    </b>
    <template v-slot:append>
      {{ Formatters.date(props.paiement.Date) }}
    </template>
  </EventItem>
</template>

<script setup lang="ts">
import { CurrencyLabels, type Paiement } from "@/clients/backoffice/logic/api";
import { Formatters } from "@/utils";

const props = defineProps<{
  paiement: Paiement;
}>();
</script>

<style scoped></style>
