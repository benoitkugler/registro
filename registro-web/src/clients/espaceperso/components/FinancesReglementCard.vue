<template>
  <v-card title="Régler la pension" class="mt-2">
    <v-card-text> TODO </v-card-text>
  </v-card>
</template>

<script setup lang="ts">
import { type DossierExt } from "../logic/api";
const props = defineProps<{
  token: string;
  dossier: DossierExt;
}>();

const emit = defineEmits<{
  (e: "refresh"): void;
}>();
</script>
