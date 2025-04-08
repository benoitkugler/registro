// Composables
import { createRouter, createWebHistory, type LocationQuery } from "vue-router";
import Index from "../pages/index.vue";

const routes = [{ path: "/", component: Index }];

// TODO: check that
const baseUrl = import.meta.env.DEV
  ? "/src/clients/espaceperso"
  : "/espaceperso";

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

function enforceNumber<T extends number>(id: T | undefined) {
  return id ? (Number(id) as T) : undefined;
}

export type InscriptionsTab = "insc" | "participants";

export type QueryURLInscriptions = {
  tab?: InscriptionsTab;
  //   idDossier?: IdDossier;
};

export function parseQueryURLInscriptions(
  query: LocationQuery
): QueryURLInscriptions {
  const q = query as QueryURLInscriptions;
  return { tab: q.tab };
}

export default router;
