<template>
  <v-row no-gutters>
    <v-col class="mr-1">
      <TelField
        :label="props.labels ? props.labels[0] : 'Téléphone'"
        v-model="modelValue[0]"
        :hide-details="props.hideDetails"
        :disabled="props.disabled"
        :rule="props.rules ? props.rules[0] : undefined"
      ></TelField>
    </v-col>
    <v-col class="ml-1">
      <TelField
        :label="props.labels ? props.labels[1] : 'Autre numéro (optionnel)'"
        v-model="modelValue[1]"
        :hide-details="props.hideDetails"
        :disabled="props.disabled"
        :rule="props.rules ? props.rules[1] : undefined"
      ></TelField>
    </v-col>
  </v-row>
  <!-- <v-row>
    <v-col>
      <v-row v-for="(tel, index) in modelValue">
        <v-col align-self="center">
          <TelField
            label="Téléphone"
            v-model="modelValue![index]"
            hide-details
          ></TelField>
        </v-col>
        <v-col cols="auto" align-self="center">
          <v-btn
            icon="mdi-delete"
            size="x-small"
            color="red"
            @click="modelValue?.splice(index, 1)"
          ></v-btn>
        </v-col>
      </v-row>
    </v-col>
    <v-col cols="auto" align-self="center">
      <v-btn
        icon="mdi-plus"
        size="x-small"
        color="green"
        @click="modelValue?.push('')"
      ></v-btn>
    </v-col>
  </v-row> -->
</template>

<script setup lang="ts">
import type { Tels } from "@/clients/backoffice/logic/api";
import TelField from "./TelField.vue";
import type { FormRules } from "@/utils";
const props = defineProps<{
  labels?: [string, string];
  hideDetails?: boolean;
  disabled?: boolean;
  rules?: [FormRules.TelRule, FormRules.TelRule];
}>();

const modelValue = defineModel<Tels>({ required: true });
</script>

<style scoped></style>
