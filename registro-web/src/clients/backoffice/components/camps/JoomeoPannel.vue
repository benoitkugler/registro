<template>
  <v-card
    title="Albums Joomeo"
    subtitle="Seuls les camps en cours sont affichés."
  >
    <template #append>
      <v-btn :disabled="!enableCreate" @click="createAlbums">
        <template #prepend>
          <v-icon color="green">mdi-plus</v-icon>
        </template>
        Créer les albums</v-btn
      >
      <v-btn
        :disabled="!enableAddDirecteurs"
        @click="addDirecteurs"
        class="ml-2"
        prepend-icon="mdi-account-eye"
      >
        Ajouter les directeurs</v-btn
      >
    </template>
    <v-progress-linear v-if="isLoading" indeterminate></v-progress-linear>
    <v-card-text>
      <div class="text-center my-4" v-if="data == null">
        <v-progress-circular indeterminate></v-progress-circular>
      </div>
      <v-list v-else v-model:selected="selection" select-strategy="leaf">
        <v-list-item
          v-for="camp in data"
          :value="camp.Camp.Id"
          :title="Camps.label(camp.Camp)"
          rounded
          class="my-1"
        >
          <template #append>
            <v-row>
              <v-col>
                <v-chip :color="camp.Album.Id ? 'primary' : undefined">
                  <v-tooltip
                    activator="parent"
                    v-if="camp.Album.Id"
                    content-class="pa-0"
                  >
                    <v-card title="Album Joomeo" min-width="400px">
                      <v-card-text>
                        <v-row>
                          <v-col cols="auto">Nom</v-col>
                          <v-col class="text-right">{{
                            camp.Album.Label
                          }}</v-col>
                        </v-row>
                        <v-row>
                          <v-col cols="auto">ID</v-col>
                          <v-col class="text-right">{{ camp.Album.Id }}</v-col>
                        </v-row>
                        <v-row>
                          <v-col cols="auto">Date de création</v-col>
                          <v-col class="text-right">{{
                            new Date(camp.Album.Date).toLocaleDateString("fr")
                          }}</v-col>
                        </v-row>
                        <v-row>
                          <v-col cols="auto">Nombre de photos</v-col>
                          <v-col class="text-right">{{
                            camp.Album.FilesCount
                          }}</v-col>
                        </v-row>
                      </v-card-text>
                    </v-card>
                  </v-tooltip>
                  {{
                    camp.Album.Id == ""
                      ? "Pas d'album"
                      : `Album créé le ${Formatters.date(camp.Album.Date)}`
                  }}
                </v-chip>
              </v-col>
              <v-col v-if="camp.Album.Id">
                <v-chip
                  :color="
                    camp.HasDirecteur
                      ? camp.DirecteurPermission.contactid != ''
                        ? 'green'
                        : 'orange'
                      : 'red'
                  "
                >
                  <v-tooltip
                    activator="parent"
                    v-if="camp.DirecteurPermission.contactid"
                    content-class="pa-0"
                  >
                    <v-card
                      title="Accès à l'album"
                      subtitle="Le directeur a accès à l'album."
                    >
                      <v-card-text>
                        <v-row>
                          <v-col>Email</v-col>
                          <v-col class="text-right">{{
                            camp.DirecteurPermission.email
                          }}</v-col>
                        </v-row>
                        <v-row>
                          <v-col>Login</v-col>
                          <v-col class="text-right">{{
                            camp.DirecteurPermission.login
                          }}</v-col>
                        </v-row>
                        <v-row>
                          <v-col>Mot de passe</v-col>
                          <v-col class="text-right">{{
                            camp.DirecteurPermission.password
                          }}</v-col>
                        </v-row>
                        <v-row>
                          <v-col>Permissions (écriture)</v-col>
                          <v-col class="text-right">
                            <v-icon
                              :icon="
                                hasDirecteurAllPermissions(
                                  camp.DirecteurPermission
                                )
                                  ? 'mdi-check'
                                  : 'mdi-warning'
                              "
                              :color="
                                hasDirecteurAllPermissions(
                                  camp.DirecteurPermission
                                )
                                  ? 'green'
                                  : 'orange'
                              "
                            ></v-icon>
                          </v-col>
                        </v-row>
                      </v-card-text>
                    </v-card>
                  </v-tooltip>
                  {{
                    camp.HasDirecteur
                      ? camp.DirecteurPermission.contactid != ""
                        ? "Directeur ajouté"
                        : "Directeur à ajouter"
                      : "Aucun directeur"
                  }}
                </v-chip>
              </v-col>
            </v-row>
          </template>
        </v-list-item>
      </v-list>
    </v-card-text>
    <v-card-actions>
      <v-spacer></v-spacer>
      <!-- <v-btn :disabled="!areFieldsValid" @click="emit('save', inner)"
        >Enregistrer</v-btn
      > -->
    </v-card-actions>
  </v-card>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import {
  type Camp,
  type CampJoomeo,
  type ContactPermission,
  type IdCamp,
} from "@/clients/backoffice/logic/api";
import { Camps, Formatters, mapFromObject, recordEntries } from "@/utils";
import { controller } from "../../logic/logic";

const props = defineProps<{}>();
const emit = defineEmits<{
  (e: "save", camp: Camp): void;
}>();

onMounted(loadAlbums);

const data = ref<CampJoomeo[] | null>(null);

const selection = ref<IdCamp[]>([]);

async function loadAlbums() {
  const res = await controller.CampsLoadAlbums();
  if (res === undefined) return;
  data.value = res || [];
  selection.value = [];
}

function getCamp(id: IdCamp) {
  return data.value!.find((item) => item.Camp.Id == id)!;
}

const isLoading = ref(false);

const enableCreate = computed(() => {
  if (isLoading.value) return false;
  if (!selection.value.length || !data.value) return false;
  return selection.value.every((id) => getCamp(id).Camp.JoomeoID == "");
});

async function createAlbums() {
  isLoading.value = true;
  const res = await controller.CampsCreateAlbums({ IdCamps: selection.value });
  isLoading.value = false;
  if (res === undefined) return;
  recordEntries(res).forEach(([id, album]) => {
    getCamp(id).Album = album;
  });
  controller.showMessage("Albums Joomeo créés avec succès.");
}

const enableAddDirecteurs = computed(() => {
  if (isLoading.value) return false;
  if (!selection.value.length || !data.value) return false;
  return selection.value.every(
    (id) => getCamp(id).Camp.JoomeoID != "" && getCamp(id).HasDirecteur
  );
});

async function addDirecteurs() {
  isLoading.value = true;
  const res = await controller.CampsAddDirecteursToAlbums({
    IdCamps: selection.value,
    SendMail: true,
  });
  isLoading.value = false;
  if (res === undefined) return;
  recordEntries(res).forEach(([id, perms]) => {
    getCamp(id).DirecteurPermission = perms;
  });
  controller.showMessage("Directeurs ajoutés avec succès.");
}

function hasDirecteurAllPermissions(perms: ContactPermission) {
  return (
    perms.type == 1 &&
    perms.albumAccessRules.allowDownload &&
    perms.albumAccessRules.allowUpload &&
    perms.accessRules.allowDeleteFile &&
    perms.accessRules.allowEditFileCaption
  );
}
</script>
