<template>
  <v-card v-if="data != null" title="Documents">
    <v-card-text>
      <v-list>
        <v-list-subheader>Préférences d'affichage</v-list-subheader>
        <v-list-item>
          <v-row class="mb-2">
            <v-col>
              <v-chip
                :color="data.ToShow.LettreDirecteur ? 'green' : 'grey'"
                :append-icon="data.ToShow.LettreDirecteur ? 'mdi-check' : ''"
                >Lettre du directeur</v-chip
              >
            </v-col>
            <v-col>
              <v-chip
                :color="data.ToShow.ListeVetements ? 'green' : 'grey'"
                :append-icon="data.ToShow.ListeVetements ? 'mdi-check' : ''"
              >
                Liste de vêtements</v-chip
              >
            </v-col>
            <v-col>
              <v-chip
                :color="data.ToShow.ListeParticipants ? 'green' : 'grey'"
                :append-icon="data.ToShow.ListeParticipants ? 'mdi-check' : ''"
                >Liste des participants
              </v-chip>
            </v-col>
          </v-row>
        </v-list-item>
        <v-list-subheader>Documents du séjour</v-list-subheader>
        <v-list-item v-for="file in data.Generated" :title="file.NomClient">
          <template #append>
            <FileCardReadonly
              :file-key="file.Key"
              :is-generated-doc="true"
            ></FileCardReadonly>
          </template>
        </v-list-item>
        <v-list-item v-for="file in data.ToRead" :title="file.NomClient">
          <template #append>
            <FileCardReadonly
              :file-key="file.Key"
              :is-generated-doc="false"
            ></FileCardReadonly>
          </template>
        </v-list-item>
        <v-list-item
          v-for="file in data.ToUploadModeles"
          :title="file.NomClient"
          subtitle="Document à remplir"
        >
          <template #append>
            <FileCardReadonly
              :file-key="file.Key"
              :is-generated-doc="false"
            ></FileCardReadonly>
          </template>
        </v-list-item>
      </v-list>
    </v-card-text>
  </v-card>
  <v-skeleton-loader v-else type="card"></v-skeleton-loader>
</template>

<script setup lang="ts">
import type { CampHeader, FilesCamp } from "@/clients/backoffice/logic/api";
import { controller } from "../../logic/logic";
import { onMounted, ref } from "vue";
const props = defineProps<{
  camp: CampHeader;
}>();

onMounted(fetchData);

const data = ref<FilesCamp | null>(null);
async function fetchData() {
  const res = await controller.CampsDocuments({
    idCamp: props.camp.Camp.Camp.Id,
  });
  if (res === undefined) return;
  data.value = res;
}
</script>
