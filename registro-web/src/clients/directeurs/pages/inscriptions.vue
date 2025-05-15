<template>
  <NavBar :title="`${controller.camp?.Label} - Suivi des inscriptions`">
    <v-tabs
      :model-value="currentTab"
      @update:model-value="v => setTab(v as InscriptionsTab)"
    >
      <v-tab value="insc">Nouvelles inscriptions</v-tab>
      <v-tab value="participants">Liste des participants</v-tab>
    </v-tabs>
  </NavBar>

  <v-tabs-window :model-value="currentTab">
    <v-tabs-window-item value="insc">
      <PannelInscriptions
        @go-to="() => setTab('participants')"
      ></PannelInscriptions>
    </v-tabs-window-item>
    <v-tabs-window-item value="participants">
      <PannelParticipants ref="participants"></PannelParticipants>
    </v-tabs-window-item>
  </v-tabs-window>
</template>

<script lang="ts" setup>
import { useRouter } from "vue-router";
import NavBar from "../components/NavBar.vue";
import { computed, onMounted, useTemplateRef } from "vue";
import {
  parseQueryURLInscriptions,
  type InscriptionsTab,
  type QueryURLInscriptions,
} from "../plugins/router";
import PannelInscriptions from "../components/inscriptions/PannelInscriptions.vue";
import { controller } from "../logic/logic";
import PannelParticipants from "../components/inscriptions/PannelParticipants.vue";

const router = useRouter();

const query = computed(() =>
  parseQueryURLInscriptions(router.currentRoute.value.query)
);

onMounted(async () => {
  // go to participants if inscriptions is empty
  const res = await controller.InscriptionsGet();
  if (res === undefined) return;
  if (!res?.length) setTab("participants");
});

const currentTab = computed(() => query.value.tab || "insc");

const participants = useTemplateRef("participants");

function setTab(tab: InscriptionsTab) {
  const current = router.currentRoute.value;
  router.push({
    path: current.path,
    query: { tab: tab } satisfies QueryURLInscriptions,
  });
  if (tab == "participants") participants.value?.loadParticipants();
}
</script>
