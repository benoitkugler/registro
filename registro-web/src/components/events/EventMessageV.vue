<template>
  <EventItem
    icon="mdi-email"
    color="light-blue-darken-3"
    :time="props.event.Created"
    size="x-large"
  >
    <v-card>
      <v-card-text class="pa-1">
        <div
          v-for="(line, i) in props.content.Message.Contenu.split('\n')"
          :key="i"
        >
          {{ line }}
        </div>
      </v-card-text>
    </v-card>
  </EventItem>
</template>

<script setup lang="ts">
import { type Event, type Message } from "@/clients/backoffice/logic/api";
import EventItem from "./EventItem.vue";

const props = defineProps<{
  event: Event;
  content: Message;
}>();
</script>

<style scoped></style>
