<template>
  <!-- Two modes : SELECT an existing one or CREATE a new one-->
  <v-row>
    <v-col>
      <v-switch
        color="primary"
        label="Utiliser un taux existant"
        :model-value="modelValue.Id > 0"
        @update:model-value="onSwitch"
        hide-details
      ></v-switch>
    </v-col>
  </v-row>
  <v-row>
    <v-col>
      <v-select
        variant="outlined"
        density="compact"
        hide-details
        label="Taux"
        :items="items"
        v-model="selected"
        @update:model-value="onSelect"
        :disabled="modelValue.Id <= 0"
      ></v-select>
    </v-col>
  </v-row>
  <v-row>
    <v-col>
      <v-text-field
        variant="outlined"
        density="compact"
        label="Label"
        v-model="modelValue.Label"
        :disabled="modelValue.Id > 0"
        :error-messages="labelError"
      ></v-text-field>
    </v-col>
    <!-- <v-col>
      <IntField
        :label="CurrencyLabels[Currency.Euros]"
        v-model="modelValue.Euros"
        :disabled="modelValue.Id > 0"
      ></IntField>
    </v-col> -->
    <v-col>
      <v-text-field
        :label="`Valeur d'un ${CurrencyLabels[Currency.FrancsSuisse]}`"
        density="compact"
        variant="outlined"
        type="number"
        min="0.001"
        suffix="€"
        hint="Valeur en € d'une unité de la monnaie."
        :model-value="modelValue.FrancsSuisse / 1000"
        @update:model-value="
          (v) => (modelValue.FrancsSuisse = round(Number(v) * 1000))
        "
        :disabled="modelValue.Id > 0"
      ></v-text-field>
    </v-col>
  </v-row>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from "vue";
import {
  Currency,
  CurrencyLabels,
  type IdTaux,
  type Int,
  type Taux,
} from "@/logic/app/api";
import { controller, copy, mapFromObject, round } from "@/logic/app/logic";

onMounted(fetch);

const modelValue = defineModel<Taux>({ required: true });

const selected = ref<IdTaux | null>(null);

const items = computed(() => {
  const out = Array.from(tauxMap.value.values()).map((t) => ({
    title: t.Label,
    value: t.Id,
  }));
  out.sort((a, b) => a.title.localeCompare(b.title));
  return out;
});
const tauxMap = ref(new Map<IdTaux, Taux>());
async function fetch() {
  const res = await controller.CampsGetTaux();
  if (res === undefined) return;
  tauxMap.value = mapFromObject(res);

  if (modelValue.value.Id > 0) {
    modelValue.value = copy(tauxMap.value.get(modelValue.value.Id)!);
    selected.value = modelValue.value.Id;
  }
}

function onSwitch(useExistant: boolean | null) {
  if (useExistant) {
    if (selected.value == null) {
      selected.value = 1 as Int; // defaut
    }
    modelValue.value = copy(tauxMap.value.get(selected.value)!);
  } else {
    selected.value = null;
    modelValue.value.Id = 0 as Int;
    modelValue.value.Label = "";
  }
}

function onSelect(id: IdTaux) {
  modelValue.value = copy(tauxMap.value.get(id)!);
}

const labelError = computed(() => {
  if (modelValue.value.Id > 0) return null;
  const newLabel = modelValue.value.Label;
  const isTaken = Array.from(tauxMap.value.values())
    .map((t) => t.Label)
    .includes(newLabel);
  if (isTaken) return "Ce label est déjà utilisé.";
  return null;
});
</script>
