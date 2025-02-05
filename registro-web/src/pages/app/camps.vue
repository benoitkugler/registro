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

    <v-dialog
      :model-value="toEditTaux != null"
      @update:model-value="toEditTaux = null"
      max-width="600"
    >
      <v-card
        v-if="toEditTaux != null"
        title="Modifier le taux"
        subtitle="La modification n'est possible que si aucun participant n'est enregistré."
      >
        <v-card-text>
          <TauxSelect v-model="toEditTaux.Taux"></TauxSelect>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn :disabled="toEditTaux.Stats.Inscriptions > 0" @click="setTaux"
            >Modifier</v-btn
          >
        </v-card-actions>
      </v-card>
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
      <v-card title="Camps" :subtitle="campsData.size" class="ma-1">
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
            @edit-taux="toEditTaux = camp"
          ></CampHeaderRow>

          <v-pagination
            :length="pagesCount"
            v-model="currentPage"
          ></v-pagination>
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
const camps = computed(() => {
  const out = Array.from(campsData.values());
  // most recent first
  out.sort(
    (a, b) =>
      new Date(b.Camp.DateDebut).valueOf() -
      new Date(a.Camp.DateDebut).valueOf()
  );
  return out.slice(
    (currentPage.value - 1) * pageSize,
    currentPage.value * pageSize
  );
});

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

const pageSize = 16;
const pagesCount = computed(() => Math.ceil(campsData.size / pageSize));
const currentPage = ref(1); // 1-based

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

const toEditTaux = ref<CampHeader | null>(null);
async function setTaux() {
  const val = toEditTaux.value;
  if (val == null) return;
  toEditTaux.value = null;
  isLoading.value = true;
  const res = await controller.CampsSetTaux({
    IdCamp: val.Camp.Id,
    Taux: val.Taux,
  });
  isLoading.value = false;
  if (res === undefined) return;
  controller.showMessage("Taux modifié avec succès.");
  campsData.set(res.Camp.Id, res);
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
