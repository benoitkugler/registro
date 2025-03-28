<template>
  <v-card title="Séjours" :subtitle="campsData.size">
    <template #append>
      <v-row>
        <v-col align-self="center" style="width: 240px">
          <v-text-field
            label="Rechercher"
            variant="filled"
            density="comfortable"
            hide-details
            prepend-inner-icon="mdi-magnify"
            clearable
            v-model="filter.pattern"
            @click:clear="
              filter.pattern = '';
              ensurePageValid();
            "
            @update:model-value="ensurePageValid()"
          >
          </v-text-field>
        </v-col>
        <v-divider vertical thickness="1"></v-divider>
        <v-col align-self="center" cols="auto">
          <v-menu>
            <template #activator="{ props: innerProps }">
              <v-btn v-bind="innerProps" color="success" :disabled="isLoading">
                <template #prepend>
                  <v-icon>mdi-plus</v-icon>
                </template>
                Créer un séjour
              </v-btn>
            </template>
            <v-list density="compact">
              <v-list-item
                @click="create"
                title="Créer un séjour"
              ></v-list-item>
              <v-list-item
                title="Créer plusieurs séjours..."
                subtitle="Permet de choisir les taux de conversion"
                @click="showCreateMany = true"
              >
              </v-list-item>
            </v-list>
          </v-menu>
        </v-col>
        <v-col align-self="center" cols="auto">
          <v-menu>
            <template #activator="{ props: innerProps }">
              <v-btn
                v-bind="innerProps"
                flat
                icon="mdi-dots-vertical"
                size="small"
              >
              </v-btn>
            </template>
            <v-list density="compact">
              <v-list-item
                title="Masquer les camps fermés"
                subtitle="N'afficher que les camps ouverts aux inscriptions"
                @click="filter.openOnly = !filter.openOnly"
              >
                <template #prepend>
                  <v-checkbox
                    class="mr-2"
                    v-model="filter.openOnly"
                    density="compact"
                    hide-details
                  ></v-checkbox>
                </template>
              </v-list-item>
              <v-divider thickness="1"></v-divider>
              <v-list-item
                @click="startOpenInsc"
                title="Ouvrir les inscriptions"
                prepend-icon="mdi-lock-open"
              ></v-list-item>
            </v-list>
          </v-menu>
        </v-col>
      </v-row>
    </template>
    <v-card-text class="mt-4">
      <v-alert v-if="!isPlageTauxValid" type="warning">
        Attention, les séjours ouverts aux inscriptions ont des taux de
        conversion incompatibles !
      </v-alert>
      <v-skeleton-loader v-if="isLoading"></v-skeleton-loader>
      <i v-else-if="camps.length == 0">
        Aucun camp ne correspond aux filtres actuels.
      </i>

      <CampHeaderRow
        v-for="(camp, index) in pageList"
        :key="index"
        :camp="camp"
        @click="emit('click', camp)"
        @edit="toEdit = camp.Camp.Camp"
        @edit-taux="toEditTaux = camp"
        @delete="deleteCamp(camp)"
      ></CampHeaderRow>

      <v-pagination :length="pagesCount" v-model="currentPage"></v-pagination>
    </v-card-text>
  </v-card>

  <!-- Edition -->
  <v-dialog :model-value="toEdit != null" @update:model-value="toEdit = null">
    <CampEdit
      v-if="toEdit != null"
      :camp="toEdit"
      @save="updateCamp"
    ></CampEdit>
  </v-dialog>

  <!-- Choix du taux -->
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
        <TauxSelect v-model="toEditTaux!.Taux"></TauxSelect>
      </v-card-text>
      <v-card-actions>
        <v-spacer></v-spacer>
        <v-btn :disabled="toEditTaux.Stats.Inscriptions > 0" @click="setTaux"
          >Modifier</v-btn
        >
      </v-card-actions>
    </v-card>
  </v-dialog>

  <!-- Assitant création -->
  <v-dialog v-model="showCreateMany" max-width="600">
    <v-card
      title="Créer plusieurs séjours"
      subtitle="Les séjours seront liés au même taux de conversion"
    >
      <v-card-text>
        <v-row>
          <v-col>
            <IntField
              label="Nombre de séjours"
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

  <!-- Assitant open inscription -->
  <v-dialog v-model="showOpenInsc" max-width="600">
    <v-card
      title="Ouvrir les inscriptions"
      subtitle="Ouvre les inscriptions sur les séjours ci-dessous"
    >
      <v-card-text>
        <CampsSelector
          :all-camps="
            Array.from(campsData.values()).map((c) => Camps.toItem(c.Camp))
          "
          v-model="campToOpen"
        ></CampsSelector>
      </v-card-text>
      <v-card-actions>
        <v-spacer></v-spacer>
        <v-btn @click="openInsc" :disabled="!campToOpen.length">Ouvrir</v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script lang="ts" setup>
import { ref, computed, onMounted, reactive } from "vue";
import { controller, isCampOpen } from "@/clients/backoffice/logic/logic";
import type {
  Camp,
  CampHeader,
  IdCamp,
  IdTaux,
  Int,
  Taux,
} from "@/clients/backoffice/logic/api";
import CampEdit from "./CampEdit.vue";
import TauxSelect from "./TauxSelect.vue";
import CampHeaderRow from "./CampHeaderRow.vue";
import { Camps, normalize } from "@/utils";
import CampsSelector from "./CampsSelector.vue";

const emit = defineEmits<{
  (e: "click", camp: CampHeader): void;
}>();

onMounted(fetchCamps);

const campsData = reactive(new Map<IdCamp, CampHeader>());
const isLoading = ref(false);

const filter = reactive({ pattern: "", openOnly: false });

// with sort and filter
const camps = computed(() => {
  const pattern = normalize(filter.pattern);
  const out = Array.from(campsData.values()).filter(
    (camp) =>
      Camps.match(camp.Camp.Camp, pattern) &&
      (!filter.openOnly || isCampOpen(camp.Camp))
  );
  // most recent first
  out.sort((a, b) => {
    const da = new Date(b.Camp.Camp.DateDebut).valueOf();
    const db = new Date(a.Camp.Camp.DateDebut).valueOf();
    return da == db ? a.Camp.Camp.Id - b.Camp.Camp.Id : da - db;
  });
  return out;
});
const pageList = computed(() =>
  camps.value.slice(
    (currentPage.value - 1) * pageSize,
    currentPage.value * pageSize
  )
);

function ensurePageValid() {
  if (currentPage.value > pagesCount.value) {
    currentPage.value = pagesCount.value;
  }
}

async function fetchCamps() {
  isLoading.value = true;
  const res = await controller.CampsGet();
  isLoading.value = false;
  if (res === undefined) return;
  campsData.clear();
  res?.forEach((v) => campsData.set(v.Camp.Camp.Id, v));
}

async function create() {
  isLoading.value = true;
  const res = await controller.CampsCreate();
  isLoading.value = false;
  if (res === undefined) return;

  controller.showMessage("Camp ajouté avec succès.");
  campsData.set(res.Camp.Camp.Id, res);
  toEdit.value = res.Camp.Camp;
}

const pageSize = 16;
const pagesCount = computed(() =>
  Math.max(Math.ceil(camps.value.length / pageSize), 1)
);
const currentPage = ref(1); // 1-based

const toEdit = ref<Camp | null>(null);
async function updateCamp(camp: Camp) {
  toEdit.value = null;
  isLoading.value = true;
  const res = await controller.CampsUpdate(camp);
  isLoading.value = false;
  if (res === undefined) return;
  controller.showMessage("Séjour modifié avec succès.");
  campsData.get(res.Camp.Id)!.Camp = res;
  ensurePageValid();
}

const toEditTaux = ref<CampHeader | null>(null);
async function setTaux() {
  const val = toEditTaux.value;
  if (val == null) return;
  toEditTaux.value = null;
  isLoading.value = true;
  const res = await controller.CampsSetTaux({
    IdCamp: val.Camp.Camp.Id,
    Taux: val.Taux,
  });
  isLoading.value = false;
  if (res === undefined) return;
  controller.showMessage("Taux modifié avec succès.");
  campsData.set(res.Camp.Camp.Id, res);
}

const showCreateMany = ref(false);
const createCount = ref(5 as Int);
const createSelectedTaux = ref<Taux>({
  Id: 1 as IdTaux, // defaut taux
  Label: "",
  Euros: 0 as Int,
  FrancsSuisse: 0 as Int,
});
const areCreateFieldsValid = computed(
  () => createSelectedTaux.value.Id > 0 || createSelectedTaux.value.Label != ""
);

/** isPlageTauxValid return true if all the camp
 * open to inscriptions have the same [IdTaux]
 */
const isPlageTauxValid = computed(() => {
  const tauxCampsOpen = new Set(
    Array.from(campsData.values())
      .filter((c) => isCampOpen(c.Camp))
      .map((c) => c.Camp.Camp.IdTaux)
  );
  return tauxCampsOpen.size <= 1;
});

async function createMany() {
  showCreateMany.value = false;
  isLoading.value = true;
  const res = await controller.CampsCreateMany({
    Count: createCount.value,
    Taux: createSelectedTaux.value,
  });
  isLoading.value = false;
  if (res === undefined) return;
  controller.showMessage("Séjours créés avec succès.");
  res?.forEach((c) => campsData.set(c.Camp.Camp.Id, c));
}

async function deleteCamp(camp: CampHeader) {
  isLoading.value = true;
  const res = await controller.CampsDelete({ id: camp.Camp.Camp.Id });
  isLoading.value = false;
  if (res === undefined) return;
  controller.showMessage("Séjour supprimé avec succès.");
  campsData.delete(camp.Camp.Camp.Id);
  ensurePageValid();
}

const showOpenInsc = ref(false);
function startOpenInsc() {
  // hint : all camps not open yet
  campToOpen.value = Array.from(campsData.values())
    .filter((c) => !c.Camp.Camp.Ouvert)
    .map((c) => c.Camp.Camp.Id);
  showOpenInsc.value = true;
}
const campToOpen = ref<IdCamp[]>([]);
async function openInsc() {
  showOpenInsc.value = false;
  if (!campToOpen.value.length) return;
  const res = await controller.CampsOuvreInscriptions({
    Camps: campToOpen.value,
  });
  if (res === undefined) return;
  controller.showMessage("Séjours ouverts aux inscriptions avec succès.");
  fetchCamps();
}
</script>
