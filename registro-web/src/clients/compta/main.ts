/**
 * main.ts
 *
 * Bootstraps Vuetify and other plugins then mounts the App`
 */

// Plugins
import vuetify from "@/vuetify";
import router from "./plugins/router";

// Components
import App from "./App.vue";

// Composables
import { createApp } from "vue";

const app = createApp(App);

app.use(vuetify).use(router);

app.mount("#app");

if (import.meta.env.PROD && "serviceWorker" in navigator) {
  void navigator.serviceWorker
    .register("/static/sw.js", { scope: "/compta/" })
    .catch((error: unknown) =>
      console.error("Service worker registration failed", error),
    );
}
