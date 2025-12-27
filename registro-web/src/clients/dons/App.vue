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
              Formulaire équipier {{ data ? Camps.label(data.Camp) : "" }}
            </v-col>
          </v-row>
        </v-app-bar-title>
      </v-app-bar>

      <v-container v-if="data == null">
        <v-skeleton-loader type="card"></v-skeleton-loader>
      </v-container>
      <v-container class="py-2" style="min-height: 92%" v-else>
        <EquipierForm
          :token="token"
          :equipier="data"
          :album="joomeo"
        ></EquipierForm>
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

const data = ref<EquipierExt | null>(null);

// id token
const token = ref("");

onMounted(() => {
  // store ID token
  const query = new URLSearchParams(window.location.search);
  token.value = query.get("token") || "";
});
</script>
