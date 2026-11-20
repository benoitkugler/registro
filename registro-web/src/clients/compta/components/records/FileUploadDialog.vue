<template>
  <v-dialog
    :model-value="props.modelValue"
    @update:model-value="emit('update:modelValue', $event)"
    max-width="600px"
  >
    <v-card
      title="Téléverser le justificatif"
      subtitle="Plusieurs images seront fusionnées en seul fichier."
    >
      <v-card-text>
        <v-file-input
          multiple
          v-model="files"
          label="Fichier"
          accept=".png,.jpeg,.jpg,.pdf"
          show-size
          :rules="[validateSize, validateFormats]"
        ></v-file-input>
      </v-card-text>
      <v-card-actions>
        <v-spacer></v-spacer>
        <v-btn
          color="success"
          :disabled="
            !files.length || totalSize >= MaxUploadSize || !formatsValid
          "
          @click="uploadFile"
          >Téléverser</v-btn
        >
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script lang="ts" setup>
import { computed, ref } from "vue";

const props = defineProps<{
  modelValue: boolean;
}>();

const emit = defineEmits<{
  (e: "update:modelValue", value: boolean): void;
  (e: "upload", files: File[]): void;
}>();

const MaxUploadSize = 10 * 1024 * 1024;
const files = ref<File[]>([]);
const totalSize = computed(() =>
  files.value.reduce((sum, file) => sum + file.size, 0),
);
const formatsValid = computed(
  () => files.value.length <= 1 || areImageFiles(files.value),
);

function uploadFile() {
  if (
    !files.value.length ||
    totalSize.value >= MaxUploadSize ||
    !formatsValid.value
  )
    return;
  emit("upload", files.value);
  emit("update:modelValue", false);
  files.value = [];
}

function validateSize(selectedFiles: File[] | null) {
  if (!selectedFiles?.length) return true;
  const size = selectedFiles.reduce((sum, file) => sum + file.size, 0);
  return (
    size < MaxUploadSize ||
    `La taille totale des fichiers est limitée à ${MaxUploadSize / (1024 * 1024)}M.`
  );
}

function areImageFiles(selectedFiles: File[]) {
  return selectedFiles.every((file) => /\.(png|jpe?g)$/i.test(file.name));
}

function validateFormats(selectedFiles: File[] | null) {
  if (!selectedFiles || selectedFiles.length <= 1) return true;
  return (
    areImageFiles(selectedFiles) ||
    "Plusieurs fichiers peuvent être téléversés uniquement au format PNG ou JPEG."
  );
}
</script>
