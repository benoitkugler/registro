<template>
  <v-navigation-drawer
    app
    :model-value="!props.hideMenu"
    expand-on-hover
    permanent
    v-model:rail="rail"
  >
    <v-list-item prepend-icon="mdi-menu" :title="`Registro ${version}`">
    </v-list-item>
    <v-divider horizontal thickness="1"></v-divider>

    <v-list-item
      prepend-icon="mdi-home"
      link
      :to="{ path: '/home' }"
      color="primary"
    >
      Accueil
    </v-list-item>
    <v-list-item
      prepend-icon="mdi-chart-box"
      link
      :to="{ path: '/summary' }"
      color="primary"
    >
      Tableau de bord
    </v-list-item>
    <!-- <v-list-item
            prepend-icon="mdi-currency-eur"
            link
            :to="{ path: '/budget' }"
            color="primary"
          >
            Budget prévisionnel
          </v-list-item> -->
    <v-list-item
      prepend-icon="mdi-view-list"
      link
      :to="{ path: '/records' }"
      color="primary"
    >
      Enregistrements
    </v-list-item>
  </v-navigation-drawer>

  <v-app-bar rounded elevation="4" color="primary">
    <v-app-bar-title>
      <v-row>
        <v-col align-self="center" cols="auto" v-if="smAndUp || props.hideMenu">
          <v-img width="150" :src="logo" />
        </v-col>
        <v-col align-self="center"> Comptabilité des séjours </v-col>
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
import { useDisplay } from "vuetify";

const props = defineProps<{
  hideMenu?: boolean;
}>();

const logo = `${import.meta.env.BASE_URL}${import.meta.env.VITE_ASSO}/logo.png`;

const version = `v${VITE_APP_VERSION}`;

const { smAndUp } = useDisplay();

const rail = ref(true);
</script>
