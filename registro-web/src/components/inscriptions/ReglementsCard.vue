<template>
  <v-card
    title="Suivi du règlement"
    :subtitle="`${completCount} / ${filtered.length || 0} complet`"
  >
    <v-card-text>
      <v-list>
        <v-list-item
          v-for="participant in filtered"
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
  StatutParticipant,
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

const filtered = computed(() =>
  (props.data.Participants || []).filter(
    (p) => p.Participant.Statut == StatutParticipant.Inscrit
  )
);

const completCount = computed(
  () =>
    filtered.value.filter((p) => dossier(p).Reglement == StatutPaiement.Complet)
      .length
);
</script>
