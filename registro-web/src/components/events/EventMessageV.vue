<template>
  <EventItem
    :icon="fromUs ? 'mdi-email-arrow-right' : 'mdi-email-arrow-left'"
    color="light-blue-darken-3"
    :time="props.event.Created"
    size="x-large"
  >
    <v-row no-gutters>
      <v-col align-self="center">
        <v-card :class="colorClass">
          <v-card-text class="pa-2">
            <pre style="white-space: pre-wrap">{{
              props.content.Message.Contenu
            }}</pre>
          </v-card-text>
        </v-card>
      </v-col>
      <v-col cols="auto" align-self="center">
        <v-btn
          v-if="allowDelete"
          icon
          size="x-small"
          class="ml-2"
          @click="showConfirme = true"
        >
          <v-icon color="red">mdi-delete</v-icon>
        </v-btn>
      </v-col>
    </v-row>
    <v-row
      no-gutters
      class="mt-2"
      v-if="
        props.user == Acteur.Espaceperso &&
        props.content.Message.Origine == Acteur.FondSoutien
      "
    >
      <v-spacer></v-spacer>
      <v-btn size="small" @click="emit('replyFondSoutien')">
        <template #prepend>
          <v-icon>mdi-reply</v-icon>
        </template>
        Répondre au fonds de soutien
      </v-btn>
    </v-row>

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
  (e: "replyFondSoutien"): void;
}>();

const colorClass = computed(() =>
  origineToColor(props.content.Message.Origine)
);

const allowDelete = computed(
  () =>
    props.user == Acteur.Backoffice ||
    (props.user == Acteur.FondSoutien && fromUs)
);

function origineToColor(or: Acteur) {
  switch (or) {
    case Acteur.Backoffice:
      return "bg-light-green-lighten-3";
    case Acteur.FondSoutien:
      return "bg-yellow-darken-1";
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
