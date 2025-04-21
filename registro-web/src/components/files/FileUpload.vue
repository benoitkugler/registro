<template>
  <v-card title="Téléverser un document">
    <v-card-text>
      <FileInput @update="(f) => (toUpload = f)"></FileInput>
    </v-card-text>
    <v-card-actions>
      <v-spacer></v-spacer>
      <v-btn :disabled="toUpload == null" @click="emit('upload', toUpload!)"
        >Téléverser</v-btn
      >
    </v-card-actions>
  </v-card>
</template>

<script setup lang="ts">
import { ref } from "vue";
import FileInput from "./FileInput.vue";

const props = defineProps<{}>();

const emit = defineEmits<{
  (e: "upload", file: File): void;
}>();

const toUpload = ref<File | null>(null);
</script>

<style scoped></style>
