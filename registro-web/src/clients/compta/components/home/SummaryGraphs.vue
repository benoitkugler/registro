<template>
  <v-col cols="12" md="2">
    <Bar
      :options="{
        plugins: { legend: { display: false } },
        aspectRatio: 0.6,
      }"
      :data="{
        labels: [
          `Recettes (${CurrencyLabels[camp.Recettes.Currency]})`,
          `Budget réel(${CurrencyLabels[camp.BudgetPrevisionnel.Reel.Currency]})`,
          `Budget simulé (${CurrencyLabels[camp.BudgetPrevisionnel.Simule.Currency]})`,
          `Dépenses (${CurrencyLabels[camp.DepensesTotal.Currency]})`,
        ],
        datasets: [
          {
            label: '',
            data: [
              camp.Recettes.Cent / 100,
              camp.BudgetPrevisionnel.Reel.Cent / 100,
              camp.BudgetPrevisionnel.Simule.Cent / 100,
              camp.DepensesTotal.Cent / 100,
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
        labels: ventilations.map((v) => VentilationLabels[v.ventilation]),
        datasets: [
          {
            label: `Répartition des dépenses (${CurrencyLabels[camp.DepenseGraph.Currency]})`,
            data: ventilations.map((v) => v.cent / 100),
          },
        ],
      }"
    ></Pie>
  </v-col>
</template>

<script lang="ts" setup>
import {
  Currency,
  CurrencyLabels,
  Ventilation,
  VentilationLabels,
  type CampDetails,
} from "../../logic/api";

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
import { computed } from "vue";

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

const props = defineProps<{
  camp: CampDetails;
}>();

// filter zeros
const ventilations = computed(() =>
  props.camp.DepenseGraph.Cents.map((cent, c) => ({
    ventilation: c as Ventilation,
    cent,
  })).filter((v) => v.cent != 0),
);
</script>
