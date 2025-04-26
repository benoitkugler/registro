<template>
  <v-app>
    <v-main>
      <v-app-bar rounded elevation="4" color="primary">
        <v-app-bar-title>
          <v-row>
            <v-col align-self="center" cols="auto">
              <v-img width="60" :src="logo" />
            </v-col>
            <v-col align-self="center"> Portail </v-col>
          </v-row>
        </v-app-bar-title>
      </v-app-bar>

      <v-container class="fill-height">
        <v-responsive>
          <v-skeleton-loader v-if="isLoading"></v-skeleton-loader>
          <v-alert
            v-else-if="service == Service.TransfertFicheSanitaire"
            title="Partage de fiche sanitaire"
            type="success"
          >
            L'accès à la fiche a bien été mis à jour. <br />

            <small class="text-muted"
              >Vous pouvez quitter cette page et actualiser votre espace de
              suivi pour y accéder.</small
            >
          </v-alert>
        </v-responsive>
      </v-container>

      <v-snackbar
        style="z-index: 10000"
        app
        :model-value="message != ''"
        @update:model-value="message = ''"
        :timeout="4000"
        :color="messageColor"
        location="bottom left"
        close-on-content-click
      >
        {{ message }}
      </v-snackbar>

      <v-snackbar
        app
        :model-value="errorKind != ''"
        @update:model-value="errorKind = ''"
        :timeout="4000"
        color="red"
      >
        <b>{{ errorKind }}</b>
        <div v-html="errorHtml"></div>
      </v-snackbar>
    </v-main>
  </v-app>
</template>

<script lang="ts" setup>
import { onMounted, ref } from "vue";
import { controller } from "./logic/logic";
import { Service } from "./logic/types";

const message = ref("");
const messageColor = ref("secondary");

const errorKind = ref("");
const errorHtml = ref("");

controller.onError = (s, m) => {
  errorKind.value = s;
  errorHtml.value = m;
};

controller.showMessage = (s, color) => {
  message.value = s;
  messageColor.value = color || "success";
};

const logo = `${import.meta.env.BASE_URL}${import.meta.env.VITE_ASSO}/logo.png`;

const isLoading = ref(true);
const service = ref<Service>(1);

onMounted(() => {
  const query = new URLSearchParams(window.location.search);
  const param = query.get("service") || "";
  const v = Number(param) as Service;
  switch (v) {
    case Service.TransfertFicheSanitaire:
      service.value = v;
      const token = query.get("token") || "";
      return valideTransfertFicheSanitaire(token);
    default:
      controller.onError(
        "Service invalide",
        `Le paramètre <i>service</i> a une valeur incorrecte : ${param}`
      );
  }
});

async function valideTransfertFicheSanitaire(token: string) {
  const res = await controller.ValideTransfertFicheSanitaire({ token });
  isLoading.value = false;
  if (res === undefined) return;
}
</script>
