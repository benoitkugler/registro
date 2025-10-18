<template>
  <NavBar title="Connection au backoffice" hide-menu> </NavBar>

  <v-container class="fill-height">
    <v-card class="mx-auto" width="400px">
      <v-card-text>
        <v-text-field
          label="Clé de connection"
          v-model="password"
          :error-messages="errors"
          :type="showPassword ? undefined : 'password'"
          name="password"
          @keydown.enter.prevent="
            () => (password.length == 0 ? null : loggin())
          "
        >
          <template #append-inner>
            <v-btn
              class="ml-2"
              size="small"
              :icon="showPassword ? 'mdi-eye' : 'mdi-eye-off'"
              flat
              @click="showPassword = !showPassword"
            >
            </v-btn>
          </template>
        </v-text-field>
      </v-card-text>
      <v-card-actions>
        <v-spacer></v-spacer>
        <v-btn :disabled="!password.length" @click="loggin">Se connecter</v-btn>
      </v-card-actions>
    </v-card>
  </v-container>
</template>

<script lang="ts" setup>
import { ref } from "vue";
import NavBar from "../components/NavBar.vue";
import { controller } from "../logic/logic";
import { useRouter } from "vue-router";

const router = useRouter();

const showPassword = ref(false);
const password = ref("");
const errors = ref<string[]>([]);
async function loggin() {
  const res = await controller.Loggin({ password: password.value });
  if (res === undefined) return;

  if (res.IsValid) {
    controller.setToken(res.Token, res.IsFondSoutien);
    controller.showMessage("Bienvenue !");
    router.push({ path: "/inscriptions" });
  } else {
    errors.value = ["Clé de connection incorrecte."];
  }
}
</script>
