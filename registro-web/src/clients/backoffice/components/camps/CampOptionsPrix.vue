<template>
  <v-card
    title="Options sur le prix"
    subtitle="Permet de modifier le prix de base du séjour"
  >
    <v-card-text>
      <v-row>
        <v-col align-self="center" cols="7">
          <!-- Quotient familial -->
          <OptionQuotientFamilial v-model="innerQF"></OptionQuotientFamilial>
        </v-col>
        <v-divider vertical thickness="1" class="my-2"></v-divider>
        <v-col align-self="center">
          <!-- Option -->
          <v-row>
            <v-col align-self="center">
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
          </v-row>
          <v-row class="my-2">
            <v-col align-self="center">
              <div
                v-if="innerOption.Active == OptionPrixKind.NoOption"
                class="text-center"
              >
                <i>Aucune modification sur le prix de base.</i>
              </div>
              <OptionInscriptionRapide
                v-if="
                  innerOption.Active == OptionPrixKind.PrixInscriptionRapide
                "
                :camp="props.camp"
                v-model="innerOption.InscriptionRapide"
              ></OptionInscriptionRapide>
              <OptionJournee
                v-else-if="innerOption.Active == OptionPrixKind.PrixJour"
                :camp="props.camp"
                v-model="innerOption.PrixParJour"
              ></OptionJournee>
              <OptionPrixAge
                v-else-if="innerOption.Active == OptionPrixKind.PrixAge"
                :camp="props.camp"
                v-model="innerOption.PrixParAge"
              ></OptionPrixAge>
            </v-col>
          </v-row>
        </v-col>
      </v-row>
    </v-card-text>
  </v-card>
</template>

<script setup lang="ts">
import {
  OptionPrixKind,
  OptionPrixKindLabels,
  type Camp,
  type Int,
  type OptionPrixCamp,
  type PrixQuotientFamilial,
} from "@/clients/backoffice/logic/api";
import { selectItems } from "@/utils";
import OptionJournee from "./options/OptionJournee.vue";
import OptionPrixAge from "./options/OptionPrixAge.vue";
import OptionInscriptionRapide from "./options/OptionInscriptionRapide.vue";
import { isDateZero } from "@/components/date.ts";
import OptionQuotientFamilial from "./options/OptionQuotientFamilial.vue";
const props = defineProps<{
  camp: Camp;
}>();

const innerOption = defineModel<OptionPrixCamp>("optionPrix", {
  required: true,
});
const innerQF = defineModel<PrixQuotientFamilial>("optionQf", {
  required: true,
});

const kindItems = selectItems(OptionPrixKindLabels);

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
  if (newKind == OptionPrixKind.PrixAge && !innerOption.value.PrixParAge) {
    innerOption.value.PrixParAge = [
      {
        AgeMax: (props.camp.AgeMax + 1) as Int,
        Prix: props.camp.Prix.Cent,
      },
    ];
  }
}
</script>
