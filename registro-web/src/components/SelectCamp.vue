<template>
  <v-select
    density="compact"
    variant="outlined"
    :label="props.label"
    :readonly="props.readonly"
    :items="campItems"
    clearable
    v-model="modelValue"
  >
    <template #prepend>
      <v-tooltip
        :text="
          showTerminatedCamps
            ? 'Masquer les camps terminés'
            : 'Afficher les camps terminés'
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
import { computed, ref, watch } from "vue";
const props = defineProps<{
  label: string;
  camps: CampItem[];
  readonly?: boolean;
}>();

const modelValue = defineModel<IdCamp | null>({ required: true });

const showTerminatedCamps = ref(false);
const campItems = computed(() =>
  props.camps
    .filter((c) => (showTerminatedCamps.value ? true : !c.IsOld))
    .map((c) => ({ title: c.Label, value: c.Id }))
);
</script>

<style scoped></style>
