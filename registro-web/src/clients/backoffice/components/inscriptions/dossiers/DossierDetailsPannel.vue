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
    <v-dialog v-model="showEditDossier" max-width="600px">
      <DossierEditCard
        :responsable="props.dossier.Dossier.Responsable"
        :dossier="props.dossier.Dossier.Dossier"
        @save="
          (v) => {
            showEditDossier = false;
            emit('updateDossier', v);
          }
        "
      ></DossierEditCard>
    </v-dialog>

    <v-dialog v-model="showEditParticipants" max-width="1200px">
      <DossierParticipantsCard
        :dossier="props.dossier.Dossier"
      ></DossierParticipantsCard>
    </v-dialog>

    <template v-slot:append>
      <v-menu>
        <template v-slot:activator="{ props: menuProps }">
          <v-btn v-bind="menuProps" icon="mdi-pencil"></v-btn>
        </template>
        <v-list density="compact">
          <v-list-item
            title="Modifier le dossier..."
            @click="showEditDossier = true"
          >
          </v-list-item>
          <v-list-item
            title="Modifier les participants..."
            @click="showEditParticipants = true"
          >
          </v-list-item>
        </v-list>
      </v-menu>
    </template>
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
import { computed, ref } from "vue";
import {
  StatutPaiement,
  type Dossier,
  type DossierDetails,
} from "../../../logic/api";
import { Personnes, pseudoEventTime, type PseudoEvent } from "@/utils";
import FactureCard from "./FactureCard.vue";
import DossierEditCard from "./DossierEditCard.vue";
import DossierParticipantsCard from "./DossierParticipantsCard.vue";

const props = defineProps<{
  dossier: DossierDetails;
}>();

const emit = defineEmits<{
  (e: "updateDossier", dossier: Dossier): void;
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

const showEditDossier = ref(false);
const showEditParticipants = ref(false);
</script>
