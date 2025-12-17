<template>
  <v-slider
    hide-details
    :min="minDays"
    :max="maxDays"
    thumb-label
    :model-value="toInt(model)"
    @update:model-value="(v) => (model = fromInt(v as Int))"
  >
    <template #thumb-label>
      {{ Formatters.dateNaissance(model) }}
    </template>
  </v-slider>
</template>

<script lang="ts" setup>
import { Formatters } from "@/utils";
import { type Date_, type GroupesOut, type Int } from "../../logic/api";
import { newDate_ } from "@/components/date";

const props = defineProps<{
  groupes: GroupesOut;
}>();

const emit = defineEmits<{
  (e: "refresh"): void;
}>();

const model = defineModel<Date_>({ required: true });

const dayDuration = 1000 * 60 * 60 * 24;
const minDays = toInt(props.groupes.MinHint) - 365;
const maxDays = toInt(props.groupes.MaxHint) + 365;

function toInt(d: Date_) {
  return new Date(d).valueOf() / dayDuration;
}
function fromInt(v: Int) {
  return newDate_(new Date(dayDuration * v));
}
</script>
