<template>
  <v-skeleton-loader
    class="mx-auto"
    width="400px"
    type="card"
    v-if="data == null"
  ></v-skeleton-loader>
  <v-card
    v-else-if="data.length"
    title="Albums photos"
    subtitle="Retrouvez les souvenirs pris dans nos séjours."
    min-width="300"
  >
    <v-card-text>
      <v-list>
        <v-list-item v-for="camp in data" :title="camp.Label">
          <template #append>
            <v-btn
              target="_blank"
              :href="camp.URL"
              prepend-icon="mdi-open-in-new"
              >Aller</v-btn
            >
          </template>
        </v-list-item>
      </v-list>
    </v-card-text>
  </v-card>
  <v-alert v-else class="mx-auto" type="info">
    Il n'y a encore rien à voir. Repassez donc dans quelques jours !
  </v-alert>
</template>

<script lang="ts" setup>
import { onMounted, ref } from "vue";
import { controller } from "../logic/logic";
import type { PhotoAlbum } from "../logic/api";

const props = defineProps<{
  token: string;
}>();

onMounted(fetchData);

const data = ref<PhotoAlbum[] | null>(null);
async function fetchData() {
  const res = await controller.LoadPhotos({ token: props.token });
  if (res === undefined) return;
  data.value = res || [];
}
</script>
