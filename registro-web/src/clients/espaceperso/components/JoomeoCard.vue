<template>
  <v-skeleton-loader
    class="mx-auto"
    width="400px"
    type="card"
    v-if="data == null"
  ></v-skeleton-loader>
  <v-card
    v-else-if="data.Loggin"
    title="Identifiants et albums Joomeo"
    subtitle="Retrouvez les photos prises dans nos séjours."
    min-width="300"
  >
    <v-card-text>
      <v-row>
        <v-col cols="12" sm="6">
          Identifiant : <b>{{ data.Loggin }} </b>
        </v-col>
        <v-col cols="12" sm="6">
          Mot de passe : <b> {{ data.Password }}</b>
        </v-col>
      </v-row>
      <v-row>
        <v-col>
          <i v-if="!data.Albums?.length"
            >Aucun album n'est encore disponible.</i
          >
          <v-chip v-for="album in data.Albums" class="ma-1">{{ album }}</v-chip>
        </v-col>
      </v-row>
    </v-card-text>
    <v-card-actions>
      <v-btn
        block
        color="primary"
        variant="outlined"
        :href="data.SpaceURL"
        target="_blank"
      >
        Accéder à mon espace jooméo
      </v-btn>
    </v-card-actions>
  </v-card>
  <v-alert v-else class="mx-auto" type="info">
    Il n'y a encore rien à voir...
  </v-alert>
</template>

<script lang="ts" setup>
import { onMounted, ref } from "vue";
import { controller } from "../logic/logic";
import { type Joomeo } from "../logic/api";

const props = defineProps<{
  token: string;
}>();

onMounted(fetchData);

const data = ref<Joomeo | null>(null);
async function fetchData() {
  const res = await controller.LoadJoomeo({ token: props.token });
  if (res === undefined) return;
  data.value = res;
}
</script>
