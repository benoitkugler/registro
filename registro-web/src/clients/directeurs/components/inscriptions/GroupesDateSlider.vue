<template>
  <!-- in read mode, use the range slider, and switch to regular slider when editing -->
  <v-hover>
    <template #default="{ isHovering, props: innerProps }">
      <v-range-slider
        v-if="!isHovering"
        v-bind="innerProps"
        hide-details
        :min="minDays"
        :max="maxDays"
        :model-value="[
          isDateZero(start) ? minDays : toInt(start),
          toInt(model),
        ]"
        color="primary"
      >
      </v-range-slider>
      <v-slider
        v-else
        v-bind="innerProps"
        color="primary"
        hide-details
        :min="minDays"
        :max="maxDays"
        thumb-label
        :model-value="toInt(model)"
        @update:model-value="(v) => (model = fromInt(v as Int))"
      >
        <template v-slot:thumb-label="{ modelValue }">
          {{ Formatters.dateNaissance(fromInt(modelValue as Int)) }}
        </template>
      </v-slider>
    </template>
  </v-hover>
</template>

<script lang="ts" setup>
import { Formatters } from "@/utils";
import { type Date_, type GroupesOut, type Int } from "../../logic/api";
import { isDateZero, newDate_ } from "@/components/date";

const props = defineProps<{
  groupes: GroupesOut;
  start: Date_;
}>();

const emit = defineEmits<{
  (e: "refresh"): void;
}>();

const model = defineModel<Date_>({ required: true });

const dayDuration = 1000 * 60 * 60 * 24;
const minDays = toInt(props.groupes.MinHint) - 365; // add one year
const maxDays = toInt(props.groupes.MaxHint) + 365; // add one year

function toInt(d: Date_) {
  return new Date(d).valueOf() / dayDuration;
}
function fromInt(v: Int) {
  return newDate_(new Date(dayDuration * v));
}
</script>
