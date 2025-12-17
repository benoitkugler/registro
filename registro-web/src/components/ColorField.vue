<template>
  <v-row>
    <v-col align-self="center" class="ml-2">
      {{ props.label }}
    </v-col>
    <v-col align-self="center" cols="auto">
      <v-menu top :close-on-content-click="false">
        <template #activator="{ props: menuProps }">
          <v-btn :color="modelValue" v-bind="menuProps" width="90px"></v-btn>
        </template>
        <v-card>
          <v-card-text class="pa-0">
            <v-color-picker
              :value="modelValue || '#252EBC'"
              @update:model-value="onInput"
              flat
              :swatches="swatches"
              :show-swatches="showSwatches"
              :hide-canvas="showSwatches"
              :hide-sliders="showSwatches"
              mode="hex"
            />
          </v-card-text>
        </v-card>
      </v-menu>
    </v-col>
  </v-row>
</template>

<script setup lang="ts">
import { computed } from "vue";

const props = defineProps<{
  label: string;
  swatches: string[];
}>();

const modelValue = defineModel<string>({ required: true });

function onInput(v: string) {
  modelValue.value = v.substring(0, 7);
}

const swatches = computed(() => {
  const out: string[][] = [];
  let N = props.swatches.length;
  let chunk = 5;
  for (let i = 0; i < N; i += chunk) {
    out.push(props.swatches.slice(i, i + chunk));
  }
  return out;
});

const showSwatches = computed(() => props.swatches.length != 0);
</script>

<style scoped></style>
