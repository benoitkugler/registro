<template>
  <NavBar :title="`${controller.camp?.Label} - Communication & Documents`">
    <v-tabs
      :model-value="currentTab"
      @update:model-value="v => setTab(v as DocumentsTab)"
    >
      <v-tab :value="('documents' satisfies DocumentsTab)">Documents</v-tab>
      <v-tab :value="('lettre' satisfies DocumentsTab)"
        >Lettre aux familles</v-tab
      >
    </v-tabs>
  </NavBar>

  <v-tabs-window :model-value="currentTab">
    <v-tabs-window-item :value="('documents' satisfies DocumentsTab)">
      TODO
    </v-tabs-window-item>
    <v-tabs-window-item :value="('lettre' satisfies DocumentsTab)">
      <LettreEditor></LettreEditor>
    </v-tabs-window-item>
  </v-tabs-window>
</template>

<script lang="ts" setup>
import { useRouter } from "vue-router";
import NavBar from "../components/NavBar.vue";
import { computed, onMounted } from "vue";
import {
  parseQueryURLDocuments,
  type DocumentsTab,
  type QueryURLDocuments,
} from "../plugins/router";
import { controller } from "../logic/logic";
import LettreEditor from "../components/documents/LettreEditor.vue";

const router = useRouter();

const query = computed(() =>
  parseQueryURLDocuments(router.currentRoute.value.query)
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
