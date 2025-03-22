<template>
  <v-app>
    <v-main>
      <v-app-bar rounded elevation="4" color="primary">
        <v-app-bar-title>
          <v-row>
            <v-col align-self="center" cols="auto">
              <v-img width="60" :src="logo" />
            </v-col>
            <v-col align-self="center"> Formulaire équipier </v-col>
          </v-row>
        </v-app-bar-title>
      </v-app-bar>

      <v-container v-if="data == null">
        <v-skeleton-loader type="card"></v-skeleton-loader>
      </v-container>
      <v-container class="py-2" style="min-height: 92%" v-else>
        <EquipierForm :equipier="data" :joomeo="joomeo"></EquipierForm>
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
import type { EquipierExt, Joomeo } from "./logic/api";
import EquipierForm from "./components/EquipierForm.vue";

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

const data = ref<EquipierExt | null>(null);

// id key
const key = ref("");

onMounted(() => {
  // store ID key
  const query = new URLSearchParams(window.location.search);
  key.value = query.get("key") || "";
  fetchData();
  fetchJoomeo();
});

async function fetchData() {
  const res = await controller.Load({ key: key.value });
  if (res === undefined) return;
  data.value = res;
}

const joomeo = ref<Joomeo | null>(null);
async function fetchJoomeo() {
  const res = await controller.LoadJoomeo({ key: key.value });
  if (res === undefined) return;
  joomeo.value = res;
}
</script>
