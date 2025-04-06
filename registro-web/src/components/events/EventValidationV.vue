<template>
  <EventItem
    color="green"
    icon="mdi-map-marker-check"
    :time="props.event.Created"
  >
    Inscription confirmée
    <v-tooltip
      :text="
        'Inscription reçue et confirmée ' +
        (props.content.ByCamp
          ? `par la direction du séjour ${props.content.ByCamp}.`
          : `par le centre.`)
      "
    >
      <template #activator="{ props: tooltipProps }">
        <v-chip v-bind="tooltipProps" prepend-icon="mdi-account">
          {{ props.content.ByCamp ? props.content.ByCamp : "Centre" }}
        </v-chip>
      </template>
    </v-tooltip>
  </EventItem>
</template>

<script setup lang="ts">
import { type Event, type Validation } from "@/clients/backoffice/logic/api";

const props = defineProps<{
  event: Event;
  content: Validation;
}>();
</script>

<style scoped></style>
