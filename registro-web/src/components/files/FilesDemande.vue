<template>
  <FilesRow
    :title="title"
    :subtitle="props.demande.Description"
    :max-docs="props.demande.MaxDocs"
    :files="props.files"
    :in-upload="props.inUpload"
    @upload="(v) => emit('upload', v)"
    @delete="(v) => emit('delete', v)"
  >
    <template #prepend v-if="props.optionnelle !== undefined">
      <div style="width: 200" class="my-2 mr-2">
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
  optionnelle?: boolean;
}>();

const emit = defineEmits<{
  (e: "upload", file: File): void;
  (e: "delete", file: PublicFile): void;
}>();

const title = computed(() =>
  props.demande.Categorie == Categorie.NoBuiltin
    ? "Document à fournir"
    : CategorieLabels[props.demande.Categorie]
);
</script>

<style scoped></style>
