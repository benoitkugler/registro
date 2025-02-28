<template>
  <v-card title="Prix variable" width="800px">
    <v-card-text>
      <v-row justify="space-evenly">
        <v-col align-self="center" cols="3">
          <IntField
            v-model="qf"
            label="Quotient familial"
            hide-details
          ></IntField>
        </v-col>

        <v-divider thickness="2" vertical></v-divider>

        <v-col align-self="center" cols="8">
          <OptionJournee
            v-if="props.camp.OptionPrix.Active == OptionPrixKind.PrixJour"
            :camp="props.camp"
            v-model="options.Jour"
          ></OptionJournee>
          <OptionStatut
            v-else-if="
              props.camp.OptionPrix.Active == OptionPrixKind.PrixStatut
            "
            :camp="props.camp"
            v-model="options.IdStatut"
          ></OptionStatut>
          <div v-else>Le camp ne propose pas d'option sur le prix.</div>
        </v-col>
      </v-row>
    </v-card-text>
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

const props = defineProps<{
  camp: Camp;
}>();

const options = defineModel<OptionPrixParticipant>("options", {
  required: true,
});
const qf = defineModel<Int>("quotientFamilial", {
  required: true,
});
</script>
