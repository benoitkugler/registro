<template>
  <NavBar title="Gestion des camps">
    <div>Boutons a venir</div>
  </NavBar>

  <v-container class="fill-height" fluid>
    <v-responsive class="align-center fill-height mx-auto">
      <v-card title="Camps" :subtitle="camps.length" class="ma-1">
        <template v-slot:append>
          <v-btn color="success" @click="create" :disabled="isLoading">
            <template v-slot:prepend>
              <v-icon>mdi-plus</v-icon>
            </template>
            Créer un camp</v-btn
          >
        </template>
        <v-card-text class="mt-4">
          <v-skeleton-loader v-if="isLoading"></v-skeleton-loader>
          <i v-else-if="camps.length == 0">Aucun camp.</i>

          <CampHeaderRow
            v-for="(camp, index) in camps"
            :key="index"
            :camp="camp"
          ></CampHeaderRow>

          <v-pagination :length="pagesCount"></v-pagination>
        </v-card-text>
      </v-card>
    </v-responsive>
  </v-container>
</template>

<script lang="ts" setup>
import { ref, computed, onMounted } from "vue";
import { controller } from "@/logic/app/logic";
import type { CampHeader } from "@/logic/app/api";

const camps = ref<CampHeader[]>([]);
const isLoading = ref(false);

onMounted(fetchCamps);

async function fetchCamps() {
  isLoading.value = true;
  const res = await controller.CampsGet();
  isLoading.value = false;
  if (res === undefined) return;
  camps.value = res || [];
}

async function create() {
  isLoading.value = true;
  const res = await controller.CampsCreate();
  isLoading.value = false;
  if (res === undefined) return;
  camps.value.push(res);
}

const pageSize = 12;
const pagesCount = computed(() => Math.ceil(camps.value.length / pageSize));
</script>
