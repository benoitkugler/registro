<template>
  <NavBar></NavBar>
  <v-container class="fill-height">
    <AdminHome v-if="controller.isAdmin"></AdminHome>
    <DirecteurHome v-else></DirecteurHome>
  </v-container>
</template>

<script lang="ts" setup>
import AdminHome from "../components/home/AdminHome.vue";
import DirecteurHome from "../components/home/DirecteurHome.vue";
import NavBar from "../components/NavBar.vue";
import { controller } from "../logic/logic";
</script>
