<template>
  <v-tooltip location="top" :text="VentilationLabels[props.montant.Kind]">
    <template v-slot:activator="{ props: innerProps }">
      <v-chip v-bind="innerProps" :prepend-icon="icons[props.montant.Kind]">
        {{
          Formatters.montant({
            Currency: props.currency,
            Cent: props.montant.Cent,
          })
        }}
      </v-chip>
    </template>
  </v-tooltip>
</template>

<script lang="ts" setup>
import { Formatters } from "@/utils";
import {
  Currency,
  Ventilation,
  VentilationLabels,
  type MontantPart,
} from "../../logic/api";

const props = defineProps<{
  currency: Currency;
  montant: MontantPart;
}>();

const icons = {
  [Ventilation.NoVentilation]: "",
  [Ventilation.Activites]: "mdi-ticket",
  [Ventilation.Alimentation]: "mdi-food",
  [Ventilation.FournitureEntretien]: "mdi-broom",
  [Ventilation.FournitureAdmin]: "mdi-file-document-multiple-outline",
  [Ventilation.FourniturePedagogique]: "mdi-book-education-outline",
  [Ventilation.Medical]: "mdi-medical-bag",
  [Ventilation.Hotelier]: "mdi-room-service",
  [Ventilation.Camping]: "mdi-tent",
  [Ventilation.Carburant]: "mdi-gas-station",
  [Ventilation.MaintenanceVehic]: "mdi-car-wrench",
  [Ventilation.Travels]: "mdi-map-marker-path",
  [Ventilation.Telephone]: "mdi-phone",
  [Ventilation.Poste]: "mdi-mailbox",
  [Ventilation.Bank]: "mdi-bank",
  [Ventilation.PensionBenevole]: "mdi-hand-heart",
  [Ventilation.Autre]: "mdi-dots-horizontal-circle-outline",
};
</script>
