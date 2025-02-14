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
              Bienvenu sur le Portail des inscriptions !
            </v-col>
          </v-row>
        </v-app-bar-title>
      </v-app-bar>

      <v-container class="py-2" style="min-height: 92%">
        <InscriptionPannel></InscriptionPannel>
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
import { ref } from "vue";
import { controller } from "./logic/logic";
import InscriptionPannel from "./components/InscriptionPannel.vue";

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
</script>
