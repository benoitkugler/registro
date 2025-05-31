<template>
  <v-navigation-drawer app v-model="showSideBar" temporary>
    <v-list-item title="Registro - Portail des directeurs" :subtitle="version">
    </v-list-item>
    <v-divider></v-divider>

    <v-list-item
      prepend-icon="mdi-calendar-multiple-check"
      link
      :to="{ path: '/inscriptions' }"
      color="primary"
    >
      Inscriptions
    </v-list-item>
    <v-list-item
      prepend-icon="mdi-account"
      link
      :to="{ path: '/equipiers' }"
      color="primary"
    >
      Equipiers
    </v-list-item>
    <v-list-item
      prepend-icon="mdi-mail"
      link
      :to="{ path: '/documents' }"
      color="primary"
    >
      Communication
    </v-list-item>
    <v-divider></v-divider>
    <v-list-item
      prepend-icon="mdi-logout"
      link
      :to="{ path: '/' }"
      color="primary"
    >
      Se déconnecter
    </v-list-item>
  </v-navigation-drawer>

  <v-app-bar rounded elevation="4" color="secondary">
    <v-app-bar-nav-icon
      v-if="!props.hideMenu"
      @click="showSideBar = !showSideBar"
    ></v-app-bar-nav-icon>
    <v-app-bar-title>
      <v-row>
        <v-col align-self="center" cols="auto">
          <v-img width="60" :src="logo" />
        </v-col>
        <v-col align-self="center">
          {{ props.title }}
        </v-col>
      </v-row>
    </v-app-bar-title>

    <template #append>
      <slot></slot>
      <div class="mr-2"></div>
    </template>
  </v-app-bar>
</template>

<script setup lang="ts">
import { ref } from "vue";

const props = defineProps<{
  title: string;
  hideMenu?: boolean;
}>();

const logo = `${import.meta.env.BASE_URL}${import.meta.env.VITE_ASSO}/logo.png`;

const version = `v${VITE_APP_VERSION}`;

const showSideBar = ref(false);
</script>
