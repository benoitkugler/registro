<template>
  <v-app>
    <v-main>
      <v-app-bar rounded elevation="4" color="primary">
        <v-app-bar-title>
          <v-row>
            <v-col align-self="center" cols="auto">
              <v-img width="60" :src="logo" />
            </v-col>
            <v-col align-self="center">
              Bienvenue sur le Portail des inscriptions !
            </v-col>
          </v-row>
        </v-app-bar-title>
      </v-app-bar>

      <v-container style="min-height: 92%" v-if="data == null">
        <v-skeleton-loader type="card"></v-skeleton-loader>
      </v-container>
      <v-container class="fill-height" v-else-if="!data.Camps?.length">
        <v-responsive>
          <v-alert class="text-center">
            Aucun camp n'est encore ouvert aux inscriptions.
          </v-alert>
        </v-responsive>
      </v-container>
      <v-container class="py-2" style="min-height: 92%" v-else>
        <InscriptionPannel :data="data"></InscriptionPannel>
      </v-container>

      <v-footer color="secondary">
        <v-row no-gutters class="my-1" justify="space-between">
          <v-col
            >{{ asso }} -
            <a href="/cgu" class="text-black">Mentions légales et CGU</a></v-col
          >
          <v-col class="text-right">{{ version }}</v-col>
        </v-row>
      </v-footer>

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
import InscriptionPannel from "./components/InscriptionPannel.vue";
import type { Data } from "./logic/api";

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
const asso = import.meta.env.VITE_ASSO_TITLE;
const version = `v${VITE_APP_VERSION}`;

const data = ref<Data | null>(null);

onMounted(fetchData);

async function fetchData() {
  // forward url params
  const query = new URLSearchParams(window.location.search);
  const preselected = query.get("preselected") || "";
  const preinscription = query.get("preinscription") || "";
  const res = await controller.LoadData({ preselected, preinscription });
  if (res === undefined) return;
  // vue reactivity does not work if Participants is null
  res.InitialInscription.Participants =
    res.InitialInscription.Participants || [];
  data.value = res;
}
</script>
