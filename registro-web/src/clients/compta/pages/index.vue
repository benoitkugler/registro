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
import { CachedTokens, controller, isDirectorLoggin } from "../logic/logic";
import { useRouter } from "vue-router";
import NavBar from "../components/NavBar.vue";
import type { IdComptaCamp } from "../logic/api.ts";

const router = useRouter();

// const emit = defineEmits<{
//  }>();

onMounted(init);

async function init() {
  const directorKey = isDirectorLoggin(router.currentRoute.value);
  if (directorKey) {
    const res = await controller.Loggin({ key: directorKey, isAdmin: false });
    if (res === undefined) return;

    controller.onSuccessfullLoggin(res.IdCompta, res.Token);
    router.push({ path: "/home" });
  }
}

const showPassword = ref(false);
const password = ref("");
const errors = ref<string[]>([]);
async function loggin() {
  const res = await controller.Loggin({ key: password.value, isAdmin: true });
  if (res === undefined) return;

  if (res.IsValid) {
    controller.onSuccessfullLoggin(res.IdCompta, res.Token);
    router.push({ path: "/home" });
  } else {
    errors.value = ["Clé de connection incorrecte."];
  }
}
</script>
