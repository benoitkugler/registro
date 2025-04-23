<template>
  <v-row>
    <v-col v-for="camp in props.camps" cols="12" sm="6" md="4">
      <v-card @click="emit('clicked', camp.Id)">
        <v-row>
          <v-col cols="7">
            <v-card-title>
              {{ Camps.label(camp) }}
            </v-card-title>

            <v-card-subtitle>{{ Camps.formatPlage(camp) }}</v-card-subtitle>

            <v-card-text>
              <v-chip prepend-icon="mdi-map-marker"> {{ camp.Lieu }}</v-chip>
            </v-card-text>
          </v-col>
          <v-col align-self="center" cols="5" v-if="camp.ImageURL">
            <v-img class="mr-2" :src="camp.ImageURL"></v-img>
          </v-col>
        </v-row>
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
