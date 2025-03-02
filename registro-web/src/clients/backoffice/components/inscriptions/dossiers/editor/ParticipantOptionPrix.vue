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
          <OptionJournee
            v-if="props.camp.OptionPrix.Active == OptionPrixKind.PrixJour"
            :camp="props.camp"
            v-model="inner.options.Jour"
          ></OptionJournee>
          <OptionStatut
            v-else-if="
              props.camp.OptionPrix.Active == OptionPrixKind.PrixStatut
            "
            :camp="props.camp"
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
  type Camp,
  type Int,
  type OptionPrixParticipant,
} from "@/clients/backoffice/logic/api";
import OptionJournee from "./options/OptionJournee.vue";
import OptionStatut from "./options/OptionStatut.vue";
import { ref, watch } from "vue";
import { copy } from "@/utils";

type OptionPrixAndQF = {
  options: OptionPrixParticipant;
  quotientFamilial: Int;
};

const props = defineProps<{
  camp: Camp;
  optionAndQf: OptionPrixAndQF;
}>();

const emit = defineEmits<{
  (e: "update", optionAndQF: OptionPrixAndQF): void;
}>();

const inner = ref(copy(props.optionAndQf));

watch(
  () => props.optionAndQf,
  () => (inner.value = copy(props.optionAndQf))
);
</script>
