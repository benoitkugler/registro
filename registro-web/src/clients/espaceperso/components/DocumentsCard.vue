<template>
  <v-card
    title="Documents des séjours"
    subtitle="Vous retrouvez ici les documents des séjours auxquels vous participez, à lire ou remplir."
  >
    <template #append> </template>
    <v-card-text v-if="data != null">
      <v-list>
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
        <v-list-subheader>À fournir</v-list-subheader>
        <v-list-item v-if="!data.FilesToUpload?.length">
          <i>Aucun document à fournir.</i>
        </v-list-item>
        <template v-for="personne in data.FilesToUpload">
          <FilesDemande
            :demande="demande.Demande"
            :files="demande.Uploaded || []"
            :inUpload="false"
            :title="`Document pour ${personne.Personne}`"
            :optionnelle="null"
            :showUploadText="true"
            v-for="demande in personne.Demandes"
            @upload="
              (f) => uploadDocument(personne.IdPersonne, demande.Demande.Id, f)
            "
            @delete="deleteDocument"
          >
            <template #prepend v-if="demande.Demande.IdFile.Valid">
              <v-tooltip content-class="pa-1">
                <template #activator="{ props: tooltipProps }">
                  <v-btn
                    v-bind="tooltipProps"
                    :href="contentURL(demande.DemandeFile.Key)"
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
  </v-card>
</template>

<script setup lang="ts">
import { onMounted, ref } from "vue";
import { type Documents, type IdPersonne, type PublicFile } from "../logic/api";
import { controller } from "../logic/logic";
import type { IdDemande, Int } from "@/urls";
import { contentURL } from "@/utils";
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
  emit("updateNotifs", res.ToReadOrFillCount);
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
  emit("updateNotifs", data.value!.ToReadOrFillCount);
}

async function deleteDocument(file: PublicFile) {
  const res = await controller.DeleteDocument({ key: file.Key });
  if (res === undefined) return;
  controller.showMessage("Document supprimé avec succès.");
  await fetchData();
  emit("updateNotifs", data.value!.ToReadOrFillCount);
}
</script>
