<template>
  <v-card
    subtitle="Prix préférentiel pour une inscription avant une date limite."
  >
    <!-- <template #append>
      <v-btn icon size="x-small" @click="addStatut">
        <v-icon color="green">mdi-plus</v-icon>
      </v-btn>
    </template> -->
    <v-card-text class="py-0">
      <v-row>
        <v-col>
          <DateField
            label="Date limite (incluse)"
            v-model="modelValue.Limite"
          ></DateField>
        </v-col>
        <v-col>
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
