<template>
  <v-row>
    <v-col v-for="camp in props.camps" cols="12" sm="6" md="4">
      <v-card
        :title="Camps.label(camp)"
        :subtitle="Camps.formatPlage(camp)"
        @click="emit('clicked', camp.Id)"
      >
        <v-card-text>
          <v-chip prepend-icon="mdi-map-marker"> {{ camp.Lieu }}</v-chip>
        </v-card-text>
      </v-card>
    </v-col>
  </v-row>
</template>

<script lang="ts" setup>
import type { CampExt, IdCamp } from "../logic/api";
import { Camps } from "@/utils";

const props = defineProps<{
  camps: CampExt[];
}>();

const emit = defineEmits<{
  (e: "clicked", idCamp: IdCamp): void;
}>();
</script>
