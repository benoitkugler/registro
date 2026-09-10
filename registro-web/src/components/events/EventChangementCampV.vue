<template>
  <EventItem
    icon="mdi-swap-horizontal"
    color="info"
    :time="props.event.Created"
  >
    <i>{{ props.content.ParticipantLabel }}</i> a changé de séjour, du séjour
    {{ props.content.OldCampLabel }} vers le séjour
    <b>{{ props.content.NewCampLabel }}</b
    >.
  </EventItem>
</template>

<script setup lang="ts">
import {
  Acteur,
  type Event,
  type ChangementCampEvt,
} from "@/clients/backoffice/logic/api";

const props = defineProps<{
  event: Event;
  content: ChangementCampEvt;
  user: Acteur;
}>();

// const emit = defineEmits<{ (e: "acceptPlaceLiberee"): void }>();
</script>

<style scoped></style>
