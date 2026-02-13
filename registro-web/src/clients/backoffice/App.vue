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
import { CachedTokens, controller } from "./logic/logic";
import { type Action } from "@/utils";
import { useRouter } from "vue-router";

const message = reactive({
  text: "",
  color: "accent",
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

// if we are not yet logged in,
// redirect to the index page
const router = useRouter();
router.beforeEach(async (to, from) => {
  if (controller.hasToken()) return; // nothing to do

  // try do loggin from cached token
  const cachedToken = CachedTokens.get();
  if (cachedToken) {
    const res = await controller.Loggin({ password: cachedToken });
    if (res && res.IsValid) {
      controller.setToken(res.Token, res.IsFondSoutien);
      controller.showMessage("Bienvenue !");
      if (to.path == "/") {
        // we are on the loggin page, dont stay here !
        return { path: "/inscriptions" };
      } else {
        return; // do not redirect to login page
      }
    }
  }

  if (to.path !== "/") {
    // redirect the user to the login page
    return { path: "/" };
  }
});
</script>
