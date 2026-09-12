<template>
  <NavBar
    :title="`${controller.camp ? Camps.label(controller.camp) : ''} - Communication`"
  >
    <v-tabs
      :model-value="currentTab"
      @update:model-value="(v) => setTab(v as DocumentsTab)"
    >
      <v-tab :value="'documents' satisfies DocumentsTab">Documents</v-tab>
      <v-tab :value="'lettre' satisfies DocumentsTab">
        <template #prepend>
          <v-icon>mdi-mail</v-icon>
        </template>
        Lettre aux familles</v-tab
      >
      <v-tab :value="'vetements' satisfies DocumentsTab">
        <template #prepend>
          <v-icon>mdi-washing-machine</v-icon>
        </template>
        Vêtements</v-tab
      >
      <v-tab :value="'forms' satisfies DocumentsTab">
        <template #prepend>
          <v-icon>mdi-format-list-checks</v-icon>
        </template>
        Formulaires</v-tab
      >
    </v-tabs>
  </NavBar>

  <v-tabs-window :model-value="currentTab">
    <v-tabs-window-item :value="'documents' satisfies DocumentsTab">
      <PannelDocuments @go-to="setTab"></PannelDocuments>
    </v-tabs-window-item>
    <v-tabs-window-item :value="'lettre' satisfies DocumentsTab">
      <PannelLettre></PannelLettre>
    </v-tabs-window-item>
    <v-tabs-window-item :value="'vetements' satisfies DocumentsTab">
      <PannelVetements></PannelVetements>
    </v-tabs-window-item>
    <v-tabs-window-item :value="'forms' satisfies DocumentsTab">
      <PannelForms></PannelForms>
    </v-tabs-window-item>
  </v-tabs-window>
</template>

<script lang="ts" setup>
import { useRouter } from "vue-router";
import NavBar from "../components/NavBar.vue";
import { computed } from "vue";
import {
  parseQueryURLDocuments,
  type DocumentsTab,
  type QueryURLDocuments,
} from "../plugins/router";
import { controller } from "../logic/logic";
import PannelLettre from "../components/documents/PannelLettre.vue";
import PannelDocuments from "../components/documents/PannelDocuments.vue";
import PannelVetements from "../components/documents/PannelVetements.vue";
import PannelForms from "../components/documents/PannelForms.vue";
import { Camps } from "@/utils.ts";

const router = useRouter();

const query = computed(() =>
  parseQueryURLDocuments(router.currentRoute.value.query),
);

const currentTab = computed<DocumentsTab>(() => query.value.tab || "documents");

function setTab(tab: DocumentsTab) {
  const current = router.currentRoute.value;
  router.push({
    path: current.path,
    query: { tab: tab } satisfies QueryURLDocuments,
  });
}
</script>
