<template>
  <EventItem
    color="green"
    icon="mdi-map-marker-check"
    :time="props.event.Created"
  >
    <v-row no-gutters>
      <v-col align-self="center">
        Inscription reçue et confirmée
        {{
          props.content.ByCamp
            ? `par la direction du séjour ${props.content.ByCamp}.`
            : `par le centre.`
        }}
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
  type Event,
  type ValidationEvt,
} from "@/clients/backoffice/logic/api";

const props = defineProps<{
  event: Event;
  content: ValidationEvt;
  user: Acteur;
}>();

const emit = defineEmits<{ (e: "goToValidation"): void }>();
</script>

<style scoped></style>
