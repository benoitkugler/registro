<template>
  <v-alert class="text-center">
    <v-row justify="space-evenly">
      <v-col align-self="center"> Merci de sélectionner un séjour </v-col>
      <v-col align-self="center">
        <v-btn variant="tonal" color="primary" link :to="{ path: '/home' }"
          >Liste des séjours</v-btn
        >
      </v-col>
    </v-row>
  </v-alert>
</template>

<script lang="ts" setup></script>
