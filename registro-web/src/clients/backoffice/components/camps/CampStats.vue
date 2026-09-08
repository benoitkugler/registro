<template>
  <v-card title="Statut des inscriptions">
    <v-card-text>
      <v-list density="compact">
        <v-list-item
          title="Demandes d'inscriptions"
          subtitle="Tous status confondus "
        >
          <template #append>
            {{ props.stats.Inscriptions }}
          </template>
        </v-list-item>
        <v-list-item subtitle="Dont filles" prepend-icon="mdi-blank">
          <template #append>
            {{ props.stats.InscriptionsFilles }}
            {{
              pourcentS(
                props.stats.InscriptionsFilles,
                props.stats.Inscriptions,
              )
            }}
          </template>
        </v-list-item>
        <v-list-item subtitle="Dont suisses" prepend-icon="mdi-blank">
          <template #append>
            {{ props.stats.InscriptionsSuisses }}
            {{
              pourcentS(
                props.stats.InscriptionsSuisses,
                props.stats.Inscriptions,
              )
            }}
          </template>
        </v-list-item>
        <v-list-item title="Inscriptions acceptées">
          <template #append>
            {{ props.stats.Valides }}
            {{ pourcentS(props.stats.Valides, props.stats.Inscriptions) }}
          </template>
        </v-list-item>
        <v-list-item subtitle="Dont filles" prepend-icon="mdi-blank">
          <template #append>
            {{ props.stats.ValidesFilles }}
            {{ pourcentS(props.stats.ValidesFilles, props.stats.Valides) }}
          </template>
        </v-list-item>
        <v-list-item subtitle="Dont suisses" prepend-icon="mdi-blank">
          <template #append>
            {{ props.stats.ValidesSuisses }}
            {{ pourcentS(props.stats.ValidesSuisses, props.stats.Valides) }}
          </template>
        </v-list-item>
        <v-list-item
          title="Liste d'attente"
          subtitle="Inscriptions statuées et placées en attente"
        >
          <template #append>
            {{ props.stats.ListeAttente }}
            {{ pourcentS(props.stats.ListeAttente, props.stats.Inscriptions) }}
          </template>
        </v-list-item>
        <v-list-item
          title="Refus définitif"
          subtitle="Inscriptions statuées et refusées"
        >
          <template #append>
            {{ props.stats.Refus }}
            {{ pourcentS(props.stats.Refus, props.stats.Inscriptions) }}
          </template>
        </v-list-item>
        <v-list-item title="Demandes à statuer">
          <template #append>
            {{ props.stats.AStatuerRegular + props.stats.AStatuerException }}
          </template>
        </v-list-item>
        <v-list-item subtitle="Dont profils réguliers" prepend-icon="mdi-blank">
          <template #append>
            {{ props.stats.AStatuerRegular }}
          </template>
        </v-list-item>
        <v-list-item
          subtitle="Dont profils exceptionnels"
          prepend-icon="mdi-blank"
        >
          <template #append>
            {{ props.stats.AStatuerException }}
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
