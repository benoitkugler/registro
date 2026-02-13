<template>
  <v-navigation-drawer
    app
    :model-value="!props.hideMenu"
    expand-on-hover
    permanent
    rail
  >
    <v-list-item
      prepend-icon="mdi-menu"
      title="Registro - Directeurs"
      :subtitle="version"
    >
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
      prepend-icon="mdi-mail"
      link
      :to="{ path: '/documents' }"
      color="primary"
    >
      Documents
    </v-list-item>
    <v-divider></v-divider>
    <v-list-item
      prepend-icon="mdi-account-hard-hat"
      link
      :to="{ path: '/equipiers' }"
      color="primary"
    >
      Equipiers
    </v-list-item>
    <v-divider></v-divider>
    <v-list-item
      prepend-icon="mdi-image-album"
      link
      :to="{ path: '/photos' }"
      color="primary"
    >
      Album photos
    </v-list-item>

    <v-divider></v-divider>
    <v-list-item
      prepend-icon="mdi-comment-quote"
      link
      :to="{ path: '/avis' }"
      color="primary"
    >
      Avis
    </v-list-item>
    <v-divider></v-divider>
    <!-- This is not ready just yet -->
    <!-- <v-list-item
      prepend-icon="mdi-credit-card-outline"
      append-icon="mdi-open-in-new"
      link
      target="_blank"
      :href="controller.comptaURL"
      color="primary"
    >
      Comptabilité
    </v-list-item> -->

    <v-divider></v-divider>
    <v-list-item
      prepend-icon="mdi-logout"
      link
      :to="{ path: '/' }"
      color="primary"
      @click="CachedTokens.clear()"
    >
      Se déconnecter
    </v-list-item>
  </v-navigation-drawer>

  <v-app-bar rounded elevation="4" color="primary">
    <v-app-bar-title>
      <v-row>
        <v-col align-self="center" cols="auto">
          <v-img width="150" :src="logo" />
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
import { CachedTokens } from "../logic/logic";

const props = defineProps<{
  title: string;
  hideMenu?: boolean;
}>();

const logo = `${import.meta.env.BASE_URL}${import.meta.env.VITE_ASSO}/logo.png`;

const version = `v${VITE_APP_VERSION}`;
</script>
