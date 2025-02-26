<template>
  <EventItem
    color="green"
    icon="mdi-map-marker-check"
    :time="props.event.Created"
  >
    Accusé de réception envoyé
  </EventItem>
</template>

<script setup lang="ts">
import {
  type Event,
  type AccuseReception,
} from "@/clients/backoffice/logic/api";

const props = defineProps<{
  event: Event;
  content: AccuseReception;
}>();
</script>

<style scoped></style>
