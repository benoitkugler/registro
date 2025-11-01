<template>
  <v-card
    title="Suivi financier"
    :subtitle="
      listeAttente.length
        ? `Les participants en liste d'attente (${listeAttente.join(
            ', '
          )}) ne sont pas pris en compte.`
        : ''
    "
    min-width="900px"
  >
    <v-card-text>
      <v-row>
        <v-col cols="3">Participant</v-col>
        <v-col cols="2" class="text-center">Prix du séjour</v-col>
        <v-col cols="2" class="text-center">Prix avec option</v-col>
        <v-col cols="3" class="text-center">Aides extérieures</v-col>
        <v-col cols="2" class="text-right">Sous-total</v-col>
      </v-row>
      <v-divider thickness="1" class="my-2"></v-divider>
      <template v-for="part in participants">
        <v-row>
          <v-col align-self="center" cols="3">
            {{ part.Label }} <br />
            <span class="text-grey">
              {{ part.Camp }}
            </span>
          </v-col>
          <v-col align-self="center" cols="2" class="text-center">{{
            part.Prix
          }}</v-col>
          <v-col align-self="center" cols="2" class="text-center"
            >{{ Formatters.montant(part.Bilan.AvecOption) }} <br />
            <span class="text-grey">
              {{ part.Bilan.AvecOptionDescription }}
            </span>
          </v-col>
          <v-col align-self="center" cols="3" class="text-center">
            <span v-if="!part.Bilan.Aides?.length">-</span>
            <v-chip v-for="aide in part.Bilan.Aides">
              {{ aide.Structure }} : {{ Formatters.montant(aide.Montant) }}
            </v-chip>
          </v-col>
          <v-col align-self="center" cols="2" class="text-right">
            {{ part.Bilan.AvecAides }}
          </v-col>
        </v-row>
        <v-row v-if="hasRemise(part.Bilan)" no-gutters>
          <v-col cols="10" class="text-right">
            <i>
              {{ formatRemises(part.Bilan.Remises) }}
            </i>
          </v-col>
          <v-col cols="2" class="text-right">
            {{ part.Bilan.Net }}
          </v-col>
        </v-row>
      </template>

      <v-divider thickness="1" class="my-2"></v-divider>

      <!-- Total avant paiements -->
      <v-row>
        <v-col cols="10" class="text-right">Total demandé</v-col>
        <v-col cols="2" class="text-right">{{
          props.dossier.Bilan.Demande
        }}</v-col>
      </v-row>

      <v-divider thickness="1" class="my-2"></v-divider>

      <!-- Paiements -->
      <v-row v-for="paiement in paiements">
        <v-col>
          Paiement de {{ paiement.Payeur }}, le
          {{ Formatters.date(paiement.Time) }}
        </v-col>
        <v-col cols="2" class="text-right">
          {{ paiement.IsRemboursement ? "+" : "-"
          }}{{ Formatters.montant(paiement.Montant) }}
        </v-col>
      </v-row>
      <!-- Solde -->
      <v-row>
        <v-col cols="10" class="text-right">Solde</v-col>
        <v-col cols="2" class="text-right">
          <b>
            {{ props.dossier.Bilan.Restant }}
          </b>
        </v-col>
      </v-row>
    </v-card-text>
  </v-card>
</template>

<script setup lang="ts">
import {
  StatutParticipant,
  type BilanParticipantPub,
  type DossierExt,
  type Remises,
} from "@/clients/backoffice/logic/api";
import { newDate_ } from "@/components/date";
import { Camps, Formatters, Personnes } from "@/utils";
import { computed } from "vue";

const props = defineProps<{
  dossier: DossierExt;
}>();

const listeAttente = computed(() =>
  (props.dossier.Participants || [])
    .filter((p) => p.Participant.Statut != StatutParticipant.Inscrit)
    .map((p) => Personnes.label(p.Personne))
);

const participants = computed(() =>
  (props.dossier.Participants || [])
    .filter((p) => p.Participant.Statut == StatutParticipant.Inscrit)
    .map((p) => ({
      Label: Personnes.label(p.Personne),
      Camp: Camps.label(p.Camp),
      Prix: Formatters.montant(p.Camp.Prix),
      Bilan: props.dossier.Bilan.Inscrits![p.Participant.Id], // licite pour un Inscrit
    }))
);

function hasRemise(b: BilanParticipantPub) {
  return !!(
    b.Remises.ReducInscrits ||
    b.Remises.ReducEquipiers ||
    b.Remises.ReducSpeciale.Cent
  );
}

function formatRemises(remises: Remises) {
  const chunks = [];
  if (remises.ReducInscrits) chunks.push(`Enfants : ${remises.ReducInscrits}%`);
  if (remises.ReducEquipiers)
    chunks.push(`Enfants : ${remises.ReducEquipiers}%`);
  if (remises.ReducSpeciale.Cent)
    chunks.push(`Remise : ${Formatters.montant(remises.ReducSpeciale)}`);
  return chunks.join("  ;  ");
}

const paiements = computed(() => {
  const out = Object.values(props.dossier.Paiements || {});
  out.sort((a, b) => new Date(a.Time).valueOf() - new Date(b.Time).valueOf());
  return out;
});
</script>
