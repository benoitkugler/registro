<template>
  <v-row no-gutters class="ma-2">
    <v-col cols="5">
      <v-card title="Dossiers" subtitle="Nombre">
        <template v-slot:append>
          <v-text-field
            width="200"
            append-inner-icon="mdi-magnify"
            label="Rechercher"
            hide-details
          ></v-text-field>
        </template>

        <v-card-text>
          <v-list>
            <v-list-item></v-list-item>
          </v-list>
        </v-card-text>
      </v-card>
    </v-col>
    <v-col></v-col>
  </v-row>
</template>

<script setup lang="ts">
import { ref } from "vue";
const props = defineProps<{}>();
</script>
