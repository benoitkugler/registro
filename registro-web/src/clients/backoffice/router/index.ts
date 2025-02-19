// Composables
import { createRouter, createWebHistory } from "vue-router";
import Index from "../pages/index.vue";
import Camps from "../pages/camps.vue";
import Inscriptions from "../pages/inscriptions.vue";

const routes = [
  { path: "/", component: Index },
  { path: "/camps", component: Camps },
  { path: "/inscriptions", component: Inscriptions },
];

// TODO: check that
const baseUrl = import.meta.env.DEV ? "/src/clients/backoffice" : "/backoffice";

const router = createRouter({
  history: createWebHistory(baseUrl),
  routes,
});

// Workaround for https://github.com/vitejs/vite/issues/11804
router.onError((err, to) => {
  if (err?.message?.includes?.("Failed to fetch dynamically imported module")) {
    if (!localStorage.getItem("vuetify:dynamic-reload")) {
      console.log("Reloading page to fix dynamic import error");
      localStorage.setItem("vuetify:dynamic-reload", "true");
      location.assign(to.fullPath);
    } else {
      console.error("Dynamic import error, reloading page did not fix it", err);
    }
  } else {
    console.error(err);
  }
});

router.isReady().then(() => {
  localStorage.removeItem("vuetify:dynamic-reload");
});

export default router;
