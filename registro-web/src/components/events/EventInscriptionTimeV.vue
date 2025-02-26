<template>
  <EventItem color="grey" icon="mdi-map-marker" :time="props.time">
    Inscription effectuée.
  </EventItem>
</template>

<script setup lang="ts">
import { type Time } from "@/clients/backoffice/logic/api";
import EventItem from "./EventItem.vue";

const props = defineProps<{
  time: Time;
}>();
</script>

<style scoped></style>
