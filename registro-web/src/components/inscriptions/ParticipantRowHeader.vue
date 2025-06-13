<template>
  <v-row no-gutters class="mx-2 px-2">
    <v-col class="font-weight-bold" align-self="center" cols="3">
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
    <v-col class="font-weight-bold" align-self="center"> Commentaire </v-col>
    <v-spacer></v-spacer>
    <v-col class="font-weight-bold text-center" cols="2"> Inscription </v-col>
    <v-col cols="auto">
      <v-btn size="x-small" color="white" icon flat></v-btn>
    </v-col>
  </v-row>
</template>

<script setup lang="ts">
const props = defineProps<{}>();
</script>
