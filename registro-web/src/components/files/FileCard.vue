<template>
  <v-menu>
    <template #activator="{ props: menuProps }">
      <v-card
        class="my-2 border"
        :color="Formatters.reviewStatus(props.file.ReviewStatus).color"
        v-bind="menuProps"
      >
        <v-img
          cover
          :src="endpoints.LoadMiniature(props.file.Key)"
          width="56px"
        >
        </v-img>
      </v-card>
    </template>
    <v-card>
      <v-card-text class="px-0">
        <v-list-item
          min-width="400"
          :title="props.file.NomClient"
          :subtitle="`${Formatters.size(
            props.file.Taille,
          )} - Ajouté ${Formatters.date(props.file.Uploaded)}`"
        >
          <template #append>
            <v-chip
              class="ml-2"
              :color="Formatters.reviewStatus(file.ReviewStatus).color"
              :prepend-icon="Formatters.reviewStatus(file.ReviewStatus).icon"
            >
              {{ FileReviewStatusLabels[file.ReviewStatus] }}
            </v-chip>
          </template>
        </v-list-item>
      </v-card-text>
      <v-card-actions>
        <v-btn @click="emit('delete')" color="red" prepend-icon="mdi-delete">
          Supprimer
        </v-btn>
        <v-spacer></v-spacer>
        <v-btn :href="endpoints.LoadDocument(props.file.Key)">
          Télécharger
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-menu>
</template>

<script setup lang="ts">
import { endpoints, Formatters } from "@/utils";
import type { PublicFile } from "@/clients/equipier/logic/api";
import { FileReviewStatusLabels } from "@/clients/backoffice/logic/api";

const props = defineProps<{
  file: PublicFile;
  readonly?: boolean /** remove menu; download on click */;
}>();

const emit = defineEmits<{
  (e: "delete"): void;
}>();
</script>

<style scoped></style>
