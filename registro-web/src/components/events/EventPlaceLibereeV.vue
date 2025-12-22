<template>
  <EventItem
    icon="mdi-information-outline"
    color="yellow"
    :time="props.event.Created"
  >
    <v-row>
      <v-col align-self="center">
        Une place pour
        <i>
          {{ props.content.ParticipantLabel }}
        </i>
        s'est libérée sur le séjour {{ props.content.CampLabel }}.
      </v-col>
      <v-col cols="auto" align-self="center">
        <v-btn
          size="small"
          color="green"
          v-if="props.user == Acteur.Espaceperso && !props.content.Accepted"
          @click="emit('acceptPlaceLiberee')"
          prepend-icon="mdi-check"
        >
          Accepter la place</v-btn
        >
        <v-chip v-if="props.content.Accepted" prepend-icon="mdi-check">
          Acceptée
        </v-chip>
      </v-col>
    </v-row>
  </EventItem>
</template>

<script setup lang="ts">
import {
  Acteur,
  type Event,
  type PlaceLibereeEvt,
} from "@/clients/backoffice/logic/api";

const props = defineProps<{
  event: Event;
  content: PlaceLibereeEvt;
  user: Acteur;
}>();

const emit = defineEmits<{ (e: "acceptPlaceLiberee"): void }>();
</script>

<style scoped></style>
