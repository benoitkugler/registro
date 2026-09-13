<template>
  <v-autocomplete
    :density="props.defaultStyle ? undefined : 'compact'"
    :variant="props.defaultStyle ? undefined : 'outlined'"
    :label="props.label"
    :readonly="props.readonly"
    :items="campItems"
    clearable
    :model-value="zeroableToNullable(modelValue)"
    @update:model-value="(id) => (modelValue = nullableToZeroable(id))"
    no-data-text="Aucun séjour n'existe."
    hide-details
  >
  </v-autocomplete>
</template>

<script setup lang="ts">
import type { CampItem, IdCamp } from "@/clients/backoffice/logic/api";
import { Camps, nullableToZeroable, zeroableToNullable } from "@/utils";
import { computed, ref } from "vue";
const props = defineProps<{
  label: string;
  camps: CampItem[];
  readonly?: boolean;
  defaultStyle?: boolean;
}>();

const modelValue = defineModel<IdCamp>({ required: true });

const campItems = computed(() => {
  const newCamps: (
    | { type: "divider" }
    | { type: "subheader"; title: string }
    | { value: IdCamp; title: string }
  )[] = props.camps
    .filter((c) => !c.IsOld)
    .map((c) => ({ title: Camps.label(c), value: c.Id }));
  const oldCamps = props.camps
    .filter((c) => c.IsOld)
    .map((c) => ({ title: Camps.label(c), value: c.Id }));
  return newCamps.concat(
    [{ type: "divider" }, { type: "subheader", title: "Séjours terminés" }],
    oldCamps,
  );
});
</script>

<style scoped></style>
