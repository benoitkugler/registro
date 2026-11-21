<template>
  <v-card>
    <v-card-text>
      <v-row>
        <v-col>
          Prix préférentiel pour une inscription avant une date limite.
        </v-col>
      </v-row>
      <v-row>
        <v-col cols="12">
          <DateField
            label="Date limite (incluse)"
            v-model="modelValue.Limite"
          ></DateField>
        </v-col>
        <v-col cols="12">
          <MontantField
            label="Prix"
            :model-value="{
              Currency: props.camp.Prix.Currency,
              Cent: modelValue.Prix,
            }"
            @update:model-value="(c) => (modelValue.Prix = c.Cent)"
            readonly-currency
          ></MontantField>
        </v-col>
      </v-row>
    </v-card-text>
  </v-card>
</template>

<script setup lang="ts">
import {
  type Camp,
  type InscriptionRapide,
} from "@/clients/backoffice/logic/api";
const props = defineProps<{
  camp: Camp;
}>();

const modelValue = defineModel<InscriptionRapide>({ required: true });
</script>
