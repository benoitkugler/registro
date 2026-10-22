<template>
  <v-row no-gutters>
    <v-col :class="{ 'mr-1': smAndUp }" :cols="smAndUp ? undefined : 12">
      <TelField
        :label="props.labels ? props.labels[0] : 'Téléphone'"
        v-model="modelValue[0]"
        :hide-details="props.hideDetails"
        :disabled="props.disabled"
        :rule="props.rules ? props.rules[0] : undefined"
      ></TelField>
    </v-col>
    <v-col :class="{ 'ml-1': smAndUp }" :cols="smAndUp ? undefined : 12">
      <TelField
        :label="props.labels ? props.labels[1] : 'Autre numéro (optionnel)'"
        v-model="modelValue[1]"
        :hide-details="props.hideDetails"
        :disabled="props.disabled"
        :rule="props.rules ? props.rules[1] : undefined"
      ></TelField>
    </v-col>
  </v-row>
</template>

<script setup lang="ts">
import type { Tels } from "@/clients/backoffice/logic/api";
import TelField from "./TelField.vue";
import type { FormRules } from "@/utils";
import { useDisplay } from "vuetify";
const props = defineProps<{
  labels?: [string, string];
  hideDetails?: boolean;
  disabled?: boolean;
  rules?: [FormRules.TelRule, FormRules.TelRule];
}>();

const modelValue = defineModel<Tels>({ required: true });

const { smAndUp } = useDisplay();
</script>

<style scoped></style>
