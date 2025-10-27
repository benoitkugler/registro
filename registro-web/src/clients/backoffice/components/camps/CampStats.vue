<template>
  <v-card title="Statut des inscriptions">
    <v-card-text>
      <v-list density="compact">
        <v-list-item title="Demandes d'inscriptions">
          <template #append>
            {{ props.stats.Inscriptions }}
          </template>
        </v-list-item>
        <v-list-item
          title="Filles"
          subtitle="Parmi toutes les demandes d'inscription"
        >
          <template #append>
            {{ props.stats.InscriptionsFilles }}
            {{
              pourcentS(
                props.stats.InscriptionsFilles,
                props.stats.Inscriptions
              )
            }}
          </template>
        </v-list-item>
        <v-list-item
          title="Suisses"
          subtitle="Parmi toutes les demandes d'inscription"
        >
          <template #append>
            {{ props.stats.InscriptionsSuisses }}
            {{
              pourcentS(
                props.stats.InscriptionsSuisses,
                props.stats.Inscriptions
              )
            }}
          </template>
        </v-list-item>
        <v-list-item title="Inscriptions confirmées">
          <template #append>
            {{ props.stats.Valides }}
            {{ pourcentS(props.stats.Valides, props.stats.Inscriptions) }}
          </template>
        </v-list-item>
        <v-list-item
          title="Filles"
          subtitle="Parmi les inscriptions confirmées"
        >
          <template #append>
            {{ props.stats.ValidesFilles }}
            {{ pourcentS(props.stats.ValidesFilles, props.stats.Valides) }}
          </template>
        </v-list-item>
        <v-list-item
          title="Suisses"
          subtitle="Parmi les inscriptions confirmées"
        >
          <template #append>
            {{ props.stats.ValidesSuisses }}
            {{ pourcentS(props.stats.ValidesSuisses, props.stats.Valides) }}
          </template>
        </v-list-item>
        <v-list-item title="A statuer">
          <template #append>
            {{ props.stats.AStatuer }}
            {{ pourcentS(props.stats.AStatuer, props.stats.Inscriptions) }}
          </template>
        </v-list-item>
        <v-list-item title="Demandes d'exception">
          <template #append>
            {{ props.stats.Exceptions }}
            {{ pourcentS(props.stats.Exceptions, props.stats.Inscriptions) }}
          </template>
        </v-list-item>
        <v-list-item title="Refus">
          <template #append>
            {{ props.stats.Refus }}
            {{ pourcentS(props.stats.Refus, props.stats.Inscriptions) }}
          </template>
        </v-list-item>
      </v-list>
    </v-card-text>
  </v-card>
</template>

<script setup lang="ts">
import type { StatistiquesInscrits } from "@/clients/backoffice/logic/api";
import { Formatters } from "@/utils";
const props = defineProps<{
  stats: StatistiquesInscrits;
}>();

function pourcentS(val: number, max: number) {
  return `(${Formatters.pourcent(val, max)} %)`;
}
</script>
