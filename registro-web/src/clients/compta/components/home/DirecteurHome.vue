<template>
  <v-responsive class="align-center text-center fill-height">
    <v-alert color="primary">
      <v-row>
        <v-col align-self="center">
          Bienvenue sur votre espace de comptabilité
        </v-col>
        <v-col align-self="center">
          <v-btn @click="goToSejour">Aller au tableau de bord</v-btn>
        </v-col>
      </v-row>
    </v-alert>
  </v-responsive>
</template>

<script lang="ts" setup>
import { useRouter } from "vue-router";

const router = useRouter();

function goToSejour() {
  router.push({ name: "sejour" });
}
</script>
