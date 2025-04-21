<template>
  <v-file-input
    :label="props.label"
    density="comfortable"
    accept=".jpg,.jpeg,.png,.pdf"
    :multiple="false"
    v-model="inner"
    @update:model-value="onUpdate"
    :rules="[rule]"
    show-size
  ></v-file-input>
</template>

<script setup lang="ts">
import { ref } from "vue";

const props = defineProps<{
  label?: string;
}>();

const emit = defineEmits<{
  (e: "update", file: File | null): void;
}>();

const inner = ref<File | null>(null);

const maxSize = 5_000_000; // in bytes
function rule(list: File[] | null) {
  if (!list || !list.length) return true;
  return (
    list.every((file) => file.size <= maxSize) ||
    `Taille maximale : ${maxSize / 1_000_000}MB`
  );
}

function onUpdate() {
  // do not sync if the file is invalid
  if (inner.value && inner.value.size > maxSize) {
    return;
  }
  emit("update", inner.value);
}
</script>

<style scoped></style>
