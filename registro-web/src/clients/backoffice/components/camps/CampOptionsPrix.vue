<template>
  <v-card
    title="Options sur le prix"
    subtitle="Permet de modifier le prix de base du séjour"
    width="1000px"
  >
    <v-card-text>
      <!-- Quotient familial -->
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
        <v-row
          v-for="(tranche, index) in innerQF"
          :key="index"
          align="center"
          class="my-1"
        >
          <v-col cols="auto">
            <v-btn
              icon="mdi-delete-outline"
              variant="text"
              size="small"
              density="comfortable"
              :disabled="innerQF.length <= 1"
              aria-label="Supprimer la tranche"
              @click="removeTranche(index)"
            ></v-btn>
          </v-col>
          <v-col>
            <IntField
              :label="trancheLabel(index)"
              :min="minQuotient(index)"
              v-model="tranche.QuotientMin"
            ></IntField>
          </v-col>
          <v-col>
            <IntField
              label="Pourcentage"
              suffix="%"
              :min="0 as Int"
              :max="100 as Int"
              :disabled="index === innerQF.length - 1"
              hide-details
              v-model="tranche.Pourcentage"
            ></IntField>
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
            @update:model-value="onKindChange"
          ></v-select>
        </v-col>
        <v-col align-self="center">
          <div
            v-if="innerOption.Active == OptionPrixKind.NoOption"
            class="text-center"
          >
            <i>Aucune modification sur le prix de base.</i>
          </div>
          <OptionInscriptionRapide
            v-if="innerOption.Active == OptionPrixKind.PrixInscriptionRapide"
            :camp="props.camp"
            v-model="innerOption.InscriptionRapide"
          ></OptionInscriptionRapide>
          <OptionJournee
            v-else-if="innerOption.Active == OptionPrixKind.PrixJour"
            :camp="props.camp"
            v-model="innerOption.PrixParJour"
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
import OptionJournee from "./options/OptionJournee.vue";
import OptionInscriptionRapide from "./options/OptionInscriptionRapide.vue";
import { isDateZero } from "@/components/date.ts";
const props = defineProps<{
  camp: Camp;
}>();

const innerOption = defineModel<OptionPrixCamp>("optionPrix", {
  required: true,
});
const innerQF = defineModel<PrixQuotientFamilial>("optionQf", {
  required: true,
});

if (innerQF.value == null) innerQF.value = [];

const isQFActive = computed(() =>
  Camps.isQuotientFamilialActive(innerQF.value),
);

const kindItems = selectItems(OptionPrixKindLabels);

function onQFActiveChange(active: boolean | null) {
  innerQF.value = active ? Camps.defaultQuotientFamilial() : [];
}

function minQuotient(index: number): Int {
  if (index === 0) return 0 as Int;
  return (innerQF.value[index - 1].QuotientMin + 1) as Int;
}

function trancheLabel(index: number): string {
  const min = innerQF.value[index].QuotientMin;
  if (index === innerQF.value.length - 1) return `Plus de ${min}`;
  return `De ${min} à ${innerQF.value[index + 1].QuotientMin - 1}`;
}

function addTranche() {
  const grid = innerQF.value;
  const lastIndex = grid.length - 1;
  const last = grid[lastIndex];
  const previousMin = lastIndex > 0 ? grid[lastIndex - 1].QuotientMin : -1;
  const minBeforeLast = last.QuotientMin - 100;
  if (minBeforeLast > previousMin) {
    grid.splice(lastIndex, 0, {
      QuotientMin: minBeforeLast as Int,
      Pourcentage: 50 as Int,
    });
    return;
  }
  // Pas de place avant la dernière tranche : on étend la grille vers le haut.
  last.Pourcentage = 50 as Int;
  grid.push({
    QuotientMin: (last.QuotientMin + 100) as Int,
    Pourcentage: 100 as Int,
  });
}

function removeTranche(index: number) {
  if (innerQF.value.length <= 1) return;
  innerQF.value.splice(index, 1);
  innerQF.value[innerQF.value.length - 1].Pourcentage = 100 as Int;
}

// used to set better default values
function onKindChange(newKind: OptionPrixKind) {
  if (newKind == OptionPrixKind.PrixInscriptionRapide) {
    const opt = innerOption.value.InscriptionRapide;
    if (isDateZero(opt.Limite) && opt.Prix == 0) {
      innerOption.value.InscriptionRapide = {
        Limite: props.camp.DateDebut,
        Prix: props.camp.Prix.Cent,
      };
    }
  }
}
</script>
