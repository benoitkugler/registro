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
import { CachedTokens, controller, isDirectorLoggin } from "./logic/logic";
import { useRouter } from "vue-router";

const message = ref("");
const messageColor = ref("primary");

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

// if we are not yet logged in,
// redirect to the index page
const router = useRouter();
router.beforeEach(async (to, from) => {
  // resolve loggin strategy :
  //  - give priority to director link (query param 'key')
  //  - otherwise, try to use a cached token
  //  - fallback to admin loggin screen

  // only trigger this path at application startup (including reload)
  if (controller.authToken) return;

  // a valid director key logs the user in and sets the current compta
  const directorKey = isDirectorLoggin(to);
  if (directorKey) {
    const res = await controller.Loggin({ key: directorKey, isAdmin: false });
    if (res?.IsValid) {
      controller.onSuccessfullLoggin(res.IdCompta, res.Token);
      // the index page would log in again : go straight to home
      if (to.path == "/") return { path: "/home" };
      return; // stay on the target route
    }
    // invalid director key : fallback to a cached token (if any)
    // or to the loggin page below
  }

  const { idCompta, token } = CachedTokens.get();
  if (token) {
    // check the token is still valid
    const res = await controller.Loggin({
      key: token,
      isAdmin: idCompta == -1,
    });
    if (res && res.IsValid) {
      controller.onSuccessfullLoggin(res.IdCompta, res.Token);
      if (to.path == "/") {
        // we are on the loggin page, dont stay here !
        return { path: "/home" };
      } else {
        return; // do not redirect to login page
      }
    } else {
      CachedTokens.clear();
    }
  }

  if (to.path !== "/") {
    // redirect the user to the login page
    return { path: "/" };
  }
});
</script>
