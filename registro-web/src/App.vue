<template>
  <v-app>
    <v-main>
      <router-view />
    </v-main>
  </v-app>

  <!-- <router-view v-if="isLoggedIn" />
      <v-skeleton-loader v-else type="card"></v-skeleton-loader> -->

  <!-- <v-snackbar
        style="z-index: 10000"
        app
        :model-value="message != ''"
        @update:model-value="message = ''"
        :timeout="4000"
        :color="messageColor"
        location="bottom left"
        close-on-content-click
      >
        {{ message }}
      </v-snackbar>

      <v-snackbar
        app
        :model-value="errorKind != ''"
        @update:model-value="errorKind = ''"
        :timeout="4000"
        color="red"
      >
        <b>{{ errorKind }}</b>
        <div v-html="errorHtml"></div>
      </v-snackbar> -->
</template>

<script lang="ts" setup>
//
</script>
