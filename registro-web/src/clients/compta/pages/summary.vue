<template>
  <v-responsive class="fill-height ma-2">
    <v-dialog v-model="showStateHelp" max-width="600">
      <v-card title="Explications" subtitle="Etat du dossier">
        <v-card-text>
          Le dossier de comptabilité peut se trouver dans 4 états.
          <v-row class="mt-2">
            <v-col cols="5" align-self="center"
              ><v-chip :color="StateColors[State.EnCours]">
                {{ StateLabels[State.EnCours] }}</v-chip
              >
            </v-col>
            <v-col>Le directeur est en train de déclarer les dépenses. </v-col>
          </v-row>
          <v-row>
            <v-col cols="5" align-self="center"
              ><v-chip :color="StateColors[State.ToBeValidated]">
                {{ StateLabels[State.ToBeValidated] }}</v-chip
              ></v-col
            >
            <v-col>
              Le directeur a fini et demande à la comptabilité de vérifier les
              enregistrements.
            </v-col>
          </v-row>
          <v-row>
            <v-col cols="5" align-self="center"
              ><v-chip :color="StateColors[State.Valide]">
                {{ StateLabels[State.Valide] }}</v-chip
              ></v-col
            >
            <v-col>
              Tous les enregistrements ont été validés. Le directeur peut encore
              modifier le dossier. La comptabilité peut archiver (clôturer) le
              dossier.
            </v-col>
          </v-row>
          <v-row>
            <v-col cols="5" align-self="center"
              ><v-chip :color="StateColors[State.Archive]">
                {{ StateLabels[State.Archive] }}</v-chip
              ></v-col
            >
            <v-col>
              Le dossier est complet et le directeur ne peut plus l'éditer.
            </v-col>
          </v-row>
        </v-card-text>
      </v-card>
    </v-dialog>
    <NoSelectedCamp v-if="controller.idCompta == -1"></NoSelectedCamp>

    <v-skeleton-loader type="card" v-else-if="camp == null"></v-skeleton-loader>
    <v-card
      v-else
      :title="camp.Label"
      :subtitle="`Numéro d'identification : ` + camp.Serial"
    >
      <v-card-text>
        <v-row class="my-1 mx-0">
          <v-col
            cols="12"
            md="8"
            align-self="center"
            class="bg-grey-lighten-4 rounded-lg"
          >
            <v-row>
              <v-col align-self="center" cols="12" md="4">
                Directeur : <b>{{ camp.Directeur }}</b></v-col
              >
              <v-col align-self="center" cols="12" md="3">
                Lieu :
                <v-tooltip
                  v-if="camp.AtMDR"
                  location="top"
                  text="Le bugdet prévisionnel et les dépenses sont ajustés pour un séjour à la Maison du Rocher."
                >
                  <template v-slot:activator="{ isActive, props }">
                    <v-chip v-on="{ isActive }" v-bind="props" color="primary"
                      >Maison du Rocher</v-chip
                    >
                  </template>
                </v-tooltip>
                <b v-else>{{ camp.Lieu }}</b>
              </v-col>
              <v-col align-self="center" cols="12" md="5" class="text-right">
                du <b>{{ Formatters.date(camp.DateDebut) }}</b> au
                <b>{{ Formatters.date(camp.DateFin) }}</b> :
                <b>{{ camp.Duree }}</b> jours</v-col
              >
            </v-row>
            <v-row>
              <v-col cols="12" md="6">
                <v-row no-gutters>
                  <v-col>Prix : {{ camp.Prix }}€</v-col>
                  <v-col>Nombre d'inscrits : {{ camp.Inscrits }}</v-col></v-row
                >
                <v-row no-gutters>
                  <v-col>Nombre total d'équipiers : {{ camp.Equipiers }}</v-col>
                  <v-col>Nombre d'animateurs : {{ camp.Animateurs }}</v-col>
                </v-row>
              </v-col>
              <v-col cols="12" md="6">
                <v-select
                  variant="underlined"
                  density="comfortable"
                  hide-details
                  label="Compte banquaire associé"
                  :items="accountItems"
                  :model-value="camp.Compta.Account"
                  @update:model-value="updateAccount"
                ></v-select>
              </v-col>
            </v-row>
          </v-col>

          <v-col align-self="center">
            <v-card title="État du dossier">
              <template v-slot:append>
                <v-btn
                  icon="mdi-help"
                  size="x-small"
                  @click="showStateHelp = true"
                ></v-btn>
              </template>
              <v-card-text class="text-center pb-0">
                <v-chip
                  :color="StateColors[camp.Compta.State]"
                  :append-icon="StateIcons[camp.Compta.State]"
                  >{{ StateLabels[camp.Compta.State] }}</v-chip
                >
              </v-card-text>
              <v-card-actions>
                <v-spacer></v-spacer>
                <v-btn
                  :disabled="
                    !controller.isAdmin ||
                    camp.Compta.State == State.EnCours ||
                    camp.Compta.State == State.ToBeValidated
                  "
                  @click="toogleArchive()"
                  >{{
                    camp.Compta.State == State.Archive
                      ? "Déverouiller"
                      : "Archiver"
                  }}</v-btn
                >
              </v-card-actions>
            </v-card>
          </v-col>
        </v-row>
        <v-row class="mt-1 mx-0" justify="space-between">
          <v-col
            cols="12"
            md="5"
            align-self="center"
            class="bg-pink-lighten-5 rounded-lg"
          >
            <v-row justify="center">
              <v-col cols="auto" align-self="center">
                <v-chip link class="mx-0" elevation="1">
                  <!-- TODO: :to="{ path: '/budget' }" -->
                  Budget prévisionnel :
                  {{ camp.BudgetPrevisionnelReel.toFixed(2) }}€ (réel) /
                  {{ camp.BudgetPrevisionnelSimule.toFixed(2) }}€ (simulé)
                </v-chip>
              </v-col>
            </v-row>

            <v-row justify="center">
              <v-col cols="auto">
                <v-chip
                  link
                  :to="{ name: 'records' }"
                  class="mx-0"
                  elevation="1"
                >
                  <template v-slot:append>
                    <v-tooltip
                      text="Somme des dépenses enregistrées, les remboursements étant déduits"
                    >
                      <template v-slot:activator="{ isActive, props }">
                        <v-icon v-on="{ isActive }" v-bind="props" class="ml-2"
                          >mdi-help-circle-outline</v-icon
                        >
                      </template>
                    </v-tooltip>
                  </template>
                  Dépenses : {{ camp.DepenseTotal.toFixed(2) }}€
                  <span class="d-none d-sm-inline ml-1">
                    (compte dédié : {{ camp.DepenseOwnAccount.toFixed(2) }}€)
                  </span>
                </v-chip>
              </v-col>
            </v-row>

            <v-row justify="space-between">
              <v-col cols="auto">
                <v-tooltip
                  text="Recettes attendues provenant des pensions (factures et aides)"
                >
                  <template v-slot:activator="{ isActive, props }">
                    <div v-on="{ isActive }" v-bind="props">
                      Budget total: <b>{{ camp.Budget.toFixed(2) }}€</b>
                    </div>
                  </template>
                </v-tooltip>
              </v-col>
              <v-col cols="auto">
                Réductions : {{ camp.PercentageRemises.toFixed(1) }}%
              </v-col>
              <v-col cols="auto">
                Sans réductions : {{ camp.BudgetSansRemises.toFixed(2) }}€
              </v-col>
            </v-row>
            <v-row>
              <v-col align-self="center" cols="8" md="3"
                >Contribution aux frais fixes:
              </v-col>
              <v-col align-self="center" cols="4" md="auto"
                >{{ camp.Marge.toFixed(2) }}%</v-col
              >
              <v-spacer></v-spacer>
              <v-col align-self="center" cols="8" md="4"
                >Dépense moyenne par jour et par inscrit :</v-col
              >
              <v-col align-self="center" cols="4" md="auto">
                {{ camp.DepenseJourneeInscrit.toFixed(2) }}€</v-col
              >
            </v-row>
          </v-col>
          <v-col cols="12" md="2">
            <Bar
              :options="{
                plugins: { legend: { display: false } },
                aspectRatio: 0.6,
              }"
              :data="{
                labels: [
                  'Recettes (€)',
                  'Budget réel(€)',
                  'Budget simulé (€)',
                  'Dépenses (€)',
                ],
                datasets: [
                  {
                    label: '',
                    data: [
                      camp.Budget,
                      camp.BudgetPrevisionnelReel,
                      camp.BudgetPrevisionnelSimule,
                      camp.DepenseTotal,
                    ],
                    backgroundColor: ['teal', 'lightblue', 'lightblue', 'pink'],
                  },
                ],
              }"
            ></Bar>
          </v-col>
          <v-col cols="12" md="4">
            <Pie
              :options="{
                plugins: {
                  legend: { display: true },
                  autocolors: {
                    mode: 'data',
                  },
                },
                aspectRatio: 1.4,
              }"
              :data="{
                labels: (camp.DepenseGraph || []).map(
                  (v) => VentilationLabels[v.Ventilation],
                ),
                datasets: [
                  {
                    label: 'Répartition des dépenses (€)',
                    data: (camp.DepenseGraph || []).map((v) => v.Montant),
                  },
                ],
              }"
            ></Pie>
          </v-col>
        </v-row>
      </v-card-text>
    </v-card>
  </v-responsive>
</template>

<script lang="ts" setup>
import { onMounted, ref } from "vue";
import { Bar, Pie } from "vue-chartjs";

import autocolors from "chartjs-plugin-autocolors";

import {
  Chart as ChartJS,
  Title,
  Tooltip,
  Legend,
  BarElement,
  ArcElement,
  CategoryScale,
  LinearScale,
} from "chart.js";
import { controller, StateColors, StateIcons } from "../logic/logic";
import {
  AccountACVE,
  AccountACVELabels,
  State,
  StateLabels,
  VentilationLabels,
  type CampDetails,
} from "../logic/api";
import { Formatters, selectItems } from "@/utils";
import NoSelectedCamp from "../components/NoSelectedCamp.vue";

ChartJS.register(
  Title,
  Tooltip,
  Legend,
  BarElement,
  ArcElement,
  CategoryScale,
  LinearScale,
  autocolors,
);

const camp = ref<CampDetails | null>(null);
const showStateHelp = ref(false);

onMounted(fetch);

async function fetch() {
  if (controller.idCompta == -1) return;
  const res = await controller.GetCampDetails({
    IdCompta: controller.idCompta,
  });
  if (res === undefined) return;
  camp.value = res;
}

const accountItems = selectItems(AccountACVELabels);

async function updateAccount(account: AccountACVE) {
  if (camp.value == null) return;
  camp.value.Compta.Account = account;
  const res = await controller.UpdateCompta(camp.value.Compta);
  if (res === undefined) return;
  controller.showMessage("Compte principal mis à jour avec succès.");
}

// Archive -> ToBeValidated or Valide->Archive
async function toogleArchive() {
  if (camp.value == null) return;
  const current = camp.value.Compta.State;
  let state: State;
  let message: string;
  if (current == State.Archive) {
    state = State.ToBeValidated;
    message = "Dossier déverrouillé avec succès.";
  } else if (current == State.Valide) {
    state = State.Archive;
    message = "Dossier archivé avec succès.";
  } else {
    return;
  }
  const res = await controller.SetComptaState({
    IdCompta: camp.value.Compta.Id,
    State: state,
    SendMail: false,
  });
  if (res === undefined) return;

  controller.showMessage(message);
  camp.value.Compta = res;
}
</script>
