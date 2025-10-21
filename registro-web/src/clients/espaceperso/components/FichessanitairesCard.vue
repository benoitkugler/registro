<template>
  <v-card title="Fiches sanitaires">
    <template #append>
      <v-tabs v-model="tab">
        <v-tab v-for="fiche in fiches" :value="fiche.Fichesanitaire.IdPersonne">
          {{ fiche.Personne }}
          <v-icon
            class="ml-2"
            v-if="fiche.State != FichesanitaireState.UpToDate"
            color="warning"
          >
            mdi-alert
          </v-icon>
          <v-icon v-else class="ml-2" color="success"> mdi-check </v-icon>
        </v-tab>
      </v-tabs>
    </template>
    <v-card-text>
      <v-tabs-window :model-value="tab">
        <v-tabs-window-item
          v-for="(fiche, index) in fiches"
          :value="fiche.Fichesanitaire.IdPersonne"
        >
          <FichesanitaireForm
            :fiche="fiche"
            @save="save"
            @transfert="transfert(fiche.Fichesanitaire)"
          ></FichesanitaireForm>
          <!-- @upload-vaccin="(f) => uploadVaccin(index, f)"
            @delete-vaccin="(f) => deleteVaccin(index, f)" -->
        </v-tabs-window-item>
      </v-tabs-window>
    </v-card-text>
  </v-card>
</template>

<script setup lang="ts">
import { onMounted, ref } from "vue";
import {
  FichesanitaireState,
  type Fichesanitaire,
  type FichesanitaireExt,
  type IdPersonne,
  type Int,
  type PublicFile,
} from "../logic/api";
import { controller } from "../logic/logic";
import FichesanitaireForm from "./FichesanitaireForm.vue";
const props = defineProps<{
  token: string;
}>();

const emit = defineEmits<{
  (e: "close"): void;
  (e: "updateNotifs", toFill: Int): void;
}>();

onMounted(fetchData);

const tab = ref<IdPersonne>(0 as IdPersonne);

const fiches = ref<FichesanitaireExt[]>([]);
async function fetchData() {
  const res = await controller.LoadFichesanitaires({ token: props.token });
  if (res === undefined) return;
  fiches.value = res.Fiches || [];
  if (fiches.value.length && tab.value == 0) {
    tab.value = fiches.value[0].Fichesanitaire.IdPersonne;
  }
  emit("updateNotifs", res.ToFillCount);
}

async function save(fiche: Fichesanitaire) {
  const res = await controller.UpdateFichesanitaire({
    Token: props.token,
    Fichesanitaire: fiche,
  });
  if (res === undefined) return;
  controller.showMessage("La fiche sanitaire a bien été enregistrée. Merci !");
  fetchData();
}

async function uploadVaccin(index: number, file: File) {
  const fiche = fiches.value[index];
  const res = await controller.UploadVaccin(file, {
    token: props.token,
    idPersonne: fiche.Fichesanitaire.IdPersonne,
  });
  if (res === undefined) return;
  fetchData();
  controller.showMessage("Le vaccin a bien été téléversé. Merci !");
}

async function deleteVaccin(index: number, file: PublicFile) {
  const fiche = fiches.value[index];
  const res = await controller.DeleteVaccin({ key: file.Key });
  if (res === undefined) return;
  fetchData();
  controller.showMessage("Le vaccin a bien été supprimé.");
}

async function transfert(fiche: Fichesanitaire) {
  emit("close");
  const res = await controller.TransfertFicheSanitaire({
    token: props.token,
    idPersonne: fiche.IdPersonne,
  });
  if (res === undefined) return;
  controller.showMessage("Mail de transfert envoyé avec succès.");
}
</script>
