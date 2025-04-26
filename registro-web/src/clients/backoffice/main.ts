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
