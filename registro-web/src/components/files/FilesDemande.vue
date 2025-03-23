<template>
  <v-dialog v-model="showUpload" max-width="600px">
    <FileUpload @upload=""></FileUpload>
  </v-dialog>

  <v-list-item :title="title" :subtitle="props.demande.Description">
    <template #append>
      <v-row>
        <v-col> {{ props.files }}</v-col>
        <v-col>
          <v-btn icon size="x-small" @click="showUpload = true">
            <v-icon color="green">mdi-upload</v-icon>
          </v-btn>
        </v-col>
      </v-row>
    </template>
  </v-list-item>
</template>

<script setup lang="ts">
import {
  Categorie,
  CategorieLabels,
  type Demande,
  type FilePublic,
} from "@/clients/equipier/logic/api";
import { computed, ref } from "vue";

const props = defineProps<{
  demande: Demande;
  files: FilePublic[];
  optionnelle?: boolean;
}>();

const title = computed(() =>
  props.demande.Categorie == Categorie.NoBuiltin
    ? "Document à fournir"
    : CategorieLabels[props.demande.Categorie]
);

const showUpload = ref(false);
</script>

<style scoped></style>
