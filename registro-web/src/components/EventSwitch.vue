<template>
  <EventInscriptionTimeV
    v-if="props.event.Kind == 'inscription-time'"
    :time="props.event.Time"
  ></EventInscriptionTimeV>
  <EventPaiementV
    v-else-if="props.event.Kind == 'paiement'"
    :paiement="props.event.Paiement"
    @edit="emit('editPaiement', props.event.Paiement)"
  >
  </EventPaiementV>
  <EventSupprimeV
    v-else-if="
      props.event.Kind == 'event' &&
      props.event.event.Content.Kind == EventContentKind.Supprime
    "
    :event="props.event.event"
    :content="props.event.event.Content.Data"
  ></EventSupprimeV>
  <EventAccuseReceptionV
    v-else-if="
      props.event.Kind == 'event' &&
      props.event.event.Content.Kind == EventContentKind.AccuseReception
    "
    :event="props.event.event"
    :content="props.event.event.Content.Data"
  ></EventAccuseReceptionV>
  <EventMessageV
    v-else-if="
      props.event.Kind == 'event' &&
      props.event.event.Content.Kind == EventContentKind.Message
    "
    :event="props.event.event"
    :content="props.event.event.Content.Data"
    @delete="emit('deleteMessage', props.event.event)"
  ></EventMessageV>
  <EventPlaceLibereeV
    v-else-if="
      props.event.Kind == 'event' &&
      props.event.event.Content.Kind == EventContentKind.PlaceLiberee
    "
    :event="props.event.event"
    :content="props.event.event.Content.Data"
  ></EventPlaceLibereeV>
  <EventFactureV
    v-else-if="
      props.event.Kind == 'event' &&
      props.event.event.Content.Kind == EventContentKind.Facture
    "
    :event="props.event.event"
    :content="props.event.event.Content.Data"
  ></EventFactureV>
  <EventCampDocsV
    v-else-if="
      props.event.Kind == 'event' &&
      props.event.event.Content.Kind == EventContentKind.CampDocs
    "
    :event="props.event.event"
    :content="props.event.event.Content.Data"
  ></EventCampDocsV>
  <EventAttestationV
    v-else-if="
      props.event.Kind == 'event' &&
      props.event.event.Content.Kind == EventContentKind.Attestation
    "
    :event="props.event.event"
    :content="props.event.event.Content.Data"
  ></EventAttestationV>
  <EventSondageV
    v-else-if="
      props.event.Kind == 'event' &&
      props.event.event.Content.Kind == EventContentKind.Sondage
    "
    :event="props.event.event"
    :content="props.event.event.Content.Data"
  ></EventSondageV>
</template>

<script setup lang="ts">
import {
  EventContentKind,
  type Event,
  type Paiement,
} from "@/clients/backoffice/logic/api";
import EventMessageV from "./events/EventMessageV.vue";
import EventPlaceLibereeV from "./events/EventPlaceLibereeV.vue";
import EventSupprimeV from "./events/EventSupprimeV.vue";
import EventAccuseReceptionV from "./events/EventAccuseReceptionV.vue";
import EventFactureV from "./events/EventFactureV.vue";
import EventCampDocsV from "./events/EventCampDocsV.vue";
import EventAttestationV from "./events/EventAttestationV.vue";
import EventSondageV from "./events/EventSondageV.vue";
import type { PseudoEvent } from "@/utils";

const props = defineProps<{
  event: PseudoEvent;
}>();

const emit = defineEmits<{
  (e: "editPaiement", paiement: Paiement): void;
  (e: "deleteMessage", event: Event): void;
}>();
</script>

<style scoped></style>
