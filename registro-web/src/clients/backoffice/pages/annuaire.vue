<template>
  <NavBar title="Annuaire : personnes et organismes"> </NavBar>

  <v-card class="my-2 mx-auto" max-width="1000px">
    <template #append>
      <v-row>
        <v-col align-self="center">
          <DebounceField
            density="compact"
            width="250px"
            hide-details
            prepend-inner-icon="mdi-magnify"
            label="Recherche"
            v-model="search"
            @update:model-value="doSearch"
          ></DebounceField>
        </v-col>
        <v-col align-self="center" cols="auto">
          <v-btn color="green" @click="create" prepend-icon="mdi-plus"
            >Créer un nouveau profil</v-btn
          >
        </v-col>
      </v-row>
    </template>
    <v-card-text>
      <v-list>
        <v-list-item v-if="!list.length">
          <i v-if="!search.length"
            >Utiliser le champ de recherche pour afficher un profil.</i
          >
          <i v-else>Aucun profil ne correspond à votre recherche.</i>
        </v-list-item>
        <v-list-item
          v-for="personne in list"
          :title="personne.Label"
          :subtitle="Formatters.dateNaissance(personne.DateNaissance)"
          :prepend-icon="Formatters.sexeIcon(personne.Sexe)"
          @click="goToPersonne(personne.Id)"
        >
          <template #append v-if="personne.IsTemp">
            <v-chip prepend-icon="mdi-alert" color="warning">
              Ce profil est temporaire et devrait être identifié.
            </v-chip>
          </template>
        </v-list-item>
      </v-list>
    </v-card-text>

    <!-- edit dialog -->
    <v-dialog
      v-if="toEdit != null"
      :model-value="toEdit != null"
      @update:model-value="
        toEdit = null;
        goToPersonne(undefined);
      "
      max-width="1000px"
    >
      <PersonneEdit :personne="toEdit" @save="updatePersonne"></PersonneEdit>
    </v-dialog>
  </v-card>
</template>

<script lang="ts" setup>
import { computed, ref, watch } from "vue";
import NavBar from "../components/NavBar.vue";
import { useRouter } from "vue-router";
import { goToPersonne, parseQueryURLPersonnes } from "../plugins/router";
import type { IdPersonne, Personne, PersonneHeader } from "../logic/api";
import PersonneEdit from "../components/annuaire/PersonneEdit.vue";
import { controller } from "../logic/logic";
import { Formatters } from "@/utils";

const router = useRouter();

const queryURL = computed(() =>
  parseQueryURLPersonnes(router.currentRoute.value.query)
);

const currentId = computed(() => queryURL.value.idPersonne);

watch(
  () => currentId.value,
  () => (currentId.value !== undefined ? loadAndEdit(currentId.value) : null),
  { immediate: true }
);

const list = ref<PersonneHeader[]>([]);

const search = ref("");
async function doSearch() {
  if (!search.value) return;
  const res = await controller.PersonnesGet({ search: search.value });
  if (res === undefined) return;
  list.value = res || [];
}

async function create() {
  const res = await controller.PersonnesCreate();
  if (res === undefined) return;
  controller.showMessage("Profil créé avec succès.");

  list.value.push(res);
  goToPersonne(res.Id);
}

const toEdit = ref<Personne | null>(null);
async function updatePersonne(pr: Personne) {
  toEdit.value = null;
  const res = await controller.PersonnesUpdate(pr);
  if (res === undefined) return;
  controller.showMessage("Profil modifié avec succès.");
  goToPersonne();

  const index = list.value.findIndex((p) => p.Id == pr.Id);
  list.value[index] = res;
}

async function loadAndEdit(id: IdPersonne) {
  const res = await controller.PersonnesLoad({ id });
  if (res === undefined) return;
  toEdit.value = res;
}
</script>
