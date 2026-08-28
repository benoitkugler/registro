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
        <v-col align-self="center" cols="4"
          >Lien Équipe (permission d'ajout)</v-col
        >
        <v-col align-self="center">
          <a target="_blank" :href="data.Album.EquipiersURL">{{
            data.Album.EquipiersURL
          }}</a>
        </v-col>
        <v-col align-self="center" cols="1"></v-col>
      </v-row>
      <v-row>
        <v-col align-self="center" cols="4"
          >Lien Inscrits (lecture seule)</v-col
        >
        <v-col align-self="center">
          <a target="_blank" :href="data.Album.InscritsURL">{{
            data.Album.InscritsURL
          }}</a>
        </v-col>
        <v-col align-self="center" cols="1">
          <v-tooltip
            :text="
              data.IsAlbumVisible
                ? 'Le lien est visible sur les espaces de suivi.'
                : 'Envoyer un mail aux responsables pour afficher le lien sur les espaces de suivi.'
            "
          >
            <template #activator="{ props: tooltipProps }">
              <v-icon
                v-bind="tooltipProps"
                :icon="data.IsAlbumVisible ? 'mdi-eye' : 'mdi-lock'"
              ></v-icon>
            </template>
          </v-tooltip>
        </v-col>
      </v-row>
    </v-card-text>

    <v-dialog v-model="showConfirmeInvite" max-width="800px">
      <v-card title="Envoyer le lien de l'album">
        <v-card-text>
          <v-row>
            <v-col> Confirmez-vous l'envoi d'un mail d'invitation ? </v-col>
          </v-row>
          <v-row>
            <v-col
              ><v-checkbox
                density="compact"
                label="Aux équipiers"
                v-model="inviteArgs.ToEquipiers"
                persistent-hint
                hint="Le lien permet l'ajout de photos."
              ></v-checkbox
            ></v-col>
            <v-col
              ><v-checkbox
                density="compact"
                label="Aux responsables"
                v-model="inviteArgs.ToResponsables"
                persistent-hint
                hint="Le lien permet uniquement la visualisation des photos."
              ></v-checkbox
            ></v-col>
          </v-row>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn
            color="green"
            :disabled="!(inviteArgs.ToResponsables || inviteArgs.ToEquipiers)"
            @click="invite"
            >Partager le lien par mail</v-btn
          >
        </v-card-actions>
      </v-card>
    </v-dialog>

    <v-dialog :model-value="invitingProgress != null">
      <RequestProgressCard
        v-if="invitingProgress != null"
        title="Envoi des mails"
        :progress="invitingProgress"
      ></RequestProgressCard>
    </v-dialog>
  </v-card>
</template>

<script lang="ts" setup>
import { onMounted, ref } from "vue";
import NavBar from "../components/NavBar.vue";
import { controller } from "../logic/logic";
import { Formatters, readJSONStream } from "@/utils";
import type { Photos, PhotosInviteIn, SendProgress } from "../logic/api";

onMounted(loadData);

const data = ref<Photos | null>(null);
async function loadData() {
  const res = await controller.PhotosLoad();
  if (res === undefined) return;

  data.value = res;
}

const showConfirmeInvite = ref(false);
const inviteArgs = ref<PhotosInviteIn>({
  ToResponsables: true,
  ToEquipiers: true,
});
const invitingProgress = ref<SendProgress | null>(null);
async function invite() {
  if (!data.value) return;
  showConfirmeInvite.value = false;
  const res = await controller.PhotosInvite(inviteArgs.value);
  if (res === undefined) return;
  await readJSONStream(
    res,
    (v) => (invitingProgress.value = v),
    (err) => controller.onError("Envoi du mail", err),
  );
  invitingProgress.value = null;
  controller.showMessage("Mails envoyés avec succès.");
  loadData();
}
</script>
