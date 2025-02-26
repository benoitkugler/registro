<template>
  <EventItem
    color="purple"
    icon="mdi-chat-question"
    :time="props.event.Created"
  >
    Un sondage concernant le camp
    <a href="">
      {{ props.content.CampLabel }}
    </a>
    a été envoyé.
  </EventItem>
</template>

<script setup lang="ts">
import { type Event, type Sondage } from "@/clients/backoffice/logic/api";

const props = defineProps<{
  event: Event;
  content: Sondage;
}>();
</script>

<style scoped></style>
