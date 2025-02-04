<template>
  <NavBar title="Gestion des camps">
    <div>Boutons a venir</div>
  </NavBar>

  <v-container class="fill-height" fluid>
    <v-dialog :model-value="toEdit != null" @update:model-value="toEdit = null">
      <CampEdit
        v-if="toEdit != null"
        :camp="toEdit"
        @save="updateCamp"
      ></CampEdit>
    </v-dialog>

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
            @edit="toEdit = camp.Camp"
          ></CampHeaderRow>

          <v-pagination :length="pagesCount"></v-pagination>
        </v-card-text>
      </v-card>
    </v-responsive>
  </v-container>
</template>

<script lang="ts" setup>
import { ref, computed, onMounted, reactive } from "vue";
import { controller, copy } from "@/logic/app/logic";
import type { Camp, CampHeader, IdCamp } from "@/logic/app/api";

onMounted(fetchCamps);

const campsData = reactive(new Map<IdCamp, CampHeader>());
const isLoading = ref(false);

// with sort and filter
const camps = computed(() => Array.from(campsData.values()));

async function fetchCamps() {
  isLoading.value = true;
  const res = await controller.CampsGet();
  isLoading.value = false;
  if (res === undefined) return;
  campsData.clear();
  res?.forEach((v) => campsData.set(v.Camp.Id, v));
}

async function create() {
  isLoading.value = true;
  const res = await controller.CampsCreate();
  isLoading.value = false;
  if (res === undefined) return;

  controller.showMessage("Camp ajouté avec succès.");
  campsData.set(res.Camp.Id, res);
  toEdit.value = res.Camp;
}

const pageSize = 12;
const pagesCount = computed(() => Math.ceil(camps.value.length / pageSize));

const toEdit = ref<Camp | null>(null);
async function updateCamp(camp: Camp) {
  toEdit.value = null;
  isLoading.value = true;
  const res = await controller.CampsUpdate(camp);
  isLoading.value = false;
  if (res === undefined) return;
  controller.showMessage("Camp modifié avec succès.");
  campsData.get(res.Id)!.Camp = res;
}
</script>
