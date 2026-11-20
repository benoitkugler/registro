<template>
  <div>
    <FileUploadDialog
      v-model="showUpload"
      @upload="emit('upload', $event)"
    ></FileUploadDialog>

    <v-progress-circular
      indeterminate
      :size="props.small ? '20px' : '40px'"
      width="6"
      v-if="props.inUpload"
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
import FileUploadDialog from "./FileUploadDialog.vue";

const props = defineProps<{
  record: ComptaSubrecord;
  inUpload: boolean;
  readonly: boolean;
  small?: boolean;
}>();

const emit = defineEmits<{
  (e: "upload", files: File[]): void;
  (e: "delete"): void;
}>();

const miniatureURL = computed(
  () =>
    controller.GetMiniature(props.record.Id, controller.authToken) +
    `&t=${Date.now()}`, // disable cache, used when replacing a file
);
const downloadURL = computed(() =>
  controller.DownloadJustificatif(props.record.Id, controller.authToken),
);

const showUpload = ref(false);
</script>
