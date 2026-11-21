<template>
  <v-text-field
    :label="props.label"
    density="compact"
    variant="outlined"
    type="number"
    :min="props.min === undefined ? 1 : props.min"
    :max="props.max"
    :model-value="model"
    @update:model-value="
      (s: string) => {
        if (!isValid(s)) return;
        props.debounce ? throughDebounce(s) : (model = Number(s) as Int);
      }
    "
    :hide-details="!hint"
    :hint="props.hint"
    :persistent-hint="props.persistentHint"
    :disabled="props.disabled"
    :suffix="props.suffix"
  ></v-text-field>
</template>

<script setup lang="ts">
import type { Int } from "@/clients/backoffice/logic/api";
import { round } from "@/utils";
const props = defineProps<{
  label: string;
  hint?: string;
  min?: Int;
  max?: Int;
  disabled?: boolean;
  suffix?: string;
  persistentHint?: boolean;
  debounce?: boolean;
}>();

const model = defineModel<Int>({ required: true });

// debounce feature
let timerId: ReturnType<typeof setTimeout>;
function throughDebounce(s: string) {
  const debounceDelay = 500;

  // cancel pending call
  clearTimeout(timerId);

  const v = round(Number(s));

  // delay new call 500ms
  timerId = setTimeout(() => (model.value = v), debounceDelay);
}

function isValid(s: string) {
  const val = round(Number(s));
  if (props.min !== undefined) {
    if (val < props.min) return false;
  }
  if (props.max !== undefined) {
    if (val > props.max) return false;
  }
  return true;
}
</script>
