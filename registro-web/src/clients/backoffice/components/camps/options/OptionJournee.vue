<template>
  <v-card
    subtitle="Le prix est variable en fonction des jours de participation au camp."
  >
    <template v-slot:append>
      <v-tooltip text="Appliquer le prix du premier jour à tous.">
        <template v-slot:activator="{ props: btnProps }">
          <v-btn
            v-bind="btnProps"
            icon="mdi-skip-next"
            size="x-small"
            :disabled="!(modelValue || [])[0]"
            @click="propagateFirst"
          ></v-btn>
        </template>
      </v-tooltip>
    </template>
    <v-card-text>
      <v-row>
        <v-col cols="3" v-for="(value, index) in modelValue">
          <MontantField
            :label="`Jour ${index + 1}`"
            :model-value="{
              Cent: value,
              Currency: props.camp.Prix.Currency,
            }"
            @update:model-value="(v) => (modelValue![index] = v.Cent)"
            readonly-currency
          >
          </MontantField>
        </v-col>
      </v-row>
    </v-card-text>
  </v-card>
</template>

<script setup lang="ts">
import {
  type Camp,
  type Int,
  type Jours,
  type Montant,
} from "@/clients/backoffice/logic/api";
import { watch } from "vue";
const props = defineProps<{
  camp: Camp;
}>();

const modelValue = defineModel<Jours>({ required: true });

watch(() => modelValue.value, ensureSize, { immediate: true });

function ensureSize() {
  const duree = props.camp.Duree;
  if (modelValue.value?.length != duree) {
    modelValue.value = Array.from({ length: duree }).map((i) => 0 as Int);
  }
}

function propagateFirst() {
  const l = modelValue.value!;
  const val = l[0];
  // propagate the first field to the other
  for (let i = 0; i < l.length; i++) {
    l[i] = val;
  }
}
</script>
