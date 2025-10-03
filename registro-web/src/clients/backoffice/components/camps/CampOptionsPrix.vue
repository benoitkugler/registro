<template>
  <v-card
    title="Options sur le prix"
    subtitle="Permet de modifier le prix de base du séjour"
    width="1000px"
  >
    <v-card-text>
      <!-- Quotient familial -->
      <v-row>
        <v-col align-self="center">
          <v-checkbox
            :model-value="isQFActive"
            @update:model-value="(b) => (innerQF[3] = (b ? 100 : 0) as Int)"
            hide-details
            label="Prix au quotient familial"
          ></v-checkbox>
        </v-col>
        <v-col align-self="center"
          ><IntField
            suffix="%"
            :min="0 as Int"
            :max="100 as Int"
            :disabled="!isQFActive"
            hide-details
            readonly-currency
            label="entre 1 et 359"
            v-model="innerQF[0]"
          ></IntField
        ></v-col>

        <v-col align-self="center"
          ><IntField
            suffix="%"
            :min="0 as Int"
            :max="100 as Int"
            :disabled="!isQFActive"
            hide-details
            readonly-currency
            label="entre 360 et 564"
            v-model="innerQF[1]"
          ></IntField
        ></v-col>
        <v-col align-self="center"
          ><IntField
            suffix="%"
            :min="0 as Int"
            :max="100 as Int"
            :disabled="!isQFActive"
            hide-details
            readonly-currency
            label="entre 565 et 714"
            v-model="innerQF[2]"
          ></IntField
        ></v-col>
        <v-col align-self="center"
          ><IntField
            suffix="%"
            :min="0 as Int"
            :max="100 as Int"
            :disabled="true"
            hide-details
            readonly-currency
            label="plus de 715"
            v-model="innerQF[3]"
          ></IntField
        ></v-col>
      </v-row>

      <v-divider thickness="1" class="my-2"></v-divider>

      <!-- Option -->
      <v-row class="my-2">
        <v-col align-self="center" cols="4">
          <v-select
            variant="outlined"
            density="comfortable"
            hide-details
            :items="kindItems"
            label="Option sur le prix"
            v-model="innerOption.Active"
          ></v-select>
        </v-col>
        <v-col align-self="center">
          <div
            v-if="innerOption.Active == OptionPrixKind.NoOption"
            class="text-center"
          >
            <i>Aucune modification sur le prix de base.</i>
          </div>
          <OptionStatut
            v-else-if="innerOption.Active == OptionPrixKind.PrixStatut"
            :camp="props.camp"
            v-model="innerOption.Statuts"
          ></OptionStatut>
          <OptionJournee
            v-else-if="innerOption.Active == OptionPrixKind.PrixJour"
            :camp="props.camp"
            v-model="innerOption.Jours"
          ></OptionJournee>
        </v-col>
      </v-row>
    </v-card-text>
  </v-card>
</template>

<script setup lang="ts">
import { computed } from "vue";
import {
  OptionPrixKind,
  OptionPrixKindLabels,
  type Camp,
  type Int,
  type OptionPrixCamp,
  type PrixQuotientFamilial,
} from "@/clients/backoffice/logic/api";
import { Camps, selectItems } from "@/utils";
import OptionStatut from "./options/OptionStatut.vue";
import OptionJournee from "./options/OptionJournee.vue";
const props = defineProps<{
  camp: Camp;
}>();

const innerOption = defineModel<OptionPrixCamp>("optionPrix", {
  required: true,
});
const innerQF = defineModel<PrixQuotientFamilial>("optionQf", {
  required: true,
});

const isQFActive = computed(() =>
  Camps.isQuotientFamilialActive(innerQF.value)
);

const kindItems = selectItems(OptionPrixKindLabels);
</script>
