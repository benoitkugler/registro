// Composables
import {
  createRouter,
  createWebHistory,
  type LocationQuery,
  type RouteLocation,
} from "vue-router";
import Index from "../pages/index.vue";

const routes = [{ path: "/", component: Index }];

// TODO: check that
const baseURL = import.meta.env.DEV
  ? "/src/clients/espaceperso"
  : "/espaceperso";

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

function hasQueryParams(route: RouteLocation) {
  return !!Object.keys(route.query).length;
}

router.beforeEach((to, from, next) => {
  if (!hasQueryParams(to) && hasQueryParams(from)) {
    next({ name: to.name, query: from.query });
  } else {
    next();
  }
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
