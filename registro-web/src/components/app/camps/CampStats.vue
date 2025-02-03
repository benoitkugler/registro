<template>
  <v-card title="Statistiques">
    <v-card-text>
      <v-row>
        <v-col cols="4"> Inscriptions : {{ props.stats.Inscriptions }} </v-col>
        <v-col class="text-right">
          Filles : {{ $props.stats.InscriptionsFilles }} ({{
            pourcent($props.stats.InscriptionsFilles, $props.stats.Inscriptions)
          }}
          %), Suisses : {{ $props.stats.InscriptionsSuisses }} ({{
            pourcent(
              $props.stats.InscriptionsSuisses,
              $props.stats.Inscriptions
            )
          }}
          %)
        </v-col>
      </v-row>
      <v-row>
        <v-col cols="4"> Confirmées : {{ props.stats.Valides }} </v-col>
        <v-col class="text-right">
          Filles : {{ $props.stats.ValidesFilles }} ({{
            pourcent($props.stats.ValidesFilles, $props.stats.Valides)
          }}
          %), Suisses : {{ $props.stats.ValidesSuisses }} ({{
            pourcent($props.stats.ValidesSuisses, $props.stats.Valides)
          }}
          %)
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

function pourcent(val: number, max: number) {
  return (max == 0 ? 0 : (100 * val) / max).toFixed(0);
}

const fillesPourcent = computed(() =>
  (props.stats.Inscriptions == 0
    ? 0
    : (100 * props.stats.InscriptionsFilles) / props.stats.Inscriptions
  ).toFixed(0)
);
const suissesPourcent = computed(() =>
  (props.stats.Inscriptions == 0
    ? 0
    : (100 * props.stats.InscriptionsSuisses) / props.stats.Inscriptions
  ).toFixed(0)
);
</script>
