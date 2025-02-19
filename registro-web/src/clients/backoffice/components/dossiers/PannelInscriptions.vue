<template>
  <v-card title="Inscriptions" subtitle="En attente de validation" class="ma-2">
    <template v-slot:append>
      <v-text-field
        width="300"
        append-inner-icon="mdi-magnify"
        label="Rechercher"
        hide-details
      ></v-text-field>
    </template>
    <v-card-text>
      <v-skeleton-loader v-if="isLoading"></v-skeleton-loader>
    </v-card-text>
  </v-card>
</template>

<script setup lang="ts">
import { ref } from "vue";
const props = defineProps<{
  title: string;
}>();

const isLoading = ref(false);
</script>
