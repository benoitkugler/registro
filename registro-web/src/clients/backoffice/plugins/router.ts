// Composables
import { createRouter, createWebHistory, type LocationQuery } from "vue-router";
import Index from "../pages/index.vue";
import Camps from "../pages/camps.vue";
import Inscriptions from "../pages/inscriptions.vue";
import type {
  IdCamp,
  IdDossier,
  IdParticipant,
  IdPersonne,
} from "../logic/api";
import Annuaire from "../pages/annuaire.vue";

const routes = [
  { path: "/", component: Index },
  { path: "/camps", component: Camps },
  { path: "/inscriptions", component: Inscriptions },
  { path: "/annuaire", component: Annuaire },
];

// TODO: check that
const baseURL = import.meta.env.DEV ? "/src/clients/backoffice" : "/backoffice";

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

export type InscriptionsTab = "insc" | "doss";

export type QueryURLInscriptions = {
  tab?: InscriptionsTab;
  idDossier?: IdDossier;
};

export function parseQueryURLInscriptions(
  query: LocationQuery
): QueryURLInscriptions {
  const q = query as QueryURLInscriptions;
  return { tab: q.tab, idDossier: enforceNumber(q.idDossier) };
}

export function goToDossier(idDossier?: IdDossier) {
  router.push({
    path: "/inscriptions",
    query: { tab: "doss", idDossier } satisfies QueryURLInscriptions,
  });
}

export type QueryURLPersonnes = {
  idPersonne?: IdPersonne;
};

export function parseQueryURLPersonnes(
  query: LocationQuery
): QueryURLPersonnes {
  const q = query as QueryURLPersonnes;
  return {
    idPersonne: enforceNumber(q.idPersonne),
  };
}

export function goToPersonne(idPersonne?: IdPersonne) {
  router.push({
    path: "/annuaire",
    query: { idPersonne } satisfies QueryURLPersonnes,
  });
}

export default router;
