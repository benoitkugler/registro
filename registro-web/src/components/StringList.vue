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
    :rules="props.rule ? [rule] : undefined"
    hint="Valider avec entrée ou en cliquant à côté du champ."
  ></v-combobox>
</template>

<script setup lang="ts">
const props = defineProps<{
  label: string;
  formatter?: (s: string) => string;
  hideDetails?: boolean;
  readonly?: boolean;
  rule?: (l: string[]) => true | string;
}>();

const modelValue = defineModel<string[] | null>({ required: true });

function format(v: string[] | null) {
  v = v || [];
  const fmt = props.formatter ?? ((s: string) => s);
  modelValue.value = v.map(fmt);
}

function rule(v: string[]) {
  return props.rule!(v);
}
</script>

<style scoped></style>
