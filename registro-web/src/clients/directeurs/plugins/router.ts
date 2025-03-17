// Composables
import { createRouter, createWebHistory, type LocationQuery } from "vue-router";
import Index from "../pages/index.vue";
import type { IdCamp, IdParticipant } from "../logic/api";

const routes = [{ path: "/", component: Index }];

// TODO: check that
const baseUrl = import.meta.env.DEV ? "/src/clients/directeurs" : "/directeurs";

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

export type QueryURLCamps = {
  idCamp?: IdCamp;
  idParticipant?: IdParticipant;
};

export function parseQueryURLCamps(query: LocationQuery): QueryURLCamps {
  const q = query as QueryURLCamps;
  return {
    idCamp: enforceNumber(q.idCamp),
    idParticipant: enforceNumber(q.idParticipant),
  };
}

export function goToParticipant(participant: {
  IdCamp: IdCamp;
  Id: IdParticipant;
}) {
  router.push({
    path: "/camps",
    query: {
      idCamp: participant.IdCamp,
      idParticipant: participant.Id,
    } satisfies QueryURLCamps,
  });
}

export default router;
