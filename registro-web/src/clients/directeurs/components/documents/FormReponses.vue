<template>
  <v-card
    title="Réponses des inscrits"
    subtitle="Ces réponses sont incluses dans la liste Excel des participants."
  >
    <v-skeleton-loader v-if="!data" type="table"></v-skeleton-loader>
    <v-card-text v-else>
      <v-table striped="even" density="compact">
        <thead>
          <tr>
            <th style="width: 200px">Nom</th>
            <th v-for="champ in props.form.Champs" style="text-align: center">
              {{ champ.Titre }}
            </th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="inscrit in data.Inscrits">
            <th>{{ inscrit.Inscrit }}</th>
            <td v-for="reponse in inscrit.Reponses">
              <MultilineText :text="reponse"></MultilineText>
            </td>
          </tr>
        </tbody>
      </v-table>
    </v-card-text>
  </v-card>
</template>

<script lang="ts" setup>
import { onMounted, ref } from "vue";
import type { Form, FormReponses } from "../../logic/api.ts";
import { controller } from "../../logic/logic.ts";

const props = defineProps<{
  form: Form;
}>();

const emit = defineEmits<{}>();

onMounted(fetchData);

const data = ref<FormReponses | null>(null);
async function fetchData() {
  const res = await controller.FormsLoadReponses({ id: props.form.Id });
  if (res === undefined) return;
  data.value = res;
}
</script>

<style>
table {
  table-layout: fixed;
}
</style>
