<template>
  <FilesRow
    :title="title"
    :subtitle="props.demande.Description"
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
        <v-tooltip v-if="props.demande.JoursValide > 0">
          <template #activator="{ props: tooltipProps }">
            <v-icon v-bind="tooltipProps" class="mx-2"
              >mdi-clock-outline</v-icon
            >
          </template>
          Ce document sera supprimé automatiquement
          <b>{{ props.demande.JoursValide }} jours</b> après son ajout.
        </v-tooltip>
      </div>
      <template v-else>
        <slot name="prepend"></slot>
      </template>
    </template>
  </FilesRow>
</template>

<script setup lang="ts">
import {
  Categorie,
  CategorieLabels,
  type Demande,
  type PublicFile,
} from "@/clients/equipier/logic/api";
import { computed } from "vue";

const props = defineProps<{
  demande: Demande;
  files: PublicFile[];
  inUpload: boolean;
  optionnelle: boolean | null;
  showUploadText?: boolean;
  title?: string;
}>();

const emit = defineEmits<{
  (e: "upload", file: File): void;
  (e: "delete", file: PublicFile): void;
}>();

const title = computed(() =>
  props.title
    ? props.title
    : props.demande.Categorie == Categorie.NoBuiltin
    ? "Document à fournir"
    : CategorieLabels[props.demande.Categorie]
);
</script>

<style scoped></style>
