<template>
  <v-slider
    :label="props.label"
    v-model="modelValue"
    :min="props.min"
    :max="props.max"
    step="1"
    show-ticks
    thumb-label
    color="primary"
    hide-details
  >
    <template #append>
      {{ Formatters.date(resolvedDate) }}
    </template>
  </v-slider>
</template>

<script setup lang="ts">
import type { Date_, Int } from "@/clients/backoffice/logic/api";
import { computed } from "vue";
import { addDays } from "./date";
import { Formatters } from "@/utils";
const props = defineProps<{
  label: string;
  min: number;
  max: number;
  refDate: Date_;
  readonly?: boolean;
}>();

const modelValue = defineModel<Int>({ required: true });

const resolvedDate = computed(() => addDays(props.refDate, modelValue.value));
</script>
