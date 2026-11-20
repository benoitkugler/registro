<template>
  <FilesRow
    :title="title"
    :subtitle="props.subtitle"
    :max-docs="props.demande.MaxDocs"
    :files="props.files"
    :in-upload="props.inUpload"
    :show-upload-text="props.showUploadText"
    @upload="(v) => emit('upload', v)"
    @delete="(v) => emit('delete', v)"
  >
    <template #prepend>
      <div
        style="width: 200"
        class="my-2 mr-2"
        v-if="props.optionnelle !== null"
      >
        <v-chip label :color="props.optionnelle ? undefined : 'pink'">
          {{ props.optionnelle ? "Optionnel" : "Requis" }}
        </v-chip>
      </div>
      <template v-else>
        <slot name="prepend"></slot>
      </template>
    </template>
  </FilesRow>
</template>

<script setup lang="ts">
import { type Demande, type PublicFile } from "@/clients/equipier/logic/api";
import { Formatters } from "@/utils";
import { computed } from "vue";

const props = defineProps<{
  demande: Demande;
  files: PublicFile[];
  inUpload: boolean;
  optionnelle: boolean | null;
  showUploadText?: boolean;
  subtitle?: string;
}>();

const emit = defineEmits<{
  (e: "upload", file: File): void;
  (e: "delete", file: PublicFile): void;
}>();

const title = computed(() => Formatters.demande(props.demande));
</script>

<style scoped></style>
