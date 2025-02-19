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
      <div class="text-center font-italic" v-else-if="!data.length">
        Il n'y a aucune inscription à valider.
      </div>
      <div v-else>
        <InscriptionRow
          class="my-1"
          v-for="(insc, i) in data"
          :key="i"
          :inscription="insc"
        ></InscriptionRow>
      </div>
    </v-card-text>
  </v-card>
</template>

<script setup lang="ts">
import { onMounted, ref } from "vue";
import { controller } from "../../logic/logic";
import type { Inscription } from "../../logic/api";
import InscriptionRow from "./InscriptionRow.vue";
const props = defineProps<{}>();

const isLoading = ref(false);

onMounted(fetchInscriptions);

const data = ref<Inscription[]>([]);

async function fetchInscriptions() {
  isLoading.value = true;
  const res = await controller.InscriptionsGet();
  isLoading.value = false;
  if (res === undefined) return;
  data.value = res || [];
}
</script>
