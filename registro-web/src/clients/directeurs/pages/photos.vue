<template>
  <NavBar :title="`${controller.camp?.Label} - Album photos`"> </NavBar>

  <div v-if="data == null" class="text-center my-6">
    <v-progress-circular indeterminate></v-progress-circular>
  </div>
  <v-alert v-else-if="!data.HasAlbum" type="warning" class="ma-2"
    >Aucun album n'est associé au séjour.</v-alert
  >
  <v-card v-else title="Album" class="ma-2">
    <template #append>
      <v-btn
        @click="showConfirmeInvite = true"
        :disabled="invitingProgress != null"
      >
        <template #prepend>
          <v-icon color="green">mdi-send</v-icon>
        </template>
        Inviter...
      </v-btn>
    </template>
    <v-card-text>
      <v-row>
        <v-col align-self="center"
          >Nom de l'album : <b>{{ data.Album.albumName }}</b></v-col
        >
        <v-col align-self="center"
          >Créé le : <b>{{ Formatters.date(data.Album.createdAt) }}</b></v-col
        >
        <v-col align-self="center"
          >Nombre de photos : <b>{{ data.Album.assetCount }}</b></v-col
        >
      </v-row>
      <v-row>
        <v-col cols="4">Lien Équipe (permission d'ajout)</v-col>
        <v-col>
          <a target="_blank" :href="data.Album.EquipiersURL">{{
            data.Album.EquipiersURL
          }}</a>
        </v-col>
      </v-row>
      <v-row>
        <v-col cols="4">Lien Inscrits (lecture seule)</v-col>
        <v-col>
          <a target="_blank" :href="data.Album.InscritsURL">{{
            data.Album.InscritsURL
          }}</a>
        </v-col>
      </v-row>
    </v-card-text>

    <v-dialog v-model="showConfirmeInvite" max-width="800px">
      <v-card title="Envoyer le lien de l'album">
        <v-card-text>
          Confirmez-vous l'envoi d'un mail aux responsables et aux équipiers ?
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="green" @click="invite">Partager le lien par mail</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <v-dialog :model-value="invitingProgress != null">
      <RequestProgressCard
        v-if="invitingProgress != null"
        title="Envoi des mails"
        :current="invitingProgress.Current"
        :total="invitingProgress.Total"
      ></RequestProgressCard>
    </v-dialog>
  </v-card>
</template>

<script lang="ts" setup>
import { onMounted, ref } from "vue";
import NavBar from "../components/NavBar.vue";
import { controller } from "../logic/logic";
import { Formatters, readJSONStream } from "@/utils";
import type { Photos, SendProgress } from "../logic/api";

onMounted(loadData);

const data = ref<Photos | null>(null);
async function loadData() {
  const res = await controller.PhotosLoad();
  if (res === undefined) return;

  data.value = res;
}

const showConfirmeInvite = ref(false);
const invitingProgress = ref<SendProgress | null>(null);
async function invite() {
  if (!data.value) return;
  showConfirmeInvite.value = false;
  const res = await controller.PhotosInvite();
  if (res === undefined) return;
  await readJSONStream(
    res,
    (v) => (invitingProgress.value = v),
    (err) => controller.onError("Envoi du mail", err)
  );
  invitingProgress.value = null;
  controller.showMessage("Mails envoyés avec succès.");
}
</script>
