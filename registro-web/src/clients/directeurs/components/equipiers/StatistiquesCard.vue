<template>
  <v-card subtitle="Statistiques" width="300px">
    <v-card-text>
      <v-row class="my-1" no-gutters justify="space-between">
        <v-col> Equipiers </v-col>
        <v-col cols="auto">
          <v-badge color="primary" inline :content="props.equipiers.length">
          </v-badge>
        </v-col>
      </v-row>
      <v-row class="my-1" no-gutters justify="space-between">
        <v-col> Animateurs </v-col>
        <v-col cols="auto">
          <v-badge
            color="primary"
            inline
            :content="props.equipiers.filter((p) => p.IsAnimateur).length"
          >
          </v-badge>
        </v-col>
      </v-row>
      <v-row class="my-1" no-gutters justify="space-between">
        <v-col> Au pair </v-col>
        <v-col cols="auto">
          <v-badge
            color="primary"
            inline
            :content="props.equipiers.filter((p) => p.IsAuPair).length"
          >
          </v-badge>
        </v-col>
      </v-row>
      <v-row class="my-1" no-gutters justify="space-between">
        <v-col> Anniversaires </v-col>
        <v-col cols="auto">
          <v-badge
            color="amber-lighten-1"
            inline
            :content="props.equipiers.filter((p) => p.HasBirthday).length"
          >
          </v-badge>
        </v-col>
      </v-row>
    </v-card-text>
  </v-card>
</template>

<script lang="ts" setup>
import { type EquipierExt } from "../../logic/api";

const props = defineProps<{
  equipiers: EquipierExt[];
}>();
</script>
