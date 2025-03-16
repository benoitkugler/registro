<template>
  <NavBar title="Annuaire" subtitle="Personnes et organismes"> TODO </NavBar>

  <v-card>
    <v-dialog
      v-if="toEdit != null"
      :model-value="toEdit != null"
      @update:model-value="toEdit = null"
      max-width="800px"
    >
      <PersonneEdit :personne="toEdit" @save="updatePersonne"></PersonneEdit>
    </v-dialog>
  </v-card>
</template>

<script lang="ts" setup>
import { computed, ref, watch } from "vue";
import NavBar from "../components/NavBar.vue";
import { useRouter } from "vue-router";
import { parseQueryURLPersonnes } from "../plugins/router";
import type { IdPersonne, Personne } from "../logic/api";
import PersonneEdit from "../components/annuaire/PersonneEdit.vue";
import { controller } from "../logic/logic";

const router = useRouter();

const queryURL = computed(() =>
  parseQueryURLPersonnes(router.currentRoute.value.query)
);

const current = computed(() => queryURL.value.idPersonne);

watch(
  () => current.value,
  () => (current.value !== undefined ? loadAndEdit(current.value) : null)
);

const toEdit = ref<Personne | null>(null);

async function updatePersonne() {
  if (toEdit.value == null) return;
  // TODO
  toEdit.value = null;
  controller.showMessage("Profil modifié avec succès.");
}

async function loadAndEdit(id: IdPersonne) {
  // TODO
}
</script>
