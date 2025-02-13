/**
 * main.ts
 *
 * Bootstraps Vuetify and other plugins then mounts the App`
 */

// Components
import App from "./App.vue";

// Plugins
import vuetify from "./plugins/vuetify";

// Composables
import { createApp } from "vue";

const app = createApp(App);

app.use(vuetify);

app.mount("#app");
