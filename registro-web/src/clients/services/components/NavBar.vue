<template>
  <v-app-bar rounded elevation="4" color="primary">
    <v-app-bar-title>
      <v-row>
        <v-col align-self="center" cols="auto">
          <v-img width="150" :src="logo" />
        </v-col>
        <v-col align-self="center"> {{ props.title }} </v-col>
      </v-row>
    </v-app-bar-title>
  </v-app-bar>
</template>

<script setup lang="ts">
const logo = `${import.meta.env.BASE_URL}${import.meta.env.VITE_ASSO}/logo.png`;

const version = `v${VITE_APP_VERSION}`;

const props = defineProps<{
  title: string;
}>();
</script>
