<template>
  <v-tooltip :text="StatutParticipantLabels[props.statut]">
    <template #activator="{ props: tooltipProps }">
      <v-icon
        :color="Formatters.statutParticipant(props.statut).color"
        v-bind="tooltipProps"
      >
        {{ Formatters.statutParticipant(props.statut).icon }}
      </v-icon>
    </template>
  </v-tooltip>
</template>

<script setup lang="ts">
import {
  StatutParticipant,
  StatutParticipantLabels,
  type Participant,
} from "@/clients/backoffice/logic/api";
import { Formatters } from "@/utils";
const props = defineProps<{
  statut: StatutParticipant;
}>();
// const emit = defineEmits<{
//   (e: "save", camp: Participant): void;
// }>();
</script>
