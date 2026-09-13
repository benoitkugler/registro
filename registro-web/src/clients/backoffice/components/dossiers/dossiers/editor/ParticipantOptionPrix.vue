<template>
  <v-card title="Prix variable" width="800px">
    <v-card-text>
      <v-row justify="space-evenly">
        <v-col align-self="center" cols="3">
          <IntField
            v-model="inner.quotientFamilial"
            label="Quotient familial"
            hide-details
          ></IntField>
        </v-col>

        <v-divider thickness="2" vertical></v-divider>

        <v-col align-self="center" cols="8">
          <OptionInscriptionRapide
            v-if="
              props.optionCamp.Active == OptionPrixKind.PrixInscriptionRapide
            "
            :camp="props.camp"
            :option="props.optionCamp"
          ></OptionInscriptionRapide>
          <OptionJournee
            v-else-if="props.optionCamp.Active == OptionPrixKind.PrixJour"
            :camp="props.camp"
            :option="props.optionCamp"
            v-model="inner.options.Jour"
          ></OptionJournee>
          <OptionStatut
            v-else-if="props.optionCamp.Active == OptionPrixKind.PrixStatut"
            :camp="props.camp"
            :option="props.optionCamp"
            v-model="inner.options.IdStatut"
          ></OptionStatut>
          <div v-else>Le camp ne propose pas d'option sur le prix.</div>
        </v-col>
      </v-row>
    </v-card-text>
    <v-card-actions>
      <v-spacer></v-spacer>
      <v-btn @click="emit('update', inner)">Enregistrer</v-btn>
    </v-card-actions>
  </v-card>
</template>

<script setup lang="ts">
import {
  OptionPrixKind,
  type CampItem,
  type Int,
  type OptionPrixCamp,
  type OptionPrixParticipant,
} from "@/clients/backoffice/logic/api";
import OptionJournee from "./options/OptionJournee.vue";
import OptionStatut from "./options/OptionStatut.vue";
import { ref, watch } from "vue";
import { copy } from "@/utils";
import OptionInscriptionRapide from "./options/OptionInscriptionRapide.vue";

type OptionPrixAndQF = {
  options: OptionPrixParticipant;
  quotientFamilial: Int;
};

const props = defineProps<{
  camp: CampItem;
  optionCamp: OptionPrixCamp;
  optionAndQf: OptionPrixAndQF;
}>();

const emit = defineEmits<{
  (e: "update", optionAndQF: OptionPrixAndQF): void;
}>();

const inner = ref(copy(props.optionAndQf));

watch(
  () => props.optionAndQf,
  () => (inner.value = copy(props.optionAndQf)),
);
</script>
