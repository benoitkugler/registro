<template>
  <v-app-bar rounded elevation="4" color="primary">
    <v-app-bar-title>
      <v-row>
        <v-col align-self="center" cols="auto">
          <v-img :width="xs ? 80 : 150" :src="xs ? logoSmall : logo" />
        </v-col>
        <v-col align-self="center" v-if="!smAndDown">
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
import { useDisplay } from "vuetify";

const props = defineProps<{
  title: string;
  hideMenu?: boolean;
}>();

const logo = `${import.meta.env.BASE_URL}${import.meta.env.VITE_ASSO}/logo.png`;
const logoSmall = `${import.meta.env.BASE_URL}${
  import.meta.env.VITE_ASSO
}/logo-small.png`;

const version = `v${VITE_APP_VERSION}`;

const { xs, smAndDown } = useDisplay();
</script>
