<template>
  <EventItem
    color="purple"
    icon="mdi-comment-quote"
    :time="props.event.Created"
  >
    <v-row>
      <v-col align-self="center">
        Un sondage concernant le séjour
        {{ props.content.CampLabel }}
        a été envoyé.
      </v-col>
      <v-col align-self="center" cols="auto">
        <v-btn size="small" @click="emit('goToSondage', props.content.IdCamp)">
          {{ props.user == Acteur.Espaceperso ? "répondre" : "aller" }}
        </v-btn>
      </v-col>
    </v-row>
  </EventItem>
</template>

<script setup lang="ts">
import {
  Acteur,
  type Event,
  type IdCamp,
  type Sondage,
} from "@/clients/backoffice/logic/api";

const props = defineProps<{
  event: Event;
  content: Sondage;
  user: Acteur;
}>();

const emit = defineEmits<{ (e: "goToSondage", idCamp: IdCamp): void }>();
</script>

<style scoped></style>
