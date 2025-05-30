<template>
  <v-menu>
    <template #activator="{ props: menuProps }">
      <v-card class="my-2" v-bind="menuProps">
        <v-img cover :src="miniatureURL(props.file.Key)" width="100"> </v-img>
      </v-card>
    </template>
    <v-card>
      <v-card-text class="px-0">
        <v-list-item
          min-width="400"
          :title="props.file.NomClient"
          :subtitle="Formatters.size(props.file.Taille)"
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
        <v-btn :href="contentURL(props.file.Key)"> Télécharger </v-btn>
      </v-card-actions>
    </v-card>
  </v-menu>
</template>

<script setup lang="ts">
import { contentURL, miniatureURL, Formatters } from "@/utils";
import type { PublicFile } from "@/clients/equipier/logic/api";

const props = defineProps<{
  file: PublicFile;
}>();

const emit = defineEmits<{
  (e: "delete"): void;
}>();
</script>

<style scoped></style>
