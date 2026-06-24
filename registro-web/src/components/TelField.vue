<template>
  <v-text-field
    :label="props.label"
    variant="outlined"
    density="compact"
    v-model="modelValue"
    @update:focused="onFocus"
    :hide-details="props.hideDetails"
    :disabled="props.disabled"
    :hint="props.hint"
    :persistent-hint="props.persistentHint"
    :rules="props.rule ? [props.rule] : undefined"
    type="tel"
  >
    <template #append-inner v-if="flag">
      {{ flag }}
    </template>
  </v-text-field>
</template>

<script setup lang="ts">
import type { Tel } from "@/clients/directeurs/logic/api";
import { Phones } from "@/phones";
import type { FormRules } from "@/utils";
import { computed, ref } from "vue";
const props = defineProps<{
  label: string;
  hideDetails?: boolean;
  disabled?: boolean;
  hint?: string;
  persistentHint?: boolean;
  rule?: FormRules.TelRule;
}>();

const modelValue = defineModel<Tel>({ required: true });

const flag = computed(() => Phones.parseFlag(modelValue.value));

function onFocus(b: boolean) {
  if (b) return;
  // unfocused : apply format
  modelValue.value = Phones.format(modelValue.value);
}
</script>

<style scoped></style>
