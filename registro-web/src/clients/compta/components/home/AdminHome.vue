<template>
  <v-skeleton-loader type="card" v-if="!camps.length"></v-skeleton-loader>

  <v-card v-else title="Liste des séjours" class="mx-auto" width="900px">
    <!-- Confirm delete -->
    <v-dialog
      width="600"
      :model-value="toDelete != null"
      @update:model-value="toDelete = null"
    >
      <v-card title="Confirmez la suppression" v-if="toDelete != null">
        <v-card-text>
          Confirmez-vous la suppression du dossier de comptabilité pour le
          séjour <br />
          <b>{{ toDelete.Nom }} </b> ({{ toDelete.Serial }}) ? <br /><br />

          Les justificatifs associés seront supprimés.
          <br />

          Cette opération est irréversible.
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="red" @click="deleteCompta"> Supprimer </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- Bilan selection -->
    <v-dialog width="600" v-model="showToExport">
      <SelectBilanCamps
        :camps="camps.filter((c) => c.Compta.Id != 0)"
        :curent-year="currentYear?.year || 0"
        @download="downloadBilan"
      ></SelectBilanCamps>
    </v-dialog>

    <template v-slot:append>
      <v-menu location="left">
        <template #activator="{ props }">
          <v-btn v-bind="props" variant="text" icon="mdi-dots-vertical"></v-btn>
        </template>
        <v-list>
          <v-list-item
            :title="
              showNoCompta
                ? 'Masquer  les séjours sans dossier'
                : 'Afficher les séjours sans dossier'
            "
            :prepend-icon="showNoCompta ? 'mdi-eye-off' : 'mdi-eye'"
            @click="showNoCompta = !showNoCompta"
          ></v-list-item>
          <v-divider></v-divider>
          <v-list-item
            title="Exporter un bilan..."
            @click="showToExport = true"
            prepend-icon="mdi-microsoft-excel"
          ></v-list-item>
        </v-list>
      </v-menu>
    </template>
    <v-card-text>
      <v-row no-gutters justify="center">
        <v-col cols="auto">
          <v-btn-toggle
            rounded
            density="compact"
            v-model="currentPage"
            mandatory
            :multiple="false"
          >
            <v-btn v-for="(year, i) in byYear" :key="i">
              {{ year.year }}
            </v-btn>
          </v-btn-toggle>
        </v-col>
      </v-row>
      <v-list
        style="height: 63dvh"
        v-if="currentYear != null"
        class="overflow-y-auto mb-2"
      >
        <v-list-subheader>
          Camps de l'année {{ currentYear.year }}
        </v-list-subheader>
        <v-list-item
          v-for="(camp, j) in currentYear.camps"
          :key="j"
          :title="camp.Nom"
          :subtitle="camp.Serial"
          @[!!camp.Compta.Id&&`click`]="goTo(camp)"
          rounded
        >
          <template v-slot:append>
            <v-row>
              <v-col align-self="center">
                <v-chip
                  :elevation="camp.Compta.Id == 0 ? 1 : 0"
                  :color="
                    camp.Compta.Id == 0
                      ? 'grey'
                      : StateColors[camp.Compta.State]
                  "
                  @[!camp.Compta.Id&&`click`]="createAndGoTo(camp)"
                  :prepend-icon="camp.Compta.Id == 0 ? 'mdi-plus' : undefined"
                  :append-icon="StateIcons[camp.Compta.State]"
                  >{{
                    camp.Compta.Id == 0
                      ? "Ouvrir le dossier"
                      : StateLabels[camp.Compta.State]
                  }}</v-chip
                >
              </v-col>
              <v-col v-if="camp.Compta.Id != 0" cols="auto">
                <v-menu location="left">
                  <template #activator="{ props }">
                    <v-btn
                      size="small"
                      variant="text"
                      icon="mdi-dots-vertical"
                      v-bind="props"
                    ></v-btn>
                  </template>
                  <v-list>
                    <v-list-item
                      title="Télécharger les écritures"
                      prepend-icon="mdi-file-excel"
                      :href="downloadEcrituresURL(camp.Compta.Id)"
                    ></v-list-item>
                    <v-list-item
                      title="Copier le lien d'accès"
                      prepend-icon="mdi-link"
                      @click="copyLinkToClipboard(camp)"
                    >
                    </v-list-item>
                    <v-divider></v-divider>
                    <v-list-item
                      title="Supprimer"
                      prepend-icon="mdi-delete"
                      @click="toDelete = camp"
                    ></v-list-item>
                  </v-list>
                </v-menu>
              </v-col>
            </v-row>
          </template>
        </v-list-item>
      </v-list>
    </v-card-text>
  </v-card>
</template>

<script lang="ts" setup>
import { computed } from "vue";
import { onMounted } from "vue";
import { ref } from "vue";
import { useRouter } from "vue-router";
import {
  type CampHeader,
  State,
  type IdComptaCamp,
  StateLabels,
  type BilanCampsIn,
} from "../../logic/api.ts";
import { controller, StateColors, StateIcons } from "../../logic/logic.ts";
import { Camps, saveBlobAsFile } from "@/utils.ts";
import SelectBilanCamps from "./SelectBilanCamps.vue";
import { babelParse } from "vue/compiler-sfc";

interface year {
  year: number;
  camps: CampHeader[];
}

const camps = ref<CampHeader[]>([]);
const currentPage = ref(0); // 0 indexed
const currentYear = computed(() => {
  if (!byYear.value.length) return null;
  const v = byYear.value[currentPage.value];
  return {
    year: v.year,
    camps: v.camps.filter((c) => showNoCompta.value || c.Compta.Id != 0),
  };
});
const byYear = computed<year[]>(() => {
  const tmp = new Map<number, CampHeader[]>();
  camps.value.forEach((c) => {
    const year = Camps.year(c);
    const l = tmp.get(year) || [];
    l.push(c);
    tmp.set(year, l);
  });
  return Array.from(tmp.entries()).map((item) => ({
    year: item[0],
    camps: item[1].sort((a, b) => a.Id - b.Id),
  }));
});

const showNoCompta = ref(false);

onMounted(fetch);

async function fetch() {
  const res = await controller.GetCamps();
  if (res === undefined) return;
  camps.value = res || [];
}

const router = useRouter();

function goTo(camp: CampHeader) {
  controller.idCompta = camp.Compta.Id;

  router.push({ path: "/summary" });
}

async function createAndGoTo(camp: CampHeader) {
  const res = await controller.CreateCompta({ IdCamp: camp.Id });
  if (res === undefined) return;

  controller.showMessage("Dossier ouvert avec succès.");

  camp.Compta = {
    Id: res.Id,
    State: State.EnCours,
    Account: res.Account,
  };
  goTo(camp);
}

const toDelete = ref<CampHeader | null>(null);
async function deleteCompta() {
  const camp = toDelete.value;
  if (camp == null) return;

  const res = await controller.DeleteCompta({
    IdCompta: camp.Compta.Id,
  });
  if (res === undefined) return;

  controller.showMessage("Dossier supprimé avec succès.");
  if (controller.idCompta == camp.Compta.Id) {
    controller.idCompta = -1 as IdComptaCamp;
  }
  camps.value.find((c) => c.Id == camp.Id)!.Compta = {
    Id: 0 as IdComptaCamp,
    State: 0,
    Account: 0,
  };

  toDelete.value = null;
}

function downloadEcrituresURL(id: IdComptaCamp) {
  return controller.DownloadEcritures(id, controller.authToken);
}

const showToExport = ref(false);
async function downloadBilan(toExport: BilanCampsIn) {
  showToExport.value = false;
  controller.showMessage("Téléchargement en cours...");
  const res = await controller.DownloadBilanCamps(toExport);
  controller.showMessage("");
  if (res === undefined) return;

  saveBlobAsFile(res.blob, res.filename);
}

async function copyLinkToClipboard(camp: CampHeader) {
  try {
    await navigator.clipboard.writeText(camp.DirecteurURL);
  } catch (error) {
    return;
  }
  controller.showMessage("Lien copié dans le presse-papier.");
}
</script>
