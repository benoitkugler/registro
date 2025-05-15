<template>
  <v-menu>
    <template #activator="{ props: menuProps }">
      <v-card class="my-2" v-bind="menuProps">
        <v-img cover :src="miniatureURL" width="100"> </v-img>
      </v-card>
    </template>
    <v-card>
      <v-card-text class="px-0">
        <v-list-item
          min-width="400"
          :title="props.file.NomClient"
          :subtitle="`${props.file.Taille / 1000} KB`"
        >
          <template #append>
            <div class="text-right">
              téléversé le <br />
              {{ Formatters.time(props.file.Uploaded) }}
            </div>
          </template>
        </v-list-item>
      </v-card-text>
      <v-card-actions>
        <v-btn @click="emit('delete')" color="red"> Supprimer </v-btn>
        <v-spacer></v-spacer>
        <v-btn :href="contentURL"> Télécharger </v-btn>
      </v-card-actions>
    </v-card>
  </v-menu>
</template>

<script setup lang="ts">
import type { PublicFile } from "@/clients/equipier/logic/api";
import { baseUrl, Formatters } from "@/utils";
import { computed } from "vue";

const props = defineProps<{
  file: PublicFile;
}>();

const emit = defineEmits<{
  (e: "delete"): void;
}>();

// hardcoded, global files endpoint
const miniatureURL = computed(
  () => `${baseUrl()}/api/v1/documents/miniature?key=${props.file.Key}`
);
const contentURL = computed(
  () => `${baseUrl()}/api/v1/documents?key=${props.file.Key}`
);
</script>

<style scoped></style>
