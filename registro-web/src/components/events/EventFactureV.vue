<template>
  <EventItem
    color="orange-darken-1"
    icon="mdi-invoice"
    :time="props.event.Created"
  >
    Facture envoyée
  </EventItem>
</template>

<script setup lang="ts">
import { type Event, type Facture } from "@/clients/backoffice/logic/api";

const props = defineProps<{
  event: Event;
  content: Facture;
}>();
</script>

<style scoped></style>
