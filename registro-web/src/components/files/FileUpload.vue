<template>
  <v-card title="Téléverser un document">
    <v-card-text>
      <v-file-input
        density="comfortable"
        accept=".jpg,.jpeg,.png,.pdf"
        :multiple="false"
        v-model="toUpload"
        :rules="[rule]"
        show-size
      ></v-file-input>
    </v-card-text>
    <v-card-actions>
      <v-spacer></v-spacer>
      <v-btn
        :disabled="toUpload == null || toUpload.size > maxSize"
        @click="emit('upload', toUpload!)"
        >Téléverser</v-btn
      >
    </v-card-actions>
  </v-card>
</template>

<script setup lang="ts">
import { ref } from "vue";

const props = defineProps<{}>();

const emit = defineEmits<{
  (e: "upload", file: File): void;
}>();

const toUpload = ref<File | null>(null);
const maxSize = 5_000_000; // in bytes
function rule(list: File[] | null) {
  if (!list || !list.length) return true;
  return (
    list.every((file) => file.size <= maxSize) ||
    `Taille maximale : ${maxSize / 1_000_000}MB`
  );
}
</script>

<style scoped></style>
