<template>
  <v-list-item
    :title="props.title"
    :subtitle="props.subtitle"
    rounded
    class="bg-grey-lighten-4 my-1"
  >
    <template #prepend>
      <slot name="prepend"></slot>
    </template>
    <template #append>
      <v-row>
        <v-col align-self="center" v-for="file in props.files">
          <FileCard :file="file" @delete="toDelete = file"></FileCard>
        </v-col>
        <v-col align-self="center" cols="auto">
          <v-btn
            class="my-2"
            icon
            size="x-small"
            @click="showUpload = true"
            :disabled="
              props.inUpload ||
              (props.maxDocs != 0 && props.files.length >= props.maxDocs)
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
  </v-list-item>
</template>

<script setup lang="ts">
import { type PublicFile } from "@/clients/equipier/logic/api";
import { ref } from "vue";

const props = defineProps<{
  title: string;
  subtitle?: string;
  maxDocs: number;
  files: PublicFile[];
  inUpload: boolean;
}>();

const emit = defineEmits<{
  (e: "upload", file: File): void;
  (e: "delete", file: PublicFile): void;
}>();

const showUpload = ref(false);

const toDelete = ref<PublicFile | null>(null);
</script>

<style scoped></style>
