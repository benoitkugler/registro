<template>
  <v-skeleton-loader v-if="data == null" type="table"></v-skeleton-loader>
  <v-card
    v-else
    title="Dons"
    :subtitle="
      filteredDons.length == data.Dons?.length
        ? `${data.Dons.length}`
        : `${filteredDons.length} visibles sur ${data.Dons?.length}`
    "
    width="100%"
  >
    <template #append>
      <v-row>
        <v-col align-self="center">
          <v-chip>Total annuel particuliers {{ totals.Particuliers }}</v-chip>
        </v-col>
        <v-col align-self="center">
          <v-chip>Total annuel organismes {{ totals.Organismes }}</v-chip>
        </v-col>
        <v-col>
          <v-text-field
            width="200px"
            label="Recherche"
            variant="outlined"
            density="compact"
            hide-details
            placeholder="Rechercher..."
            clearable
            @click:clear="filter.pattern = ''"
            v-model="filter.pattern"
          ></v-text-field>
        </v-col>
        <v-col>
          <v-select
            label="Année"
            v-model="filter.year"
            @update:model-value="(v) => emit('update:selectedYear', v)"
            variant="outlined"
            density="compact"
            :items="yearItems"
            hide-details
          ></v-select>
        </v-col>
      </v-row>
    </template>
    <v-card-text class="mt-2">
      <v-list-item v-if="!paginatedDons.length"><i>Aucun don.</i></v-list-item>
      <v-row v-for="don in paginatedDons">
        <v-col cols="auto" align-self="center">
          <v-icon
            :icon="don.Don.IdPersonne.Valid ? '' : 'mdi-account-group'"
          ></v-icon>
        </v-col>
        <v-col cols="5" align-self="center">
          <v-list-item-title>{{ don.Donateur }}</v-list-item-title>
          <v-list-item-subtitle
            >{{
              [don.Don.Affectation, don.Don.Details]
                .filter((s) => s)
                .join(" ; ")
            }}
          </v-list-item-subtitle>
        </v-col>
        <v-col align-self="center" class="text-center">
          <v-chip variant="outlined">
            {{ don.Montant }}
          </v-chip>
        </v-col>
        <v-col align-self="center" cols="1" class="text-center">
          <v-tooltip>
            <template #activator="{ props: tooltipProps }">
              <v-icon
                v-bind="tooltipProps"
                :icon="Formatters.paiementIcon(don.Don.ModePaiement)"
              >
              </v-icon>
            </template>
            {{ ModePaiementLabels[don.Don.ModePaiement] }}
          </v-tooltip>
        </v-col>
        <v-col align-self="center" class="text-center">
          {{ Formatters.date(don.Don.Date, true, false) }}
        </v-col>
        <v-col align-self="center" class="text-center" cols="auto">
          <v-chip
            :color="don.Don.Remercie ? 'green' : 'orange'"
            :prepend-icon="don.Don.Remercie ? 'mdi-check' : ''"
            elevation="1"
            @click="toogleRemercie(don.Don)"
            >{{ don.Don.Remercie ? "Remercié" : "À remercier" }}</v-chip
          >
        </v-col>
        <v-col cols="auto">
          <v-btn icon size="small" flat>
            <v-icon>mdi-dots-vertical</v-icon>
            <v-menu activator="parent">
              <v-list density="compact">
                <v-list-item
                  title="Modifier"
                  prepend-icon="mdi-pencil"
                  @click="toEdit = don.Don"
                ></v-list-item>
                <v-list-item
                  title="Supprimer"
                  prepend-icon="mdi-delete"
                  @click="toDelete = don"
                ></v-list-item>
              </v-list>
            </v-menu>
          </v-btn>
        </v-col>
      </v-row>
      <v-pagination :length="pagesCount" v-model="currentPage"></v-pagination>
    </v-card-text>

    <v-dialog
      :model-value="toEdit != null"
      @update:model-value="toEdit = null"
      max-width="600px"
    >
      <DonPannel
        v-if="toEdit"
        :don="toEdit"
        :affectation-hints="affectationsHints"
        @save="updateDon"
      ></DonPannel>
    </v-dialog>

    <v-dialog
      :model-value="toDelete != null"
      @update:model-value="toDelete = null"
      max-width="600px"
    >
      <v-card v-if="toDelete" title="Confirmation">
        <v-card-text
          >Confirmez-vous la suppression du don de {{ toDelete.Donateur }} d'un
          montant de {{ toDelete.Montant }} ? <br /><br />
          Attention, cette opération est irréversible.</v-card-text
        >
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="red" @click="deleteDon">Supprimer</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-card>
</template>

<script lang="ts" setup>
import { ref, computed, onMounted, reactive } from "vue";
import { controller } from "../logic/logic";
import {
  ModePaiementLabels,
  type Don,
  type DonExt,
  type DonsOut,
  type Int,
} from "../logic/api";
import { Formatters, normalize } from "@/utils";
import DonPannel from "./DonPannel.vue";

const emit = defineEmits<{
  (e: "update:selectedYear", year: Int): void;
}>();

onMounted(() => {
  fetchDons();
  emit("update:selectedYear", currentYear());
});

const data = ref<DonsOut | null>(null);

const affectationsHints = computed(() =>
  Array.from(new Set(data.value?.Dons?.map((d) => d.Don.Affectation)).keys())
    .filter((s) => s != "")
    .sort()
);

const filter = reactive({
  pattern: "",
  year: currentYear(),
});

// with sort and filter
const filteredDons = computed(() => {
  const pattern = normalize(filter.pattern);
  const out = (data.value?.Dons || []).filter(
    (don) =>
      new Date(don.Don.Date).getFullYear() == filter.year &&
      (pattern == "" ||
        normalize(don.Donateur).includes(pattern) ||
        normalize(don.Don.Details).includes(pattern))
  );
  // most recent first
  out.sort((a, b) => {
    const da = new Date(b.Don.Date).valueOf();
    const db = new Date(a.Don.Date).valueOf();
    return da == db ? a.Don.Id - b.Don.Id : da - db;
  });
  return out;
});
const paginatedDons = computed(() =>
  filteredDons.value.slice(
    (currentPage.value - 1) * pageSize,
    currentPage.value * pageSize
  )
);

function ensurePageValid() {
  if (currentPage.value > pagesCount.value) {
    currentPage.value = pagesCount.value;
  }
}

const pageSize = 16;
const pagesCount = computed(() =>
  Math.max(Math.ceil(filteredDons.value.length / pageSize), 1)
);
const currentPage = ref(1); // 1-based

function currentYear() {
  return new Date(Date.now()).getFullYear() as Int;
}

const yearItems = computed(() => {
  const s = new Set(Object.keys(data.value?.YearTotals || {}).map(Number));
  s.add(currentYear());
  return Array.from(s.keys()).sort();
});

const totals = computed(
  () =>
    (data.value?.YearTotals || {})[filter.year] || {
      Particuliers: "",
      Organismes: "",
    }
);

async function fetchDons() {
  const res = await controller.LoadDons();
  if (res === undefined) return;
  data.value = res || [];
  ensurePageValid();
}

// async function create() {
//   isLoading.value = true;
//   const res = await controller.CampsCreate();
//   isLoading.value = false;
//   if (res === undefined) return;

//   controller.showMessage("Camp ajouté avec succès.");
//   campsData.set(res.Camp.Camp.Id, res);
//   toEdit.value = res;
// }

const toEdit = ref<Don | null>(null);
async function updateDon(don: Don) {
  toEdit.value = null;
  const res = await controller.UpdateDon(don);
  if (res === undefined) return;
  controller.showMessage("Don modifié avec succès.");
  fetchDons();
}

const toDelete = ref<DonExt | null>(null);
async function deleteDon() {
  if (!toDelete.value) return;
  const don = toDelete.value;
  toDelete.value = null;
  const res = await controller.DeleteDon({ id: don.Don.Id });
  if (res === undefined) return;
  controller.showMessage("Don supprimé avec succès.");
  fetchDons();
}

async function toogleRemercie(don: Don) {
  don.Remercie = !don.Remercie;
  const res = await controller.UpdateDon(don);
  if (res === undefined) return;
  controller.showMessage("Don modifié avec succès.");
  ensurePageValid();
}
</script>
