<template>
  <EventItem :color="icon.color" :icon="icon.name" :time="props.event.Created">
    <v-row no-gutters>
      <v-col align-self="center">
        Inscription de <i>{{ props.content.Participant }}</i> pour le séjour
        <i>{{ props.content.OriginCamp }}</i>
        <template v-if="props.content.Statut == StatutParticipant.Inscrit">
          <b> acceptée</b>
        </template>
        <template v-else-if="props.content.Statut == StatutParticipant.Refuse">
          <b> refusée définitivement</b>
        </template>
        <template v-else>
          <b> placée en liste d'attente</b>
        </template>
        {{ props.content.IsBackoffice ? `par le centre` : `par la direction` }}.
      </v-col>
      <v-col align-self="center" cols="auto">
        <v-btn
          v-if="props.user == Acteur.Espaceperso"
          icon="mdi-information"
          flat
          class="mr-2"
          @click="emit('goToValidation')"
        ></v-btn>
      </v-col>
    </v-row>
  </EventItem>
</template>

<script setup lang="ts">
import {
  Acteur,
  StatutParticipant,
  type Event,
  type StatuationEvt,
} from "@/clients/backoffice/logic/api";
import { computed } from "vue";

const props = defineProps<{
  event: Event;
  content: StatuationEvt;
  user: Acteur;
}>();

const emit = defineEmits<{ (e: "goToValidation"): void }>();

const icon = computed(() => {
  switch (props.content.Statut) {
    case StatutParticipant.Inscrit:
      return { name: "mdi-map-marker-check", color: "green" };
    case StatutParticipant.Refuse:
      return { name: "mdi-map-marker-remove-variant", color: "red" };
    default:
      return { name: "mdi-map-marker-alert", color: "orange" };
  }
});
</script>

<style scoped></style>
