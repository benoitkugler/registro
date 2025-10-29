<template>
  <NavBar title="Mentions légales"> </NavBar>

  <v-container class="fill-height">
    <CguAcve v-if="asso == 'acve'"></CguAcve>
    <CguRepere v-else-if="asso == 'repere'"></CguRepere>
  </v-container>
</template>

<script lang="ts" setup>
import CguAcve from "../components/CguAcve.vue";
import CguRepere from "../components/CguRepere.vue";
import NavBar from "../components/NavBar.vue";

const asso = import.meta.env.VITE_ASSO;
</script>
