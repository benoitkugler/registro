<template>
  <v-app>
    <v-main>
      <router-view />

      <v-snackbar
        style="z-index: 10000"
        app
        :model-value="message.text != ''"
        @update:model-value="message.text = ''"
        :timeout="message.timeout"
        :color="message.color"
        location="bottom left"
        close-on-content-click
      >
        {{ message.text }}
        <v-btn
          v-if="message.action"
          @click="message.action.action()"
          class="ml-2"
        >
          {{ message.action.title }}
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
</template>

<script lang="ts" setup>
import { reactive, ref } from "vue";
import { controller } from "./logic/logic";
import type { Action } from "@/utils";

const message = reactive({
  text: "",
  color: "primary",
  action: undefined as Action | undefined,
  timeout: 4000,
});

const errorKind = ref("");
const errorHtml = ref("");

controller.onError = (s, m) => {
  errorKind.value = s;
  errorHtml.value = m;
};

controller.showMessage = (s, color, action) => {
  message.text = s;
  message.color = color || "success";
  message.action = action;
  // reset the timeout by changing its value
  message.timeout = 10_000;
  message.timeout = 4000;
};
</script>
