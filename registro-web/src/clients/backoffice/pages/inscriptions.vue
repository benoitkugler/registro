<template>
  <NavBar title="Suivi des inscriptions">
    <v-tabs :model-value="tab" @update:model-value="v => setTab(v as tabValue)">
      <v-tab value="insc">Nouvelles inscriptions</v-tab>
      <v-tab value="doss">Suivi des dossiers</v-tab>
    </v-tabs>
  </NavBar>

  <v-tabs-window :model-value="tab">
    <v-tabs-window-item value="insc">
      <PannelInscriptions @go-to="goToDossier"></PannelInscriptions>
    </v-tabs-window-item>
    <v-tabs-window-item value="doss">
      <PannelDossiers ref="dossiers"></PannelDossiers>
    </v-tabs-window-item>
  </v-tabs-window>
</template>

<script lang="ts" setup>
import { computed, nextTick, ref, useTemplateRef } from "vue";
import NavBar from "../components/NavBar.vue";
import PannelInscriptions from "../components/inscriptions/PannelInscriptions.vue";
import PannelDossiers from "../components/inscriptions/PannelDossiers.vue";
import type { IdDossier, Personne } from "../logic/api";
import { useRouter } from "vue-router";

type tabValue = "insc" | "doss";
const dossiersPannel = useTemplateRef("dossiers");

const router = useRouter();

const tab = computed(
  () => (router.currentRoute.value.query["tab"] || "insc") as tabValue
);

function goToDossier(id: IdDossier, responsable: Personne) {
  setTab("doss");
  nextTick(() => {
    dossiersPannel.value?.showDossier(
      id,
      `${responsable.Prenom} ${responsable.Nom}`
    );
  });
}

function setTab(tab: tabValue) {
  const current = router.currentRoute.value;
  router.push({ path: current.path, query: { tab: tab } });
}
</script>
