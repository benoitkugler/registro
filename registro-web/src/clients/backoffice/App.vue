<template>
  <v-app>
    <v-main>
      <router-view />

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
        <v-btn
          v-if="messageAction"
          @click="messageAction.action()"
          class="ml-2"
        >
          {{ messageAction.title }}
        </v-btn>
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

  <!-- <router-view v-if="isLoggedIn" />
      <v-skeleton-loader v-else type="card"></v-skeleton-loader> -->
</template>

<script lang="ts" setup>
import { ref } from "vue";
import { controller, type Action } from "./logic/logic";

const message = ref("");
const messageColor = ref("secondary");
const messageAction = ref<Action | undefined>(undefined);

const errorKind = ref("");
const errorHtml = ref("");

controller.onError = (s, m) => {
  errorKind.value = s;
  errorHtml.value = m;
};

controller.showMessage = (s, color, action) => {
  message.value = s;
  messageColor.value = color || "success";
  messageAction.value = action;
};
</script>
