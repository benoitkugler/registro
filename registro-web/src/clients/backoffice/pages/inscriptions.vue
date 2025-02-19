<template>
  <NavBar title="Suivi des inscriptions">
    <div>Boutons a venir</div>
  </NavBar>

  <v-tabs grow v-model="tab">
    <v-tab value="1">Nouvelles inscriptions</v-tab>
    <v-tab value="2">Suivi des dossiers</v-tab>
  </v-tabs>
  <v-tabs-window v-model="tab">
    <v-tabs-window-item value="1">
      <PannelInscriptions></PannelInscriptions>
    </v-tabs-window-item>
    <v-tabs-window-item value="2">
      <PannelDossiers></PannelDossiers>
    </v-tabs-window-item>
  </v-tabs-window>
</template>

<script lang="ts" setup>
import { ref } from "vue";
import NavBar from "../components/NavBar.vue";
import PannelInscriptions from "../components/dossiers/PannelInscriptions.vue";
import PannelDossiers from "../components/dossiers/PannelDossiers.vue";

const tab = ref("1");
</script>
