<template>
  <EventItem icon="mdi-mail" color="light-blue" :time="props.event.Created">
    <v-row>
      <v-col align-self="center">
        Les documents du séjour <i> {{ props.content.CampLabel }} </i> ont été
        envoyés.
      </v-col>
      <v-col align-self="center" cols="auto">
        <v-btn
          size="small"
          @click="emit('goToDocuments', props.content.IdCamp)"
          v-if="props.user == Acteur.Espaceperso"
          class="mr-2"
        >
          Lire
        </v-btn>
      </v-col>
    </v-row>
  </EventItem>
</template>

<script setup lang="ts">
import {
  type Event,
  type CampDocsEvt,
  type IdCamp,
  Acteur,
} from "@/clients/backoffice/logic/api";

const props = defineProps<{
  event: Event;
  content: CampDocsEvt;
  user: Acteur;
}>();

const emit = defineEmits<{ (e: "goToDocuments", idCamp: IdCamp): void }>();
</script>

<style scoped></style>
