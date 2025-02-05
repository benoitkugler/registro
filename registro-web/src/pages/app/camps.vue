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

    <v-dialog v-model="showCreateMany" max-width="600">
      <v-card
        title="Créer plusieurs camps"
        subtitle="Les camps seront liés au même taux de conversion"
      >
        <v-card-text>
          <v-row>
            <v-col>
              <IntField
                label="Nombre de camps"
                v-model="createCount"
              ></IntField>
            </v-col>
          </v-row>
          <v-divider thickness="2" class="my-2"></v-divider>
          <TauxSelect v-model="createSelectedTaux"></TauxSelect>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn
            color="success"
            @click="createMany"
            :disabled="!areCreateFieldsValid"
            >Créer</v-btn
          >
        </v-card-actions>
      </v-card>
    </v-dialog>

    <v-responsive class="align-center fill-height mx-auto">
      <v-card title="Camps" :subtitle="camps.length" class="ma-1">
        <template v-slot:append>
          <v-menu>
            <template v-slot:activator="{ props }">
              <v-btn v-bind="props" color="success" :disabled="isLoading">
                <template v-slot:prepend>
                  <v-icon>mdi-plus</v-icon>
                </template>
                Créer un camp...
              </v-btn>
            </template>
            <v-list density="compact">
              <v-list-item @click="create" title="Créer un camp"></v-list-item>
              <v-list-item
                title="Créer plusieurs camps..."
                subtitle="Permet de choisir les taux de conversion"
                @click="showCreateMany = true"
              >
              </v-list-item>
            </v-list>
          </v-menu>
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
import { controller } from "@/logic/app/logic";
import type { Camp, CampHeader, IdCamp, Int, Taux } from "@/logic/app/api";

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

const showCreateMany = ref(false);
const createCount = ref(5 as Int);
const createSelectedTaux = ref<Taux>({
  Id: 1 as Int, // defaut taux
  Label: "",
  Euros: 0 as Int,
  FrancsSuisse: 0 as Int,
});
const areCreateFieldsValid = computed(
  () => createSelectedTaux.value.Id > 0 || createSelectedTaux.value.Label != ""
);

async function createMany() {
  showCreateMany.value = false;
  isLoading.value = true;
  const res = await controller.CampsCreateMany({
    Count: createCount.value,
    Taux: createSelectedTaux.value,
  });
  isLoading.value = false;
  if (res === undefined) return;
  controller.showMessage("Camps créés avec succès.");
  res?.forEach((c) => campsData.set(c.Camp.Id, c));
}
</script>
