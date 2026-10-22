<template>
  <NavBar
    :title="
      smAndUp
        ? `${controller.camp ? Camps.label(controller.camp) : ''} - Suivi des inscriptions`
        : `Inscriptions`
    "
  >
    <v-tabs
      :model-value="currentTab"
      @update:model-value="(v) => setTab(v as InscriptionsTab)"
      :stacked="!smAndUp"
    >
      <v-tab value="inscriptions">
        <v-icon icon="mdi-clock"></v-icon>
        {{ smAndUp ? "Inscriptions en attente" : "" }}
      </v-tab>
      <v-tab value="participants">
        <v-icon icon="mdi-view-list"></v-icon>
        {{ smAndUp ? "Liste des participants" : "" }}
      </v-tab>
    </v-tabs>
  </NavBar>

  <v-tabs-window :model-value="currentTab">
    <v-tabs-window-item value="inscriptions">
      <PannelInscriptions
        @go-to="() => setTab('participants')"
        ref="inscriptions"
      ></PannelInscriptions>
    </v-tabs-window-item>
    <v-tabs-window-item value="participants">
      <PannelParticipants
        ref="participants"
        @go-to-inscription="goToInscription"
      ></PannelParticipants>
    </v-tabs-window-item>
  </v-tabs-window>
</template>

<script lang="ts" setup>
import { useRouter } from "vue-router";
import NavBar from "../components/NavBar.vue";
import { computed, nextTick, onMounted, useTemplateRef } from "vue";
import {
  parseQueryURLInscriptions,
  type InscriptionsTab,
  type QueryURLInscriptions,
} from "../plugins/router";
import PannelInscriptions from "../components/inscriptions/PannelInscriptions.vue";
import { controller } from "../logic/logic";
import PannelParticipants from "../components/inscriptions/PannelParticipants.vue";
import type { IdDossier } from "../logic/api";
import { Camps } from "@/utils.ts";
import { useDisplay } from "vuetify";

const router = useRouter();
const { smAndUp } = useDisplay();

const query = computed(() =>
  parseQueryURLInscriptions(router.currentRoute.value.query),
);

onMounted(async () => {
  // go to participants if inscriptions is empty
  const res = await controller.InscriptionsGet();
  if (res === undefined) return;
  if (!res.Inscriptions?.length) setTab("participants");
});

const currentTab = computed(() => query.value.tab || "inscriptions");

const inscriptions = useTemplateRef("inscriptions");
const participants = useTemplateRef("participants");

function setTab(tab: InscriptionsTab) {
  const current = router.currentRoute.value;
  router.push({
    path: current.path,
    query: { tab: tab } satisfies QueryURLInscriptions,
  });
  if (tab == "participants") participants.value?.loadParticipants();
}

function goToInscription(id: IdDossier) {
  setTab("inscriptions");
  nextTick(() => {
    inscriptions.value?.goToInscription(id);
  });
}
</script>
