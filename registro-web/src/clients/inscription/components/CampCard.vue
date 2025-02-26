<template>
  <v-card class="my-0">
    <v-card-text>
      <v-row no-gutters class="mb-1">
        <v-col cols="8">
          <v-chip prepend-icon="mdi-calendar-blank">
            {{ Camps.formatPlage(props.camp) }}
          </v-chip>
        </v-col>

        <v-col cols="4" class="text-right">
          <v-chip prepend-icon="mdi-currency-eur">
            {{ props.camp.Prix }}
          </v-chip>
        </v-col>
      </v-row>
      <v-row no-gutters class="my-1">
        <v-col cols="8">
          <v-chip prepend-icon="mdi-map-marker"> {{ props.camp.Lieu }}</v-chip>
        </v-col>
        <v-col cols="4" class="text-right">
          <v-chip prepend-icon="mdi-human-child">
            de {{ props.camp.AgeMin }} à {{ props.camp.AgeMax }} ans
          </v-chip>
        </v-col>
      </v-row>
      <v-row no-gutters class="my-1">
        <v-col>
          <v-chip prepend-icon="mdi-account-supervisor"
            >Direction : {{ props.camp.Direction }}</v-chip
          >
        </v-col>
      </v-row>
      <v-row v-if="props.camp.Description.length">
        <v-col>
          <v-alert type="info">
            {{ props.camp.Description }}
          </v-alert>
        </v-col>
      </v-row>
      <v-row v-if="props.camp.Navette.Actif">
        <v-col>
          <v-alert color="info" icon="mdi-bus">
            {{ props.camp.Navette.Commentaire }}
          </v-alert>
        </v-col>
      </v-row>
    </v-card-text>
  </v-card>
</template>

<script lang="ts" setup>
import type { CampExt } from "../logic/api";
import { Camps, Formatters } from "@/utils";

const props = defineProps<{
  camp: CampExt;
}>();
</script>
