<template>
  <NavBar title="Connection à la page Directeur" hide-menu> </NavBar>

  <v-container class="fill-height">
    <v-card title="Connection" class="mx-auto" width="400px">
      <v-card-text>
        <SelectCamp
          label="Camp"
          v-model="selected"
          :camps="camps"
          name="user"
          default-style
        ></SelectCamp>

        <v-text-field
          label="Mot de passe"
          v-model="password"
          @update:model-value="errors = []"
          :error-messages="errors"
          :type="showPassword ? undefined : 'password'"
          name="password"
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
        <v-btn :disabled="!selected || !password.length" @click="loggin"
          >Se connecter</v-btn
        >
      </v-card-actions>
    </v-card>
  </v-container>
</template>

<script lang="ts" setup>
import { useRouter } from "vue-router";
import NavBar from "../components/NavBar.vue";
import { onMounted, ref } from "vue";
import { controller } from "../logic/logic";
import type { CampItem, IdCamp } from "../logic/api";

const router = useRouter();

onMounted(loadCamps);

const camps = ref<CampItem[]>([]);
async function loadCamps() {
  const res = await controller.GetCamps();
  if (res === undefined) return;
  camps.value = res || [];
}

const selected = ref<IdCamp>(0 as IdCamp);

const showPassword = ref(false);
const password = ref("");
const errors = ref<string[]>([]);
async function loggin() {
  const item = camps.value.find((item) => item.Id == selected.value);
  if (!item) return;

  const res = await controller.Loggin({
    password: password.value,
    idCamp: selected.value,
  });
  if (res === undefined) return;

  if (res.IsValid) {
    controller.setCamp(item, res.Token);
    controller.showMessage("Bienvenue !");
    router.push({ path: "/inscriptions" });
  } else {
    errors.value = ["Mot de passe incorrect."];
  }
}
</script>
