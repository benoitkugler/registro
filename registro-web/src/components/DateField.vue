<template>
  <v-text-field
    v-if="props.readonly"
    class="center-text"
    :variant="props.variant || 'outlined'"
    density="compact"
    :hide-details="props.hideDetails"
    :label="props.label"
    :model-value="Formatters.date(modelValue, true, false)"
    readonly
  >
  </v-text-field>
  <v-menu v-else :close-on-content-click="false" v-model="showMenu">
    <template #activator="{ props: innerProps }">
      <v-text-field
        class="center-text"
        v-bind="innerProps"
        :variant="props.variant || 'outlined'"
        density="compact"
        :hide-details="props.hideDetails"
        :label="props.label"
        :model-value="Formatters.date(modelValue, true, false)"
        :rules="
          props.required ? [FormRules.requiredDate('Date requise')] : undefined
        "
      >
      </v-text-field>
    </template>

    <v-date-picker
      :model-value="
        isDateZero(modelValue) ? null : adapter.parseISO(modelValue)
      "
      @update:model-value="(d) => onUpdate(d as Date)"
    ></v-date-picker>
  </v-menu>
</template>

<script setup lang="ts">
import type { Date_ } from "@/clients/backoffice/logic/api";
import { ref } from "vue";
import { useDate } from "vuetify";
import { isDateZero, newDate_ } from "./date";
import type { VTextField } from "vuetify/components";
import { Formatters, FormRules } from "@/utils";
const props = defineProps<{
  label: string;
  hideDetails?: boolean;
  readonly?: boolean;
  variant?: VTextField["variant"];
  required?: boolean;
}>();

const modelValue = defineModel<Date_>({ required: true });

const adapter = useDate();

function onUpdate(d: Date) {
  showMenu.value = false;
  modelValue.value = newDate_(d);
}

const showMenu = ref(false);
</script>

<style scoped>
.center-text :deep(input) {
  text-align: center;
}
</style>
