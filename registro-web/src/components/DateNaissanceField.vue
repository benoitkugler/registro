<template>
  <v-text-field
    variant="outlined"
    density="compact"
    placeholder="JJ/MM/AAAA"
    :label="props.label"
    :model-value="inner"
    @update:model-value="onType"
    :readonly="props.readonly"
    :error-messages="props.error ? [props.error] : null"
    :hide-details="props.hideDetails"
    @blur="onBlur"
  >
  </v-text-field>
</template>

<script setup lang="ts">
import type { Date_, Int } from "@/clients/backoffice/logic/api";
import { ref, watch } from "vue";
import { useDate } from "vuetify";
import { autocomplete, parse } from "./date";
const props = defineProps<{
  label: string;
  hideDetails?: boolean;
  readonly?: boolean;
  error?: string;
}>();

const modelValue = defineModel<Date_>({ required: true });

const adapter = useDate();

const inner = ref(
  (adapter.parseISO(modelValue.value) as Date).toLocaleDateString()
);

function onType(s: string) {
  s = autocomplete(s);
  inner.value = s;

  const parsed = parse(s);
  if (parsed === undefined) return;
  modelValue.value = parsed;
}

function onBlur() {
  if (parse(inner.value) === undefined) {
    modelValue.value = "0001-01-01" as Date_;
  }
}
</script>

<style scoped></style>
