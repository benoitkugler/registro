<template>
  <v-select
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
    <template #prepend>
      <v-tooltip
        :text="
          showTerminatedCamps
            ? 'Masquer les séjours terminés'
            : 'Afficher les séjours terminés'
        "
      >
        <template #activator="{ props: tooltipProps }">
          <v-btn
            size="small"
            :variant="showTerminatedCamps ? 'tonal' : 'flat'"
            v-bind="tooltipProps"
            icon="mdi-history"
            @click="showTerminatedCamps = !showTerminatedCamps"
          ></v-btn>
        </template>
      </v-tooltip>
    </template>
  </v-select>
</template>

<script setup lang="ts">
import type { CampItem, IdCamp } from "@/clients/backoffice/logic/api";
import { nullableToZeroable, zeroableToNullable } from "@/utils";
import { computed, ref, watch } from "vue";
const props = defineProps<{
  label: string;
  camps: CampItem[];
  readonly?: boolean;
  defaultStyle?: boolean;
}>();

const modelValue = defineModel<IdCamp>({ required: true });

const showTerminatedCamps = ref(false);
const campItems = computed(() =>
  props.camps
    .filter((c) => (showTerminatedCamps.value ? true : !c.IsOld))
    .map((c) => ({ title: c.Label, value: c.Id }))
);
</script>

<style scoped></style>
