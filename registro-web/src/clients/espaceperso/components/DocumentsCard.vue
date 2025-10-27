<template>
  <v-card
    title="Documents des séjours"
    subtitle="Vous retrouvez ici les documents des séjours auxquels vous participez, à lire ou remplir."
  >
    <template #append v-if="data">
      <v-chip v-if="data.NewCount == 0" color="green" append-icon="mdi-check">
        A jour</v-chip
      >
      <v-badge inline :content="data.NewCount" color="pink" v-else></v-badge>
    </template>
    <v-card-text v-if="data != null">
      <v-list>
        <!-- A lire -->
        <v-list-subheader>À lire</v-list-subheader>
        <v-list-item v-if="!data.FilesToRead?.length">
          <i>Aucun document à lire.</i>
        </v-list-item>
        <template v-for="camp in data.FilesToRead">
          <!-- generated last -->
          <v-list-item
            v-for="file in camp.Files"
            :title="file.NomClient"
            :subtitle="camp.Camp"
          >
            <template #append>
              <FileCardReadonly
                :fileKey="file.Key"
                :isGeneratedDoc="false"
              ></FileCardReadonly>
            </template>
          </v-list-item>
          <v-list-item
            v-for="file in camp.Generated"
            :title="file.NomClient"
            :subtitle="camp.Camp"
          >
            <template #append>
              <FileCardReadonly
                :fileKey="file.Key"
                :isGeneratedDoc="true"
              ></FileCardReadonly>
            </template>
          </v-list-item>
        </template>
        <!-- A fournir -->
        <v-list-subheader>À fournir</v-list-subheader>
        <v-list-item
          v-if="!(data.FilesToUpload?.length || data.Fiches?.length)"
        >
          <i>Aucun document à fournir.</i>
        </v-list-item>
        <!-- chartes -->
        <v-list-item
          v-for="charte in data.Chartes"
          title="Charte"
          :subtitle="charte.Personne"
        >
          <template #append>
            <v-chip
              v-if="charte.Accepted"
              color="green"
              @click="charteToShow = charte.Id"
              append-icon="mdi-check"
            >
              Acceptée
            </v-chip>
            <v-btn size="small" @click="charteToShow = charte.Id" v-else>
              <template #prepend>
                <v-icon>mdi-pencil</v-icon>
              </template>
              Signer</v-btn
            >
          </template>
        </v-list-item>
        <!-- fiches sanitaires -->
        <v-list-item
          v-for="fiche in data.Fiches"
          title="Fiche sanitaire"
          :subtitle="fiche.Personne"
        >
          <template #append>
            <v-chip
              v-if="fiche.State == FichesanitaireState.UpToDate"
              @click="ficheToEdit = fiche"
              color="green"
              append-icon="mdi-check"
            >
              Remplie
            </v-chip>
            <v-btn size="small" @click="ficheToEdit = fiche" v-else>
              <template #prepend>
                <v-icon>mdi-pencil</v-icon>
              </template>
              Remplir</v-btn
            >
          </template>
        </v-list-item>
        <template v-for="personne in data.FilesToUpload">
          <FilesDemande
            :demande="demande.Demande"
            :files="demande.Uploaded || []"
            :subtitle="personne.Personne"
            :inUpload="false"
            :optionnelle="null"
            :showUploadText="true"
            v-for="demande in personne.Demandes"
            @upload="
              (f) => uploadDocument(personne.IdPersonne, demande.Demande.Id, f)
            "
            @delete="deleteDocument"
          >
            <template
              #prepend
              v-if="demande.Demande.Categorie == Categorie.Vaccins"
            >
              <v-btn icon size="x-small" flat class="mr-2">
                <v-icon>mdi-help-circle-outline</v-icon>
                <v-menu activator="parent">
                  <v-card max-width="400px">
                    <v-card-text>
                      Merci de joindre le scan des pages « vaccinations » du
                      carnet de santé du participant. <br />
                      Seul le DTPolio est obligatoire pour être accueilli en
                      séjour de vacances.
                      <br />
                      <i>
                        Si le participant n’a pas les vaccins obligatoires,
                        joindre un certificat médical de contre-indication.
                      </i>
                    </v-card-text>
                  </v-card>
                </v-menu>
              </v-btn>
            </template>
            <template #prepend v-if="demande.Demande.IdFile.Valid">
              <v-tooltip content-class="pa-1">
                <template #activator="{ props: tooltipProps }">
                  <v-btn
                    v-bind="tooltipProps"
                    :href="endpoints.LoadDocument(demande.DemandeFile.Key)"
                    class="mr-2"
                    size="small"
                  >
                    <template #prepend>
                      <v-icon>mdi-download</v-icon>
                    </template>
                    Modèle</v-btn
                  >
                </template>
                <FileCardReadonly
                  :file-key="demande.DemandeFile.Key"
                  :isGeneratedDoc="false"
                  :large="true"
                ></FileCardReadonly>
              </v-tooltip>
            </template>
          </FilesDemande>
        </template>
      </v-list>
    </v-card-text>
    <v-skeleton-loader v-else></v-skeleton-loader>

    <!-- charte -->
    <v-dialog
      :model-value="charteToShow != null"
      @update:model-value="charteToShow = null"
      max-width="800px"
    >
      <CharteCard v-if="charteToShow" @accept="acceptCharte"></CharteCard>
    </v-dialog>

    <!-- fiche sanitaire -->
    <v-dialog
      :model-value="ficheToEdit != null"
      @update:model-value="ficheToEdit = null"
      max-width="800px"
    >
      <FichesanitaireForm
        v-if="ficheToEdit"
        :fiche="ficheToEdit"
        @save="saveFichesanitaire"
        @transfert="transfertFichesanitaire"
      ></FichesanitaireForm>
    </v-dialog>
  </v-card>
</template>

<script setup lang="ts">
import { onMounted, ref } from "vue";
import {
  Categorie,
  FichesanitaireState,
  type Documents,
  type Fichesanitaire,
  type FichesanitaireExt,
  type IdDemande,
  type IdPersonne,
  type PublicFile,
} from "../logic/api";
import { controller } from "../logic/logic";
import type { Int } from "@/urls";
import { endpoints } from "@/utils";
import FichesanitaireForm from "./FichesanitaireForm.vue";
import CharteCard from "./CharteCard.vue";
const props = defineProps<{
  token: string;
}>();

const emit = defineEmits<{
  (e: "updateNotifs", toReadOrFill: Int): void;
}>();

onMounted(fetchData);

const data = ref<Documents | null>(null);
async function fetchData() {
  const res = await controller.LoadDocuments({ token: props.token });
  if (res === undefined) return;
  data.value = res;
  emit("updateNotifs", res.NewCount);
}

async function uploadDocument(
  idPersonne: IdPersonne,
  idDemande: IdDemande,
  file: File
) {
  const res = await controller.UploadDocument(file, {
    token: props.token,
    idPersonne,
    idDemande,
  });
  if (res === undefined) return;
  controller.showMessage("Document téléversé avec succès. Merci !");
  await fetchData();
}

async function deleteDocument(file: PublicFile) {
  const res = await controller.DeleteDocument({ key: file.Key });
  if (res === undefined) return;
  controller.showMessage("Document supprimé avec succès.");
  await fetchData();
}

const ficheToEdit = ref<FichesanitaireExt | null>(null);

async function saveFichesanitaire(fiche: Fichesanitaire) {
  ficheToEdit.value = null;
  const res = await controller.UpdateFichesanitaire({
    Token: props.token,
    Fichesanitaire: fiche,
  });
  if (res === undefined) return;
  controller.showMessage("La fiche sanitaire a bien été enregistrée. Merci !");
  fetchData();
}

async function transfertFichesanitaire(fiche: Fichesanitaire) {
  ficheToEdit.value = null;
  const res = await controller.TransfertFicheSanitaire({
    token: props.token,
    idPersonne: fiche.IdPersonne,
  });
  if (res === undefined) return;
  controller.showMessage("Mail de transfert envoyé avec succès.");
}

const charteToShow = ref<IdPersonne | null>(null);
async function acceptCharte() {
  const id = charteToShow.value;
  if (id == null) return;
  charteToShow.value = null;
  const res = await controller.AccepteCharte({
    token: props.token,
    idPersonne: id,
  });
  if (res === undefined) return;
  controller.showMessage("La charte a bien été acceptée. Merci !");
  fetchData();
}
</script>
