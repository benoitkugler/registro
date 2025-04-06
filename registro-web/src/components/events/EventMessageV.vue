<template>
  <EventItem
    :icon="isSent ? 'mdi-email-arrow-right' : 'mdi-email-arrow-left'"
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
      <v-col align-self="center">
        <v-card :class="colorClass">
          <v-card-text class="pa-1">
            <pre>{{ props.content.Message.Contenu }}</pre>
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
  MessageOrigine,
  type Event,
  type Message,
} from "@/clients/backoffice/logic/api";
import EventItem from "./EventItem.vue";
import { computed, ref } from "vue";

const props = defineProps<{
  event: Event;
  content: Message;
}>();

const emit = defineEmits<{
  (e: "delete"): void;
}>();

const colorClass = computed(() =>
  origineToColor(props.content.Message.Origine)
);

const allowDelete = computed(
  () => props.content.Message.Origine == MessageOrigine.FromBackoffice
);

function origineToColor(or: MessageOrigine) {
  switch (or) {
    case MessageOrigine.FromBackoffice:
      return "bg-light-green-lighten-3";
    case MessageOrigine.FromDirecteur:
      return "bg-lime";
    case MessageOrigine.FromEspaceperso:
      return "bg-light-blue-lighten-3";
  }
}

const showConfirme = ref(false);

/** true is sent by us */
const isSent = computed(
  () => props.content.Message.Origine != MessageOrigine.FromEspaceperso
);
</script>

<style scoped></style>
