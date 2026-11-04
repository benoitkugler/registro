<template>
  <NavBar hideMenu></NavBar>
  <v-container class="fill-height">
    <v-card class="mx-auto" width="400px">
      <v-card-text>
        <form>
          <v-text-field
            label="Clé de connection"
            v-model="password"
            :error-messages="errors"
            :type="showPassword ? undefined : 'password'"
            name="password"
            @keydown.enter.prevent="
              () => (password.length == 0 ? null : loggin())
            "
            autocomplete="current-password"
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
        </form>
      </v-card-text>
      <v-card-actions>
        <v-spacer></v-spacer>
        <v-btn :disabled="!password.length" @click="loggin">Se connecter</v-btn>
      </v-card-actions>
    </v-card>
  </v-container>
</template>

<script lang="ts" setup>
import { onMounted, ref } from "vue";
import { CachedTokens, controller } from "../logic/logic";
import { useRouter } from "vue-router";
import NavBar from "../components/NavBar.vue";
import type { IdComptaCamp } from "../logic/api.ts";

const router = useRouter();

// const emit = defineEmits<{
//  }>();

onMounted(init);

// resolve loggin strategy :
//  - give priority to director link (query param 'key')
//  - otherwise, try to use a cached token
//  - fallback to admin loggin screen
async function init() {
  const directorKey = (router.currentRoute.value.query["key"] || "") as string;
  if (directorKey) {
    const res = await controller.Loggin({ key: directorKey, isAdmin: false });
    if (res === undefined) return;

    onSuccessfullLoggin(res.IdCompta, res.Token);
  } else {
    const { idCompta, token } = CachedTokens.get();
    if (token) {
      onSuccessfullLoggin(idCompta, token);
    }
    // otherwise, just show the loggin screen
  }
}

const showPassword = ref(false);
const password = ref("");
const errors = ref<string[]>([]);
async function loggin() {
  const res = await controller.Loggin({ key: password.value, isAdmin: true });
  if (res === undefined) return;

  if (res.IsValid) {
    onSuccessfullLoggin(res.IdCompta, res.Token);
  } else {
    errors.value = ["Clé de connection incorrecte."];
  }
}

function onSuccessfullLoggin(idCompta: IdComptaCamp, token: string) {
  CachedTokens.set(idCompta, token);

  controller.setToken(token);
  controller.isAdmin = idCompta == -1;
  controller.idCompta = idCompta;

  controller.showMessage("Bienvenue !");
  router.push({ path: "/home" });
}
</script>
