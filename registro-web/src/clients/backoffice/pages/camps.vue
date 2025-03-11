<template>
  <NavBar title="Gestion des séjours">
    <v-btn v-if="current != null" @click="goBack">
      <template #prepend>
        <v-icon>mdi-view-list</v-icon>
      </template>
      Retour à la liste</v-btn
    >
  </NavBar>

  <CampsList v-if="current == null" @click="goTo"></CampsList>
  <CampParticipants :id="current" v-else></CampParticipants>
</template>

<script lang="ts" setup>
import { computed, ref } from "vue";
import NavBar from "../components/NavBar.vue";
import CampsList from "../components/camps/CampsList.vue";
import type { CampHeader, IdCamp } from "../logic/api";
import CampParticipants from "../components/camps/CampParticipants.vue";
import { useRouter } from "vue-router";

const router = useRouter();

const current = computed(() => {
  const id = router.currentRoute.value.query["idCamp"];
  return id ? (Number(id) as IdCamp) : null;
});

function goTo(camp: CampHeader) {
  const current = router.currentRoute.value;
  router.push({ path: current.path, query: { idCamp: camp.Camp.Camp.Id } });
}

function goBack() {
  const current = router.currentRoute.value;
  router.push({ path: current.path, query: {} });
}
</script>
