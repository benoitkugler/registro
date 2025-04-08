<template>
  <v-list-item
    :title="title"
    :subtitle="props.demande.Description"
    rounded
    class="bg-grey-lighten-4 my-1"
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
    <template #append>
      <v-row>
        <v-col align-self="center" v-for="file in props.files">
          <FileCard :file="file" @delete="toDelete = file"></FileCard>
        </v-col>
        <v-col align-self="center" cols="auto">
          <v-btn
            icon
            size="x-small"
            @click="showUpload = true"
            :disabled="
              props.inUpload ||
              (props.demande.MaxDocs != 0 &&
                props.files.length >= props.demande.MaxDocs)
            "
          >
            <v-icon
              color="green"
              :icon="props.inUpload ? 'mdi-loading' : 'mdi-upload'"
              :class="props.inUpload ? 'mdi-spin' : undefined"
            >
            </v-icon>
          </v-btn>
        </v-col>
      </v-row>
    </template>
  </v-list-item>

  <!-- upload -->
  <v-dialog v-model="showUpload" max-width="600px">
    <FileUpload
      @upload="
        (file) => {
          emit('upload', file);
          showUpload = false;
        }
      "
    ></FileUpload>
  </v-dialog>

  <!-- confirme delete -->
  <v-dialog
    v-if="toDelete != null"
    :model-value="toDelete != null"
    @update:model-value="toDelete = null"
    max-width="600px"
  >
    <v-card title="Confirmation">
      <v-card-text>
        Le document va être supprimé. <br /><br />
        Attention, cette opération est irréversible.
      </v-card-text>
      <v-card-actions>
        <v-spacer></v-spacer>
        <v-btn
          color="red"
          @click="
            emit('delete', toDelete);
            toDelete = null;
          "
          >Supprimer</v-btn
        >
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
import {
  Categorie,
  CategorieLabels,
  type Demande,
  type PublicFile,
} from "@/clients/equipier/logic/api";
import { computed, ref } from "vue";

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

const showUpload = ref(false);

const toDelete = ref<PublicFile | null>(null);
</script>

<style scoped></style>
