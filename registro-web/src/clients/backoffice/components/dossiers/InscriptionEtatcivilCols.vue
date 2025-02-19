<template>
  <v-col align-self="center" cols="auto" class="mr-2">
    <v-tooltip>
      <template v-slot:activator="{ props: innerProps }">
        <v-icon
          v-bind="innerProps"
          :icon="props.personne.IsTemp ? 'mdi-account' : 'mdi-account-check'"
          :color="props.personne.IsTemp ? 'orange' : 'green'"
        ></v-icon>
      </template>
      <template v-if="props.personne.IsTemp">
        Ce profil est temporaire et doit être identifié.
      </template>
      <template v-else> Ce profil est identifié. </template>
    </v-tooltip>
  </v-col>
  <v-col align-self="center" cols="4"
    >{{ props.personne.Nom }} {{ props.personne.Prenom }}</v-col
  >
  <v-col align-self="center" cols="2"
    >{{ SexeLabels[props.personne.Sexe] }}
  </v-col>
  <v-col align-self="center" cols="3"
    >né(e) le
    {{ new Date(props.personne.DateNaissance).toLocaleDateString() }}
  </v-col>
</template>

<script setup lang="ts">
import { SexeLabels, type Personne } from "../../logic/api";
const props = defineProps<{
  personne: Personne;
}>();
</script>
