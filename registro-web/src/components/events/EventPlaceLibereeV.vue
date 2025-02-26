<template>
  <EventItem
    icon="mdi-information-outline"
    color="yellow"
    :time="props.event.Created"
  >
    Une place pour
    <a href="">
      {{ props.content.ParticipantLabel }}
    </a>
    s'est libérée sur le camp <a href=""> {{ props.content.CampLabel }} </a>.
  </EventItem>
</template>

<script setup lang="ts">
import { type Event, type PlaceLiberee } from "@/clients/backoffice/logic/api";

const props = defineProps<{
  event: Event;
  content: PlaceLiberee;
}>();
</script>

<style scoped></style>
