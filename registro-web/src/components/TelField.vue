<template>
  <v-row no-gutters>
    <v-col cols="3">
      <v-combobox
        label="Indicatif"
        variant="outlined"
        density="compact"
        :disabled="props.disabled"
        :items="['+41', '+33']"
        :model-value="parsed.indicatif"
        :hide-details="props.hideDetails"
        @update:model-value="(v) => null"
      ></v-combobox>
    </v-col>
    <v-col cols="9">
      <v-text-field
        class="mr-1"
        variant="outlined"
        density="compact"
        :hide-details="props.hideDetails"
        label="Numéro"
        type="number"
        :disabled="props.disabled"
        :model-value="parsed.number"
        @update:model-value="(v) => null"
      >
      </v-text-field>
    </v-col>
  </v-row>
</template>

<script setup lang="ts">
import type { Tel } from "@/clients/directeurs/logic/api";
import { parseTel } from "@/phones";
import { computed } from "vue";
const props = defineProps<{
  //   label: string;
  hideDetails?: boolean;
  disabled?: boolean;
}>();

const modelValue = defineModel<Tel>({ required: true });

const parsed = computed(() => parseTel(modelValue.value));
</script>

<style scoped></style>
