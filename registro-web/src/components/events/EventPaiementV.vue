<template>
  <EventItem color="#ad9726" icon="mdi-currency-eur">
    Paiement de {{ props.paiement.Payeur }} d'un montant de
    <b> {{ Formatters.montant(props.paiement.Montant) }} </b>
    <template #append>
      {{ Formatters.date(props.paiement.Date) }}
    </template>
  </EventItem>
</template>

<script setup lang="ts">
import { type Paiement } from "@/clients/backoffice/logic/api";
import { Formatters } from "@/utils";

const props = defineProps<{
  paiement: Paiement;
}>();
</script>

<style scoped></style>
