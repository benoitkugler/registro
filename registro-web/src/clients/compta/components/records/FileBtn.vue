<template>
  <div>
    <!-- Upload file dialog -->
    <v-dialog v-model="showUpload" max-width="600px">
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
            @click="uploadJustificatif"
            >Téléverser</v-btn
          >
        </v-card-actions>
      </v-card>
    </v-dialog>

    <v-progress-circular
      indeterminate
      :size="props.small ? '20px' : '40px'"
      width="6"
      v-if="false"
    ></v-progress-circular>

    <v-menu v-else-if="props.record.IdFile.Valid">
      <template v-slot:activator="{ props: innerProps }">
        <v-btn
          :size="props.small ? 'x-small' : 'small'"
          icon
          v-bind="innerProps"
        >
          <v-icon color="green">mdi-file</v-icon>
        </v-btn>
      </template>
      <FilePreview
        :miniatureURL="miniatureURL"
        :downloadURL="downloadURL"
        :readonly="props.readonly"
        @upload="showUpload = true"
        @delete="emit('delete')"
      ></FilePreview>
    </v-menu>
    <v-btn
      v-else
      :size="props.small ? 'x-small' : 'small'"
      icon
      @click="showUpload = true"
      :disabled="props.readonly"
    >
      <v-icon color="orange">mdi-file</v-icon>
    </v-btn>
  </div>
</template>

<script lang="ts" setup>
import { computed, ref } from "vue";
import type { ComptaSubrecord } from "../../logic/api.ts";
import { controller } from "../../logic/logic.ts";
import FilePreview from "./FilePreview.vue";

const props = defineProps<{
  record: ComptaSubrecord;
  readonly: boolean;
  small?: boolean;
}>();

const emit = defineEmits<{
  (e: "upload", file: File): void;
  (e: "delete"): void;
}>();

const miniatureURL = computed(() =>
  controller.GetMiniature(props.record.Id, controller.authToken),
);
const downloadURL = computed(() =>
  controller.DownloadJustificatif(props.record.Id, controller.authToken),
);

const MaxUploadSize = 10 * 1024 * 1024;
const file = ref<File | null>(null);

const showUpload = ref(false);
function uploadJustificatif() {
  if (!file.value?.size) return;
  emit("upload", file.value);
  showUpload.value = false;
  file.value = null;
}

function validateSize(fs: File[]) {
  if (!fs.length) return true;
  return (
    fs[0].size < MaxUploadSize ||
    `La taille d'un fichier est limitée à ${MaxUploadSize / (1024 * 1024)}M.`
  );
}
</script>
