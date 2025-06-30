<template>
  <EventItem
    :icon="fromUs ? 'mdi-email-arrow-right' : 'mdi-email-arrow-left'"
    color="light-blue-darken-3"
    :time="props.event.Created"
    size="x-large"
  >
    <v-dialog v-if="showConfirme" v-model="showConfirme" max-width="400px">
      <v-card title="Confirmation">
        <v-card-text>
          Confirmez-vous la suppression de ce message ?
          <br /><br />
          Attention, cette opération est irréversible.
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="red" @click="emit('delete')">Supprimer</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-row no-gutters>
      <v-col align-self="center" cols="11">
        <v-card :class="colorClass">
          <v-card-text class="pa-1">
            <pre style="white-space: pre-wrap">{{
              props.content.Message.Contenu
            }}</pre>
          </v-card-text>
        </v-card>
      </v-col>
      <v-col cols="1" align-self="center">
        <v-btn
          v-if="allowDelete"
          icon
          size="x-small"
          class="ml-2"
          @click="showConfirme = true"
        >
          <v-icon color="red">mdi-delete</v-icon>
        </v-btn>
        <div v-else>
          <!-- invisible  -->
          <v-btn icon size="x-small" variant="text" disabled></v-btn>
        </div>
      </v-col>
    </v-row>
  </EventItem>
</template>

<script setup lang="ts">
import {
  Acteur,
  type Event,
  type Message,
} from "@/clients/backoffice/logic/api";
import EventItem from "./EventItem.vue";
import { computed, ref } from "vue";

const props = defineProps<{
  event: Event;
  content: Message;
  user: Acteur;
}>();

const emit = defineEmits<{
  (e: "delete"): void;
}>();

const colorClass = computed(() =>
  origineToColor(props.content.Message.Origine)
);

const allowDelete = computed(
  () =>
    props.user == Acteur.Backoffice ||
    (props.user == Acteur.Fondsoutien && fromUs)
);

function origineToColor(or: Acteur) {
  switch (or) {
    case Acteur.Backoffice:
      return "bg-light-green-lighten-3";
    case Acteur.Backoffice:
      return "bg-orange-lighten-3";
    case Acteur.Directeur:
      return "bg-lime";
    case Acteur.Espaceperso:
      return "bg-light-blue-lighten-3";
  }
}

const showConfirme = ref(false);

/** true if message is sent by the current user */
const fromUs = computed(() => props.content.Message.Origine == props.user);
</script>

<style scoped></style>
