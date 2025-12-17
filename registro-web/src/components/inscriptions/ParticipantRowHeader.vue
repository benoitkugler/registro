<template>
  <v-row no-gutters class="mx-2 px-2">
    <v-col class="font-weight-bold" align-self="center" cols="2">
      Participant
    </v-col>
    <v-col class="font-weight-bold text-center" align-self="center" cols="1">
      Sexe
    </v-col>
    <v-col class="font-weight-bold text-center" align-self="center" cols="1">
      Date de naissance
    </v-col>
    <v-col class="font-weight-bold text-center" align-self="center" cols="1">
      Age
    </v-col>
    <v-col class="font-weight-bold text-center" align-self="center" cols="1">
      Navette
    </v-col>
    <v-col class="font-weight-bold text-center" align-self="center" cols="2">
      Commentaire
    </v-col>
    <v-col class="font-weight-bold text-center" align-self="center" cols="2">
      Inscription
    </v-col>
    <slot name="append"></slot>
    <v-spacer v-if="$slots['append'] === undefined"></v-spacer>
    <v-col cols="auto">
      <v-btn size="x-small" color="white" icon variant="plain" disabled></v-btn>
    </v-col>
  </v-row>
</template>

<script setup lang="ts">
import { onMounted, useSlots } from "vue";

const props = defineProps<{}>();
</script>
