<template>
  <div>
    <v-row align="center">
      <v-col align-self="center" cols="auto">
        <v-checkbox
          :model-value="isQFActive"
          @update:model-value="onQFActiveChange"
          hide-details
          label="Prix au quotient familial"
        ></v-checkbox>
      </v-col>
      <v-col v-if="isQFActive" class="text-caption">
        La dernière tranche est fixée à 100 % du prix de base.
      </v-col>
    </v-row>

    <template v-if="isQFActive">
      <v-row v-for="(tranche, index) in innerQF" :key="index" class="my-1">
        <v-col cols="3" align-self="center">
          {{ trancheLabel(index) }}
        </v-col>
        <v-col>
          <IntField
            label="Minimum de la tranche"
            :min="minQuotient(index)"
            :max="maxQuotient(index)"
            v-model="tranche.QuotientMin"
          ></IntField>
        </v-col>
        <v-col>
          <IntField
            label="Pourcentage"
            suffix="%"
            :min="0 as Int"
            :max="100 as Int"
            :disabled="index === innerQF!.length - 1"
            hide-details
            v-model="tranche.Pourcentage"
          ></IntField>
        </v-col>
        <v-col cols="auto" align-self="center">
          <v-btn
            icon="mdi-delete-outline"
            variant="text"
            size="small"
            density="comfortable"
            :disabled="innerQF!.length <= 1"
            aria-label="Supprimer la tranche"
            @click="removeTranche(index)"
          ></v-btn>
        </v-col>
      </v-row>

      <v-row class="mt-1">
        <v-col cols="auto">
          <v-btn
            prepend-icon="mdi-plus"
            variant="tonal"
            size="small"
            @click="addTranche"
            >Ajouter une tranche</v-btn
          >
        </v-col>
      </v-row>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import {
  type Int,
  type PrixQuotientFamilial,
} from "@/clients/backoffice/logic/api";
import { Camps } from "@/utils";
const props = defineProps<{}>();

const innerQF = defineModel<PrixQuotientFamilial>({
  required: true,
});

const isQFActive = computed(() =>
  Camps.isQuotientFamilialActive(innerQF.value),
);

function onQFActiveChange(active: boolean | null) {
  innerQF.value = active ? Camps.defaultQuotientFamilial() : [];
}

function minQuotient(index: number) {
  if (index === 0 || !innerQF.value) return 0 as Int;
  return (innerQF.value[index - 1].QuotientMin + 1) as Int;
}

function maxQuotient(index: number) {
  if (!innerQF.value || index == innerQF.value.length - 1) return undefined;
  return (innerQF.value[index + 1].QuotientMin - 1) as Int;
}

function trancheLabel(index: number): string {
  if (!innerQF.value) return "";

  const min = innerQF.value[index].QuotientMin;
  if (index === innerQF.value.length - 1) return `Plus de ${min}`;
  return `De ${min} à ${innerQF.value[index + 1].QuotientMin - 1}`;
}

function addTranche() {
  const grid = innerQF.value || [];
  if (grid.length == 0) {
    grid.push({ QuotientMin: 400 as Int, Pourcentage: 50 as Int });
    return;
  }

  const maxQF = Math.max(...grid.map((v) => v.QuotientMin));
  grid.push({
    QuotientMin: (maxQF + 100) as Int,
    Pourcentage: 100 as Int,
  });
}

function removeTranche(index: number) {
  if (!innerQF.value || innerQF.value.length <= 1) return;
  innerQF.value.splice(index, 1);
  innerQF.value[innerQF.value.length - 1].Pourcentage = 100 as Int;
}
</script>
