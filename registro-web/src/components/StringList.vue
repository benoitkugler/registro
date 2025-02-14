<template>
  <v-combobox
    variant="outlined"
    density="compact"
    :label="props.label"
    chips
    multiple
    closable-chips
    v-model="modelValue"
    @update:model-value="format"
    :error-messages="props.error ? [props.error] : null"
  ></v-combobox>
</template>

<script setup lang="ts">
const props = defineProps<{
  label: string;
  formatter?: (s: string) => string;
  hideDetails?: boolean;
  readonly?: boolean;
  error?: string;
}>();

const modelValue = defineModel<string[] | null>({ required: true });

function format(v: string[] | null) {
  v = v || [];
  const fmt = props.formatter ?? ((s: string) => s);
  modelValue.value = v.map(fmt);
}
</script>

<style scoped></style>
