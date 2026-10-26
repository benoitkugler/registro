<template>
  <v-text-field
    variant="outlined"
    density="compact"
    v-model="modelValue"
    @blur="checkSecuriteSociale"
    label="Sécurité sociale"
    :rules="[
      FormRules.required(`Merci de préciser ton numéro de Sécurité Sociale.`),
    ]"
    :error-messages="infered?.Err ? [infered.Err] : undefined"
  >
    <template #append-inner>
      <v-chip
        v-if="infered && !infered.Err"
        color="green"
        prepend-icon="mdi-check"
      >
        <span>Né{{ Formatters.accord(props.sexe) }} à&nbsp</span
        ><b>
          {{ infered.CommuneNaissance }} ;
          {{ infered.DepartementNaissance }} ({{
            infered.DepartementNaissanceCode
          }})</b
        >
      </v-chip>
    </template>
  </v-text-field>
</template>

<script lang="ts" setup>
import { Sexe, type Date, type SecuriteSocialeCheck } from "../logic/api";
import { Formatters, FormRules } from "@/utils";
import { controller } from "../logic/logic";
import { isDateZero } from "@/components/date";
import { ref } from "vue";

const props = defineProps<{
  token: string;
  sexe: Sexe;
  dateNaissance: Date;
}>();

const modelValue = defineModel<string>({ required: true });

const infered = ref<SecuriteSocialeCheck | null>(null);

// should only be called with asso == 'acve'
async function checkSecuriteSociale() {
  const args = {
    Token: props.token,
    Sexe: props.sexe,
    DateNaissance: props.dateNaissance,
    SecuriteSociale: modelValue.value,
  };
  if (args.Sexe == 0 || isDateZero(args.DateNaissance)) {
    // missing arguments, we can't check the number
    return;
  }
  const res = await controller.ParseSecuriteSociale(args);
  if (res === undefined) return;
  infered.value = res;
}
</script>
