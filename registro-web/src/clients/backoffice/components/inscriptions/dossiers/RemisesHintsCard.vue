<template>
  <v-card
    title="Estimer les remises"
    subtitle="Seuls les séjours de cette année et les nouvelles remises sont affichés."
  >
    <template #append>
      <v-btn :disabled="!selected.length" @click="applyHints"
        >Appliquer les remises</v-btn
      >
    </template>
    <v-card-text>
      <v-skeleton-loader v-if="data == null"></v-skeleton-loader>
      <v-list
        v-else
        select-strategy="leaf"
        v-model:selected="selected"
        color="primary"
      >
        <v-list-item v-if="!data.length" class="font-italic">
          Aucune nouvelle remise à appliquer.
        </v-list-item>
        <v-list-item
          v-for="(participant, index) in data"
          :title="participant.Personne"
          :subtitle="participant.Camp"
          :value="index"
          rounded
          class="my-1"
        >
          <template #append>
            <v-row style="width: 200px">
              <v-col cols="6">
                <v-chip
                  v-if="participant.Hint.ReducInscrits"
                  prepend-icon="mdi-account-multiple"
                  class="mx-1"
                  elevation="1"
                >
                  {{ participant.Hint.ReducInscrits }} %
                  <v-tooltip activator="parent" content-class="ma-0 pa-0">
                    <v-card subtitle="Autres inscrits de la même famille">
                      <v-card-text>
                        <v-chip
                          v-for="inscrit in participant.AutresInscrits"
                          class="mx-1"
                          >{{ inscrit.Personne }} ({{ inscrit.Camp }})</v-chip
                        >
                      </v-card-text>
                    </v-card>
                  </v-tooltip>
                </v-chip>
              </v-col>
              <v-col cols="6">
                <v-chip
                  v-if="participant.Hint.ReducEquipiers"
                  prepend-icon="mdi-account-hard-hat"
                  class="mx-1"
                  elevation="1"
                >
                  {{ participant.Hint.ReducEquipiers }} %
                  <v-tooltip activator="parent" content-class="ma-0 pa-0">
                    <v-card subtitle="Equipiers de la même famille">
                      <v-card-text>
                        <v-chip
                          v-for="equipier in participant.Equipiers"
                          class="mx-1"
                          >{{ equipier.Personne }} ({{ equipier.Camp }})</v-chip
                        >
                      </v-card-text>
                    </v-card>
                  </v-tooltip>
                </v-chip>
              </v-col>
            </v-row>
          </template>
        </v-list-item>
      </v-list>
    </v-card-text>
  </v-card>
</template>

<script setup lang="ts">
import type { RemisesHint } from "@/clients/backoffice/logic/api";
import { controller } from "@/clients/backoffice/logic/logic";
import { onMounted, ref } from "vue";

const props = defineProps<{}>();

onMounted(fetchHints);

const data = ref<RemisesHint[] | null>(null);
async function fetchHints() {
  const res = await controller.DossiersRemisesHint();
  if (res === undefined) return;
  data.value = res || [];
}

const selected = ref<number[]>([]); // indices in data
async function applyHints() {
  const list = data.value;
  if (!selected.value.length || !list) return;
  const arg = selected.value.map((index) => list[index]);
  const res = await controller.DossiersApplyRemisesHints(arg);
  selected.value = [];
  if (res === undefined) return;
  controller.showMessage("Remises appliquées avec succès.");
  fetchHints();
}
</script>
