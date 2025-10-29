// Composables
import { createRouter, createWebHistory, type LocationQuery } from "vue-router";
import Index from "../pages/index.vue";
import Services from "../pages/services.vue";
import Cgu from "../pages/cgu.vue";

const routes = [
  { path: "/", component: Index },
  { path: "/services", component: Services },
  { path: "/cgu", component: Cgu },
];

const baseURL = import.meta.env.DEV ? "/src/clients/services" : "/";

const router = createRouter({
  history: createWebHistory(baseURL),
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
