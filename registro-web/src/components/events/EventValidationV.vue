<template>
  <EventItem
    color="green"
    icon="mdi-map-marker-check"
    :time="props.event.Created"
  >
    <v-row no-gutters>
      <v-col align-self="center">
        Inscription reçue et confirmée
        {{
          props.content.ByCamp
            ? `par la direction du séjour ${props.content.ByCamp}.`
            : `par le centre.`
        }}
      </v-col>
      <v-col align-self="center" cols="auto">
        <v-btn
          icon="mdi-information"
          flat
          size="small"
          class="mr-2"
          @click="emit('goToValidation')"
        ></v-btn>
      </v-col>
    </v-row>
  </EventItem>
</template>

<script setup lang="ts">
import { type Event, type Validation } from "@/clients/backoffice/logic/api";

const props = defineProps<{
  event: Event;
  content: Validation;
}>();

const emit = defineEmits<{ (e: "goToValidation"): void }>();
</script>

<style scoped></style>
