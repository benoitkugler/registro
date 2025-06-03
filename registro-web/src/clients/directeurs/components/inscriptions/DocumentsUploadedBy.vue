<template>
  <v-card subtitle="Document fourni par">
    <v-card-text>
      <v-list>
        <v-list-item v-for="personne in list" :title="Fmt.label(personne)">
          <template #append>
            <v-icon
              :icon="crible.has(personne.Id) ? 'mdi-check' : 'mdi-close'"
              :color="crible.has(personne.Id) ? 'green' : 'red'"
            ></v-icon>
          </template>
        </v-list-item>
      </v-list>
    </v-card-text>
  </v-card>
</template>

<script lang="ts" setup>
import { computed } from "vue";
import { type IdPersonne, type Personnes } from "../../logic/api";
import { Personnes as Fmt } from "@/utils";

const props = defineProps<{
  personnes: Personnes;
  uploadedBy: IdPersonne[];
}>();

const crible = computed(() => new Set(props.uploadedBy));
const list = computed(() =>
  Object.values(props.personnes || {}).sort((a, b) =>
    Fmt.label(a).localeCompare(Fmt.label(b))
  )
);
</script>
