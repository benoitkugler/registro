<template>
  <v-card
    title="Suivi du règlement"
    :subtitle="`${completCount} / ${
      props.data.Participants?.length || 0
    } complet`"
  >
    <v-card-text>
      <v-list>
        <v-list-item
          v-for="participant in data.Participants"
          :title="Personnes.NOMPrenom(participant.Personne)"
          :subtitle="dossier(participant).Responsable"
        >
          <template #append>
            <v-chip
              prepend-icon="mdi-currency-eur"
              :color="
                Formatters.colorStatutPaiement(dossier(participant).Reglement)
              "
            >
              {{ StatutPaiementLabels[dossier(participant).Reglement] }}
            </v-chip>
          </template>
        </v-list-item>
      </v-list>
    </v-card-text>
  </v-card>
</template>

<script lang="ts" setup>
import { Formatters, Personnes } from "@/utils";
import {
  StatutPaiement,
  StatutPaiementLabels,
  type ParticipantExt,
} from "@/clients/backoffice/logic/api";
import { computed } from "vue";
import type { ParticipantsOut } from "@/clients/directeurs/logic/api";

const props = defineProps<{
  data: Pick<ParticipantsOut, "Dossiers" | "Participants">;
}>();

function dossier(p: ParticipantExt) {
  return (props.data.Dossiers || {})[p.Participant.IdDossier];
}

const completCount = computed(
  () =>
    (props.data.Participants || []).filter(
      (p) => dossier(p).Reglement == StatutPaiement.Complet
    ).length
);
</script>
