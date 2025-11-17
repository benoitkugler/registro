<template>
  <NavBar title="Gestion des séjours">
    <v-btn v-if="current !== null" @click="goBack" prepend-icon="mdi-view-list">
      Retour à la liste</v-btn
    >
  </NavBar>

  <CampsList v-if="current === undefined" @show-participants="goTo"></CampsList>
  <CampParticipants
    :id="current"
    :id-participant="queryURL.idParticipant"
    v-else
  ></CampParticipants>
</template>

<script lang="ts" setup>
import { computed } from "vue";
import NavBar from "../components/NavBar.vue";
import CampsList from "../components/camps/CampsList.vue";
import type { CampHeader } from "../logic/api";
import CampParticipants from "../components/camps/CampParticipants.vue";
import { useRouter } from "vue-router";
import { parseQueryURLCamps, type QueryURLCamps } from "../plugins/router";

const router = useRouter();

const queryURL = computed(() =>
  parseQueryURLCamps(router.currentRoute.value.query)
);

const current = computed(() => queryURL.value.idCamp);

function goTo(camp: CampHeader) {
  const current = router.currentRoute.value;
  router.push({ path: current.path, query: { idCamp: camp.Camp.Camp.Id } });
}

function goBack() {
  const current = router.currentRoute.value;
  router.push({ path: current.path, query: {} });
}
</script>
