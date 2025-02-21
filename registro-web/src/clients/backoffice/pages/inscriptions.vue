<template>
  <NavBar title="Suivi des inscriptions">
    <div>Boutons a venir</div>
  </NavBar>

  <v-tabs grow v-model="tab">
    <v-tab value="insc">Nouvelles inscriptions</v-tab>
    <v-tab value="doss">Suivi des dossiers</v-tab>
  </v-tabs>
  <v-tabs-window v-model="tab">
    <v-tabs-window-item value="insc">
      <PannelInscriptions @go-to="goToDossier"></PannelInscriptions>
    </v-tabs-window-item>
    <v-tabs-window-item value="doss">
      <PannelDossiers></PannelDossiers>
    </v-tabs-window-item>
  </v-tabs-window>
</template>

<script lang="ts" setup>
import { ref } from "vue";
import NavBar from "../components/NavBar.vue";
import PannelInscriptions from "../components/inscriptions/PannelInscriptions.vue";
import PannelDossiers from "../components/inscriptions/PannelDossiers.vue";
import type { IdDossier } from "../logic/api";

const tab = ref<"insc" | "doss">("insc");

function goToDossier(id: IdDossier) {
  tab.value = "doss";
  // TODO: select the proper one
}
</script>
