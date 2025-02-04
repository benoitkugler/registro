<template>
  <v-card title="Statistiques">
    <v-card-text>
      <v-row>
        <v-col cols="5">
          Demandes d'inscriptions : {{ $props.stats.Inscriptions }}
        </v-col>
        <v-col class="text-right">
          Dont Filles : {{ $props.stats.InscriptionsFilles }}
          {{
            pourcentS(
              $props.stats.InscriptionsFilles,
              $props.stats.Inscriptions
            )
          }}, Suisses : {{ $props.stats.InscriptionsSuisses }}
          {{
            pourcentS(
              $props.stats.InscriptionsSuisses,
              $props.stats.Inscriptions
            )
          }}
        </v-col>
      </v-row>
      <v-row>
        <v-col cols="4">
          Confirmées : {{ $props.stats.Valides }}
          {{ pourcentS($props.stats.Valides, $props.stats.Inscriptions) }}
        </v-col>
        <v-col class="text-right">
          Dont Filles : {{ $props.stats.ValidesFilles }}
          {{ pourcentS($props.stats.ValidesFilles, $props.stats.Valides) }} ,
          Suisses : {{ $props.stats.ValidesSuisses }}
          {{ pourcentS($props.stats.ValidesSuisses, $props.stats.Valides) }}
        </v-col>
      </v-row>
      <v-row>
        <v-col>
          A statuer : {{ $props.stats.AStatuer }}
          {{ pourcentS($props.stats.AStatuer, $props.stats.Inscriptions) }}
        </v-col>
        <v-col cols="5" class="text-center">
          Demandes d'exception : {{ $props.stats.Exceptions }}
          {{ pourcentS($props.stats.Exceptions, $props.stats.Inscriptions) }}
        </v-col>
        <v-col class="text-right">
          Refus : {{ $props.stats.Refus }}
          {{ pourcentS($props.stats.Refus, $props.stats.Inscriptions) }}
        </v-col>
      </v-row>
    </v-card-text>
  </v-card>
</template>

<script setup lang="ts">
import { ref, computed } from "vue";
import type { StatistiquesInscrits } from "@/logic/app/api";
const props = defineProps<{
  stats: StatistiquesInscrits;
}>();

function pourcentS(val: number, max: number) {
  const p = (max == 0 ? 0 : (100 * val) / max).toFixed(0);
  return `(${p} %)`;
}
</script>
