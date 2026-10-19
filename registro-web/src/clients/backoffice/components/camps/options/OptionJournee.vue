<template>
  <v-card
    subtitle="Le prix est variable en fonction des jours de participation au camp."
  >
    <v-card-text>
      <v-row>
        <v-col cols="3">
          <MontantField
            label="Prix d'une journée"
            :model-value="{
              Cent: modelValue,
              Currency: props.camp.Prix.Currency,
            }"
            @update:model-value="(v) => (modelValue = v.Cent)"
            readonly-currency
          >
          </MontantField>
        </v-col>
      </v-row>
    </v-card-text>
  </v-card>
</template>

<script setup lang="ts">
import { type Camp, type Int } from "@/clients/backoffice/logic/api";
const props = defineProps<{
  camp: Camp;
}>();

const modelValue = defineModel<Int>({ required: true });
</script>
