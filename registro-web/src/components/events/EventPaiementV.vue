<template>
  <EventItem
    color="#ad9726"
    icon="mdi-currency-eur"
    :time="props.paiement.Time"
  >
    <v-row>
      <v-col align-self="center">
        <v-list-item-title>
          {{
            props.paiement.IsAcompte
              ? "Acompte"
              : props.paiement.IsRemboursement
              ? "Remboursement"
              : "Paiement"
          }}
          <template v-if="!props.paiement.IsRemboursement">
            de <i>{{ props.paiement.Payeur }}</i>
          </template>
          :
          <b> {{ Formatters.montant(props.paiement.Montant) }} </b>
        </v-list-item-title>
        <v-list-item-subtitle>
          {{ props.paiement.Label }}
        </v-list-item-subtitle>
      </v-col>
      <v-col align-self="center" cols="auto">
        <v-btn icon="mdi-pencil" size="x-small" @click="emit('edit')"></v-btn>
      </v-col>
    </v-row>
  </EventItem>
</template>

<script setup lang="ts">
import { type Paiement } from "@/clients/backoffice/logic/api";
import { Formatters } from "@/utils";

const props = defineProps<{
  paiement: Paiement;
}>();

const emit = defineEmits<{
  (e: "edit"): void;
}>();
</script>

<style scoped></style>
