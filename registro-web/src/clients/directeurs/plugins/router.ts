// Composables
import { createRouter, createWebHistory, type LocationQuery } from "vue-router";
import Index from "../pages/index.vue";
import Inscriptions from "../pages/inscriptions.vue";
import Equipiers from "../pages/equipiers.vue";
import Documents from "../pages/documents.vue";

const routes = [
  { path: "/", component: Index },
  { path: "/inscriptions", component: Inscriptions },
  { path: "/equipiers", component: Equipiers },
  { path: "/documents", component: Documents },
];

// TODO: check that
const baseURL = import.meta.env.DEV ? "/src/clients/directeurs" : "/directeurs";

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

export type DocumentsTab = "documents" | "lettre" | "vetements";

export type QueryURLDocuments = {
  tab?: DocumentsTab;
};

export function parseQueryURLDocuments(
  query: LocationQuery
): QueryURLDocuments {
  const q = query as QueryURLDocuments;
  return { tab: q.tab };
}

export default router;
