<template>
  <v-card
    title="Albums photos"
    subtitle="Seuls les camps en cours sont affichés."
  >
    <template #append>
      <v-btn :disabled="!enableCreate" @click="createAlbums" class="mx-2">
        <template #prepend>
          <v-icon color="green">mdi-plus</v-icon>
        </template>
        Créer</v-btn
      >
      <v-btn :disabled="!enableDelete" @click="showConfirmeDelete = true">
        <template #prepend>
          <v-icon color="red">mdi-delete</v-icon>
        </template>
        Supprimer</v-btn
      >
    </template>
    <v-progress-linear v-if="isLoading" indeterminate></v-progress-linear>
    <v-card-text>
      <div class="text-center my-4" v-if="data == null">
        <v-progress-circular indeterminate></v-progress-circular>
      </div>
      <div v-else>
        <v-card variant="outlined">
          <v-card-text>
            <v-row>
              <v-col>Hébergement des photos</v-col>
              <v-col class="text-right">
                <a :href="data.HostURL">
                  {{ data.HostURL }}
                </a>
              </v-col>
            </v-row>
          </v-card-text>
        </v-card>
        <v-list v-model:selected="selection" select-strategy="leaf">
          <v-list-item
            v-for="camp in data.Camps"
            :value="camp.Camp.Id"
            :title="Camps.label(camp.Camp)"
            rounded
            class="my-1"
          >
            <template #append>
              <v-chip :color="camp.Album.Id ? 'success' : undefined">
                <v-tooltip
                  activator="parent"
                  v-if="camp.Album.Id"
                  content-class="pa-0"
                >
                  <v-card title="Album photo" min-width="400px">
                    <v-card-text>
                      <v-row>
                        <v-col cols="auto">Nom</v-col>
                        <v-col class="text-right">{{
                          camp.Album.albumName
                        }}</v-col>
                      </v-row>
                      <v-row>
                        <v-col cols="auto">ID</v-col>
                        <v-col class="text-right">{{ camp.Album.Id }}</v-col>
                      </v-row>
                      <v-row>
                        <v-col cols="auto">Date de création</v-col>
                        <v-col class="text-right">{{
                          new Date(camp.Album.createdAt).toLocaleDateString(
                            "fr"
                          )
                        }}</v-col>
                      </v-row>
                      <v-row>
                        <v-col cols="auto">Nombre de photos</v-col>
                        <v-col class="text-right">{{
                          camp.Album.assetCount
                        }}</v-col>
                      </v-row>
                    </v-card-text>
                  </v-card>
                </v-tooltip>
                {{
                  camp.Album.Id == ""
                    ? "Pas d'album"
                    : `Album créé le ${Formatters.date(camp.Album.createdAt)}`
                }}
              </v-chip>
            </template>
          </v-list-item>
        </v-list>
      </div>
    </v-card-text>

    <v-dialog v-model="showConfirmeDelete" max-width="600px">
      <v-card title="Confirmer la suppression">
        <v-card-text>
          Confirmez-vous la suppression des albums sélectionnés ? <br /><br />
          Les liens de partage ne seront plus valides. <br />
          Les photos ou vidéos elles-mêmes ne seront pas supprimées.
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="red" @click="deleteAlbums">Supprimer</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-card>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import {
  type Camp,
  type CampPhotos,
  type IdCamp,
} from "@/clients/backoffice/logic/api";
import { Camps, Formatters, recordEntries } from "@/utils";
import { controller } from "../../logic/logic";

const props = defineProps<{}>();
const emit = defineEmits<{}>();

onMounted(loadAlbums);

const data = ref<CampPhotos | null>(null);

const selection = ref<IdCamp[]>([]);

async function loadAlbums() {
  const res = await controller.CampsLoadAlbums();
  if (res === undefined) return;
  data.value = res;
  selection.value = [];
}

function getCamp(id: IdCamp) {
  return (data.value?.Camps || []).find((item) => item.Camp.Id == id)!;
}

const isLoading = ref(false);

const enableCreate = computed(() => {
  if (isLoading.value) return false;
  if (!selection.value.length || !data.value) return false;
  return selection.value.every((id) => getCamp(id).Album.Id == "");
});

async function createAlbums() {
  isLoading.value = true;
  const res = await controller.CampsCreateAlbums({ IdCamps: selection.value });
  isLoading.value = false;
  if (res === undefined) return;
  recordEntries(res).forEach(([id, album]) => {
    getCamp(id).Album = album;
  });
  controller.showMessage("Albums photos créés avec succès.");
}

const enableDelete = computed(() => {
  if (isLoading.value) return false;
  if (!selection.value.length || !data.value) return false;
  return selection.value.every((id) => getCamp(id).Album.Id != "");
});

const showConfirmeDelete = ref(false);
async function deleteAlbums() {
  showConfirmeDelete.value = false;
  isLoading.value = true;
  for (const idCamp of selection.value) {
    const res = await controller.CampsDeleteAlbum({ idCamp });
    if (res === undefined) break;
  }
  isLoading.value = false;
  controller.showMessage("Albums photos supprimés avec succès.");
  loadAlbums();
}
</script>
