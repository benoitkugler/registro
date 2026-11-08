<template>
  <v-dialog
    :model-value="props.modelValue"
    @update:model-value="emit('update:modelValue', $event)"
    max-width="600px"
  >
    <v-card title="Téléverser le justificatif">
      <v-card-text>
        <v-file-input
          :multiple="false"
          v-model="file"
          label="Fichier"
          accept=".png,.jpeg,.jpg,.pdf"
          show-size
          :rules="[validateSize]"
        ></v-file-input>
      </v-card-text>
      <v-card-actions>
        <v-spacer></v-spacer>
        <v-btn
          color="success"
          :disabled="!file?.size || file.size >= MaxUploadSize"
          @click="uploadFile"
          >Téléverser</v-btn
        >
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script lang="ts" setup>
import { ref } from "vue";

const props = defineProps<{
  modelValue: boolean;
}>();

const emit = defineEmits<{
  (e: "update:modelValue", value: boolean): void;
  (e: "upload", file: File): void;
}>();

const MaxUploadSize = 10 * 1024 * 1024;
const file = ref<File | null>(null);

function uploadFile() {
  if (!file.value?.size || file.value.size >= MaxUploadSize) return;
  emit("upload", file.value);
  emit("update:modelValue", false);
  file.value = null;
}

function validateSize(fs: File | null) {
  if (!fs) return true;
  return (
    fs.size < MaxUploadSize ||
    `La taille d'un fichier est limitée à ${MaxUploadSize / (1024 * 1024)}M.`
  );
}
</script>
