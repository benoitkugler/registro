<template>
  <EventItem
    color="orange-darken-1"
    icon="mdi-invoice"
    :time="props.event.Created"
  >
    Demande de règlement envoyée.
    <i v-if="props.content.IsRappel">Il s'agit d'un rappel.</i>
  </EventItem>
</template>

<script setup lang="ts">
import { type Event, type FactureEvt } from "@/clients/backoffice/logic/api";

const props = defineProps<{
  event: Event;
  content: FactureEvt;
}>();
</script>

<style scoped></style>
