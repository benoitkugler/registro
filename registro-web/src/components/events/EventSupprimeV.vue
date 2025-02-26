<template>
  <EventItem
    color="grey"
    icon="mdi-close"
    :time="props.event.Created"
    size="small"
  >
    Message supprimé
  </EventItem>
</template>

<script setup lang="ts">
import { type Event, type Supprime } from "@/clients/backoffice/logic/api";

const props = defineProps<{
  event: Event;
  content: Supprime;
}>();
</script>

<style scoped></style>
