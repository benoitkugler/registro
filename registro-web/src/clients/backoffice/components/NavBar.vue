<template>
  <v-navigation-drawer
    app
    :model-value="!props.hideMenu"
    expand-on-hover
    permanent
    rail
  >
    <v-list-item prepend-icon="mdi-menu" :title="`Registro ${version}`">
    </v-list-item>
    <v-divider></v-divider>

    <v-list-item
      prepend-icon="mdi-bed"
      link
      :to="{ path: '/camps' }"
      color="primary"
    >
      Séjours
    </v-list-item>
    <v-list-item
      prepend-icon="mdi-calendar-multiple-check"
      link
      :to="{ path: '/inscriptions' }"
      color="primary"
    >
      Inscriptions
    </v-list-item>
    <v-divider></v-divider>
    <v-list-item
      prepend-icon="mdi-account-multiple"
      link
      :to="{ path: '/annuaire' }"
      color="primary"
    >
      Annuaire
    </v-list-item>
    <v-divider> </v-divider>

    <template #append>
      <v-list-item
        prepend-icon="mdi-account-circle-outline"
        :title="isFondSoutien ? 'Fonds de soutien' : 'Administrateur'"
      >
      </v-list-item>
      <v-list-item
        prepend-icon="mdi-logout"
        link
        :to="{ path: '/' }"
        color="primary"
      >
        Se déconnecter
      </v-list-item>
    </template>
  </v-navigation-drawer>

  <v-app-bar rounded elevation="4" color="secondary">
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
import { computed, ref } from "vue";
import { controller } from "../logic/logic";

const logo = `${import.meta.env.BASE_URL}${import.meta.env.VITE_ASSO}/logo.png`;

const version = `v${VITE_APP_VERSION}`;

const props = defineProps<{
  title: string;
  hideMenu?: boolean;
}>();

const isFondSoutien = computed(() => controller.isFondsSoutien);
</script>
