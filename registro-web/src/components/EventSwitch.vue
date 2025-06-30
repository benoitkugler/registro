<template>
  <EventInscriptionTimeV
    v-if="props.event.Kind == 'inscription-time'"
    :time="props.event.Time"
  ></EventInscriptionTimeV>
  <EventPaiementV
    v-else-if="props.event.Kind == 'paiement'"
    :paiement="props.event.Paiement"
    :user="props.event.User"
    @edit="emit('editPaiement', props.event.Paiement)"
  >
  </EventPaiementV>
  <EventSupprimeV
    v-else-if="
      props.event.Kind == 'event' &&
      props.event.Event.Content.Kind == EventContentKind.Supprime
    "
    :event="props.event.Event"
    :content="props.event.Event.Content.Data"
  ></EventSupprimeV>
  <EventValidationV
    v-else-if="
      props.event.Kind == 'event' &&
      props.event.Event.Content.Kind == EventContentKind.Validation
    "
    :event="props.event.Event"
    :content="props.event.Event.Content.Data"
    @go-to-validation="emit('goToValidation')"
  ></EventValidationV>
  <EventMessageV
    v-else-if="
      props.event.Kind == 'event' &&
      props.event.Event.Content.Kind == EventContentKind.Message
    "
    :event="props.event.Event"
    :content="props.event.Event.Content.Data"
    :user="props.event.User"
    @delete="emit('deleteMessage', props.event.Event)"
  ></EventMessageV>
  <EventPlaceLibereeV
    v-else-if="
      props.event.Kind == 'event' &&
      props.event.Event.Content.Kind == EventContentKind.PlaceLiberee
    "
    :event="props.event.Event"
    :content="props.event.Event.Content.Data"
    :user="props.event.User"
    @accept-place-liberee="emit('acceptPlaceLiberee', props.event.Event.Id)"
  ></EventPlaceLibereeV>
  <EventFactureV
    v-else-if="
      props.event.Kind == 'event' &&
      props.event.Event.Content.Kind == EventContentKind.Facture
    "
    :event="props.event.Event"
    :content="props.event.Event.Content.Data"
  ></EventFactureV>
  <EventCampDocsV
    v-else-if="
      props.event.Kind == 'event' &&
      props.event.Event.Content.Kind == EventContentKind.CampDocs
    "
    :event="props.event.Event"
    :content="props.event.Event.Content.Data"
    :user="props.event.User"
    @go-to-documents="(id) => emit('goToDocuments', id)"
  ></EventCampDocsV>
  <EventAttestationV
    v-else-if="
      props.event.Kind == 'event' &&
      props.event.Event.Content.Kind == EventContentKind.Attestation
    "
    :event="props.event.Event"
    :content="props.event.Event.Content.Data"
  ></EventAttestationV>
  <EventSondageV
    v-else-if="
      props.event.Kind == 'event' &&
      props.event.Event.Content.Kind == EventContentKind.Sondage
    "
    :event="props.event.Event"
    :content="props.event.Event.Content.Data"
    :user="props.event.User"
    @go-to-sondage="(id) => emit('goToSondage', id)"
  ></EventSondageV>
</template>

<script setup lang="ts">
import {
  EventContentKind,
  type Event,
  type IdCamp,
  type IdEvent,
  type Paiement,
} from "@/clients/backoffice/logic/api";
import EventMessageV from "./events/EventMessageV.vue";
import EventPlaceLibereeV from "./events/EventPlaceLibereeV.vue";
import EventSupprimeV from "./events/EventSupprimeV.vue";
import EventValidationV from "./events/EventValidationV.vue";
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
  (e: "goToSondage", idCamp: IdCamp): void;
  (e: "goToDocuments", idCamp: IdCamp): void;
  (e: "goToValidation"): void;
  (e: "acceptPlaceLiberee", idEvent: IdEvent): void;
}>();
</script>

<style scoped></style>
