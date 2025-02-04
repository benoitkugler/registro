<template>
  <v-navigation-drawer app v-model="showSideBar" temporary>
    <v-list-item title="Registro - Backoffice" :subtitle="version">
    </v-list-item>
    <v-divider></v-divider>

    <v-list-item
      prepend-icon="mdi-home"
      link
      :to="{ path: '/' }"
      color="primary"
    >
      Accueil
    </v-list-item>
    <v-list-item
      prepend-icon="mdi-bed"
      link
      :to="{ path: '/app/camps' }"
      color="primary"
    >
      Camps
    </v-list-item>
    <!-- <v-list-item
        prepend-icon="mdi-currency-eur"
        link
        :to="{ name: 'budget' }"
        color="primary"
      >
        Budget prévisionnel
      </v-list-item>
      <v-list-item
        prepend-icon="mdi-view-list"
        link
        :to="{ name: 'records' }"
        color="primary"
      >
        Enregistrements
      </v-list-item> -->
  </v-navigation-drawer>

  <v-app-bar rounded elevation="4" color="secondary">
    <v-app-bar-nav-icon
      @click="showSideBar = !showSideBar"
    ></v-app-bar-nav-icon>
    <v-app-bar-title>
      <v-row>
        <v-col align-self="center" cols="auto">
          <v-img width="60" :src="logo" />
        </v-col>
        <v-col align-self="center">
          {{ $props.title }}
        </v-col>
      </v-row>
    </v-app-bar-title>

    <template v-slot:append>
      <slot></slot>
      <div class="mr-2"></div>
    </template>
  </v-app-bar>
</template>

<script setup lang="ts">
import { ref } from "vue";

const logo = `${import.meta.env.BASE_URL}${import.meta.env.VITE_ASSO}/logo.png`;

const version = `v${VITE_APP_VERSION}`;

const showSideBar = ref(false);
const props = defineProps<{
  title: string;
}>();
</script>
