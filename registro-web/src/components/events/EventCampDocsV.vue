<template>
  <EventItem icon="mdi-mail" color="light-blue" :time="props.event.Created">
    Les documents du séjour <a href=""> {{ props.content.CampLabel }} </a> ont
    été envoyés.
  </EventItem>
</template>

<script setup lang="ts">
import { type Event, type CampDocs } from "@/clients/backoffice/logic/api";

const props = defineProps<{
  event: Event;
  content: CampDocs;
}>();
</script>

<style scoped></style>
