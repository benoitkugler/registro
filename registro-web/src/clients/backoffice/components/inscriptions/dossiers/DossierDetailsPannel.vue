<template>
  <v-card
    :title="`Dossier de ${props.dossier.Dossier.Responsable}`"
    :subtitle="
      props.dossier.Dossier.Participants?.map((p) =>
        Personnes.label(p.Personne)
      ).join(', ')
    "
    class="ml-2"
  >
    <v-card-text>
      <!-- récap financier -->
      <v-row>
        <v-col class="ml-2">
          <v-menu>
            <template v-slot:activator="{ props: menuProps }">
              <v-chip
                v-bind="menuProps"
                prepend-icon="mdi-currency-eur"
                :color="statutColor(props.dossier.Dossier.Bilan.Statut)"
              >
                {{ props.dossier.Dossier.Bilan.Recu }} payé sur
                {{ props.dossier.Dossier.Bilan.Demande }}
              </v-chip>
            </template>
            <FactureCard :dossier="props.dossier.Dossier"></FactureCard>
          </v-menu>
        </v-col>
      </v-row>
      <!-- fil des messages -->
      <v-timeline side="end" class="mt-4" density="compact">
        <EventSwitch
          :event="event"
          v-for="(event, i) in events"
          :key="i"
        ></EventSwitch>
      </v-timeline>
    </v-card-text>
  </v-card>
</template>

<script setup lang="ts">
import { computed } from "vue";
import {
  StatutPaiement,
  type DossierDetails,
  type IdentTarget,
} from "../../../logic/api";
import { Personnes, pseudoEventTime, type PseudoEvent } from "@/utils";
import FactureCard from "./FactureCard.vue";

const props = defineProps<{
  dossier: DossierDetails;
}>();

const emit = defineEmits<{
  (e: "identifie", params: IdentTarget): void;
}>();

function statutColor(s: StatutPaiement) {
  switch (s) {
    case StatutPaiement.NonCommence:
      return "red";
    case StatutPaiement.EnCours:
      return "orange";
    case StatutPaiement.Complet:
      return "green";
  }
}

// add the inscription time and paiements
// and sort by time
const events = computed(() => {
  const evList: PseudoEvent[] = (props.dossier.Dossier.Events || []).map(
    (ev) => ({
      Kind: "event",
      event: ev,
    })
  );
  const paiements: PseudoEvent[] = Object.values(
    props.dossier.Dossier.Paiements || {}
  ).map((p) => ({
    Kind: "paiement",
    Paiement: p,
  }));
  const out: PseudoEvent[] = [
    {
      Kind: "inscription-time",
      Time: props.dossier.Dossier.Dossier.MomentInscription,
    },
    ...evList,
    ...paiements,
  ];
  out.sort(
    (a, b) => pseudoEventTime(a).valueOf() - pseudoEventTime(b).valueOf()
  );
  return out;
});
</script>
