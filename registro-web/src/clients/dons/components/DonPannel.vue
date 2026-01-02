<template>
  <v-card title="Modifier le don">
    <template #append> </template>
    <v-card-text>
      <v-row>
        <v-col></v-col>
      </v-row>
      <v-row>
        <v-col>
          <MontantField v-model="inner.Montant" label="Montant"></MontantField>
        </v-col>
        <v-col>
          <ModePaiementField v-model="inner.ModePaiement"></ModePaiementField>
        </v-col>
      </v-row>
      <v-row>
        <v-col>
          <v-text-field
            v-model="inner.Details"
            label="Détails"
            density="compact"
            variant="outlined"
          ></v-text-field>
        </v-col>
        <v-col>
          <v-combobox
            label="Affectation"
            :items="props.affectationHints"
            v-model="inner.Affectation"
            density="compact"
            variant="outlined"
          ></v-combobox>
        </v-col>
      </v-row>
      <v-row>
        <v-col>
          <DateField v-model="inner.Date" label="Date"></DateField>
        </v-col>
        <v-col>
          <v-checkbox
            density="compact"
            v-model="inner.Remercie"
            label="Remercié ?"
            hide-details
          ></v-checkbox>
        </v-col>
      </v-row>
    </v-card-text>
    <v-card-actions>
      <v-spacer></v-spacer>
      <v-btn @click="emit('save', inner)">Enregistrer</v-btn>
    </v-card-actions>
  </v-card>
</template>

<script lang="ts" setup>
import { ref } from "vue";
import { type Don } from "../logic/api";
import { copy } from "@/utils";

const props = defineProps<{
  don: Don;
  affectationHints: string[];
}>();

const emit = defineEmits<{
  (e: "save", don: Don): void;
}>();

const inner = ref(copy(props.don));
</script>
