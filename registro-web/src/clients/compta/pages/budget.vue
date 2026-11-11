<template>
  <NavBar></NavBar>

  <v-dialog v-model="showConfigs" max-width="1200">
    <ConfigsPannel></ConfigsPannel>
  </v-dialog>

  <v-responsive class="fill-height ma-2">
    <NoSelectedCamp v-if="controller.idCompta == -1"></NoSelectedCamp>
    <v-skeleton-loader type="card" v-else-if="data == null"></v-skeleton-loader>
    <!-- existing budget -->
    <v-card
      v-else-if="data.IsValid"
      title="Budget prévisionnel"
      :subtitle="subtitle"
    >
      <template v-slot:append>
        <v-btn-toggle mandatory v-model="simulated">
          <v-btn :value="false" prepend-icon="mdi-database-check"
            >Réel
            <v-tooltip activator="parent">
              Le nombre d'inscrits et d'équipiers est calculé depuis la base de
              données.
            </v-tooltip>
          </v-btn>
          <v-btn :value="true" append-icon="mdi-robot-industrial">
            Simulé
            <v-tooltip activator="parent">
              Le nombre d'inscrits et d'équipiers est à renseigner.
            </v-tooltip>
          </v-btn>
        </v-btn-toggle>
        <v-divider vertical thickness="1" class="mx-2"></v-divider>
        <v-btn flat :disabled="!controller.isAdmin" @click="showConfigs = true">
          <template v-slot:prepend><v-icon>mdi-cog</v-icon></template>
          Réglages</v-btn
        >
      </template>
      <v-card-text v-if="budget && camp">
        <BudgetPannel
          :camp="camp"
          :origin="data.Origin"
          :resolved="budget"
          :simulated="simulated"
          :config="data.Config"
          @update="update"
        ></BudgetPannel>
      </v-card-text>
    </v-card>
    <!-- no budget so far -->
    <NewBudget :camp="camp" v-else-if="camp" @create="createBudget"></NewBudget>
  </v-responsive>
</template>

<script lang="ts" setup>
import { computed, onMounted, ref } from "vue";
import type {
  CampHeader,
  ComptaBudget,
  EstimatedBudget,
  IdComptaCamp,
} from "../logic/api";
import { controller } from "../logic/logic";
import { Camps, nullableToZeroable } from "@/utils";
import NoSelectedCamp from "../components/NoSelectedCamp.vue";
import ConfigsPannel from "../components/budget/ConfigsPannel.vue";
import BudgetPannel from "../components/budget/BudgetPannel.vue";
import NewBudget from "../components/budget/NewBudget.vue";
import NavBar from "../components/NavBar.vue";

onMounted(fetchBudget);

const showConfigs = ref(false);

const subtitle = computed(() =>
  camp.value == null ? "" : `${Camps.label(camp.value)}`,
);
const simulated = ref(true);
const budget = computed(() =>
  simulated.value ? data.value?.Simulated : data.value?.Real,
);
const data = ref<EstimatedBudget | null>(null);
const camp = ref<CampHeader | null>(null);
async function fetchBudget() {
  const header = await controller.GetCompta({
    IdCompta: controller.idCompta,
  });
  if (header === undefined) return;
  const res = await controller.GetEstimatedBudget({
    "id-compta": controller.idCompta,
  });
  if (res === undefined) return;
  data.value = res;
  camp.value = header;
}

async function update(origin: ComptaBudget) {
  if (data.value == null) return;
  const res = await controller.UpdateEstimatedBudget(origin);
  if (res === undefined) return;
  controller.showMessage("Modifications enregistrées.");
  data.value = res;
}

async function createBudget(from: IdComptaCamp | null) {
  const res = await controller.CreateBudget({
    IdCompta: controller.idCompta,
    ImportFrom: nullableToZeroable(from),
  });
  if (res === undefined) return;
  controller.showMessage("Budget créé avec succès");
  data.value = res;
}
</script>
