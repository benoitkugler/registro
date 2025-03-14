<template>
  <NavBar title="Suivi des inscriptions">
    <v-tabs
      :model-value="currentTab"
      @update:model-value="v => setTab(v as InscriptionsTab)"
    >
      <v-tab value="insc">Nouvelles inscriptions</v-tab>
      <v-tab value="doss">Suivi des dossiers</v-tab>
    </v-tabs>
  </NavBar>

  <v-tabs-window :model-value="currentTab">
    <v-tabs-window-item value="insc">
      <PannelInscriptions @go-to="goToDossier"></PannelInscriptions>
    </v-tabs-window-item>
    <v-tabs-window-item value="doss">
      <PannelDossiers :initial-dossier="dossierToShow"></PannelDossiers>
    </v-tabs-window-item>
  </v-tabs-window>
</template>

<script lang="ts" setup>
import { computed } from "vue";
import NavBar from "../components/NavBar.vue";
import PannelInscriptions from "../components/inscriptions/PannelInscriptions.vue";
import PannelDossiers from "../components/inscriptions/PannelDossiers.vue";
import { useRouter } from "vue-router";
import {
  goToDossier,
  type InscriptionsTab,
  type QueryURLInscriptions,
  parseQueryURLInscriptions,
} from "../router";

const router = useRouter();

const query = computed(() =>
  parseQueryURLInscriptions(router.currentRoute.value.query)
);

const currentTab = computed(() => query.value.tab || "insc");

const dossierToShow = computed(() => query.value.idDossier);

function setTab(tab: InscriptionsTab) {
  const current = router.currentRoute.value;
  router.push({
    path: current.path,
    query: { tab: tab } satisfies QueryURLInscriptions,
  });
}
</script>
