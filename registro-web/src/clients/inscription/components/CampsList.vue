<template>
  <v-row>
    <v-col v-for="camp in props.camps" cols="12" sm="6" md="4">
      <v-card @click="emit('clicked', camp.Id)" :disabled="camp.IsClosed">
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
          <v-col align-self="center" cols="5">
            <div style="position: relative" class="mr-2">
              <v-img :src="camp.ImageURL" v-if="camp.ImageURL"></v-img>
              <div v-else style="height: 80px"></div>

              <v-overlay
                :model-value="camp.IsComplet || camp.IsClosed"
                contained
                width="100%"
                height="100%"
                :scrim="false"
                persistent
              >
                <v-row no-gutters class="h-100" justify="center">
                  <v-col align-self="center" class="text-center">
                    <v-img
                      v-if="camp.IsComplet"
                      class="mx-auto"
                      width="90%"
                      :src="complet"
                      style="opacity: 0.7"
                    ></v-img>
                    <div v-else class="d-inline-block">
                      <v-icon size="42">mdi-lock-outline</v-icon>
                    </div>
                  </v-col>
                </v-row>
              </v-overlay>
            </div>
          </v-col>
        </v-row>
      </v-card>
    </v-col>
  </v-row>
</template>

<script lang="ts" setup>
import type { CampExt, IdCamp } from "../logic/api";
import { Camps } from "@/utils";
import complet from "@/assets/complet.png";

const props = defineProps<{
  camps: CampExt[];
}>();

const emit = defineEmits<{
  (e: "clicked", idCamp: IdCamp): void;
}>();
</script>
