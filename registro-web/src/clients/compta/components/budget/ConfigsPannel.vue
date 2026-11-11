<template>
  <v-card
    title="Paramètres de calcul du budget"
    subtitle="Au maximum une configuration par année peut être définie."
  >
    <v-dialog
      :model-value="toDelete != null"
      @update:model-value="toDelete = null"
      max-width="600"
    >
      <v-card v-if="toDelete" title="Confirmer la suppression">
        <v-card-text>
          Confirmez-vous la suppression de la configuration pour l'année
          {{ toDelete.Year }} ? <br />
          Cette opération est irréversible.
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="red" @click="deleteConfig">Supprimer</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-card-text>
      <v-row>
        <v-col cols="3">
          <v-card subtitle="Années">
            <template v-slot:append>
              <v-btn icon size="x-small" @click="createConfig"
                ><v-icon color="green">mdi-plus</v-icon></v-btn
              >
            </template>
            <v-card-text>
              <v-tabs direction="vertical" v-model="currentConfig">
                <v-row
                  v-for="(config, i) in budgetConfigs"
                  :key="config.Year"
                  justify="space-between"
                >
                  <v-col align-self="center" cols="8">
                    <v-tab :value="i">
                      {{ config.Year }}
                    </v-tab>
                  </v-col>
                  <v-col align-self="center" cols="auto">
                    <v-btn
                      icon
                      size="x-small"
                      class="ma-1"
                      @click="toDelete = config"
                      ><v-icon color="red">mdi-delete</v-icon></v-btn
                    >
                  </v-col>
                </v-row>
              </v-tabs>
            </v-card-text>
          </v-card>
        </v-col>
        <v-col>
          <div
            v-if="!budgetConfigs.length"
            class="font-italic text-center my-4"
          >
            Aucune configuration n'est encore définie.
          </div>
          <v-window v-model="currentConfig">
            <v-window-item
              v-for="(config, i) in budgetConfigs"
              :key="config.Year"
              :value="i"
            >
              <ComptaConfigDetails
                :config="config"
                @update="updateConfig"
              ></ComptaConfigDetails>
            </v-window-item>
          </v-window>
        </v-col>
      </v-row>
    </v-card-text>
  </v-card>
</template>

<script lang="ts" setup>
import { onMounted, ref } from "vue";
import ComptaConfigDetails from "./ComptaConfigDetails.vue";
import { controller } from "../../logic/logic.ts";
import type { ComptaConfig } from "../../logic/api.ts";

// const props = defineProps<{
//   config: ComptaConfig;
// }>();

// const emit = defineEmits<{
//   (e: "update"): void;
// }>();

onMounted(fetchConfigs);

const currentConfig = ref(0);
const budgetConfigs = ref<ComptaConfig[]>([]);
async function fetchConfigs() {
  const res = await controller.GetComptaConfigs();
  if (res === undefined) return;
  budgetConfigs.value = res || [];
}

async function createConfig() {
  const res = await controller.CreateComptaConfig();
  if (res === undefined) return;
  controller.showMessage("Configuration ajoutée avec succès.");
  fetchConfigs();
}

const toDelete = ref<ComptaConfig | null>(null);
async function deleteConfig() {
  if (toDelete.value == null) return;
  const res = await controller.DeleteComptaConfig({ id: toDelete.value.Id });
  if (res === undefined) return;
  controller.showMessage("Configuration supprimée avec succès.");
  budgetConfigs.value = budgetConfigs.value.filter(
    (c) => c.Id != toDelete.value?.Id,
  );
  toDelete.value = null;
}

async function updateConfig(config: ComptaConfig) {
  const res = await controller.UpdateComptaConfig(config);
  if (res === undefined) return;
  controller.showMessage("Configuration mise à jour avec succès.");
  fetchConfigs();
}
</script>
