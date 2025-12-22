<template>
  <EventItem
    icon="mdi-invoice-check"
    color="light-green-darken-1"
    :time="props.event.Created"
  >
    Une
    <b>
      {{
        props.content.IsPresence
          ? "attestation de présence"
          : "facture acquittée"
      }}
    </b>
    a été {{ action }}.
  </EventItem>
</template>

<script setup lang="ts">
import {
  type Event,
  type AttestationEvt,
  Distribution,
} from "@/clients/backoffice/logic/api";
import { computed } from "vue";

const props = defineProps<{
  event: Event;
  content: AttestationEvt;
}>();

const action = computed(() => {
  switch (props.content.Distribution) {
    case Distribution.DEspacePerso:
      return "téléchargée depuis l'espace personnel";
    case Distribution.DMail:
      return "envoyée";
    case Distribution.DMailAndDownloaded:
      return "envoyée et téléchargée";
  }
});
</script>

<style scoped></style>
