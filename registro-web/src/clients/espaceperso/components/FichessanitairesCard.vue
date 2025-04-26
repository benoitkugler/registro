<template>
  <v-card title="Fiches sanitaires et vaccins">
    <template #append>
      <v-tabs v-model="tab">
        <v-tab v-for="fiche in fiches" :value="fiche.Fichesanitaire.IdPersonne">
          {{ fiche.Personne }}
          <v-icon
            class="ml-2"
            v-if="
              fiche.State == FichesanitaireState.Empty ||
              fiche.State == FichesanitaireState.Outdated ||
              !fiche.VaccinsFiles?.length
            "
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
            @upload-vaccin="(f) => uploadVaccin(index, f)"
            @delete-vaccin="(f) => deleteVaccin(index, f)"
            @transfert="transfert(fiche.Fichesanitaire)"
          ></FichesanitaireForm>
        </v-tabs-window-item>
      </v-tabs-window>
    </v-card-text>
  </v-card>
</template>

<script setup lang="ts">
import { onMounted, ref } from "vue";
import {
  FichesanitaireState,
  type Aide,
  type Fichesanitaire,
  type FichesanitaireExt,
  type IdPersonne,
  type PublicFile,
} from "../logic/api";
import { controller } from "../logic/logic";
import FichesanitaireForm from "./FichesanitaireForm.vue";
const props = defineProps<{
  token: string;
}>();

const emit = defineEmits<{
  (e: "close"): void;
}>();

onMounted(fetchData);

const tab = ref<IdPersonne>(0 as IdPersonne);

const fiches = ref<FichesanitaireExt[]>([]);
async function fetchData() {
  const res = await controller.LoadFichesanitaires({ token: props.token });
  if (res === undefined) return;
  fiches.value = res || [];
  if (res?.length && tab.value == 0) {
    tab.value = res[0].Fichesanitaire.IdPersonne;
  }
}

async function save(fiche: Fichesanitaire, respoSecuriteSociale: string) {
  const res = await controller.UpdateFichesanitaire({
    Token: props.token,
    Fichesanitaire: fiche,
    SecuriteSocialeResponsable: respoSecuriteSociale,
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
  controller.showMessage("Le vaccin a bien été téléversé. Merci !");
  fiche.VaccinsFiles = (fiche.VaccinsFiles || []).concat(res);
}

async function deleteVaccin(index: number, file: PublicFile) {
  const fiche = fiches.value[index];
  const res = await controller.DeleteVaccin({ key: file.Id });
  if (res === undefined) return;
  controller.showMessage("Le vaccin a bien été supprimé.");
  fiche.VaccinsFiles = (fiche.VaccinsFiles || []).filter(
    (v) => v.Id != file.Id
  );
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
