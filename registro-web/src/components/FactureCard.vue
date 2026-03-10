<template>
  <v-card
    title="Détails du réglement"
    :subtitle="
      listeAttente.length
        ? `Les participants en liste d'attente (${listeAttente.join(
            ', '
          )}) ne sont pas pris en compte.`
        : ''
    "
    :min-width="hideOptionsAndAides ? '400px' : '900px'"
  >
    <v-card-text>
      <v-row>
        <v-col :cols="hideOptionsAndAides ? 8 : 3">Participant</v-col>
        <v-col cols="2" class="text-center" v-if="!hideOptionsAndAides"
          >Prix du séjour</v-col
        >
        <v-col cols="2" class="text-center" v-if="!hideOptionsAndAides"
          >Prix avec option</v-col
        >
        <v-col cols="3" class="text-center" v-if="!hideOptionsAndAides"
          >Aides extérieures</v-col
        >
        <v-col :cols="hideOptionsAndAides ? 4 : 2" class="text-right"
          >Sous-total</v-col
        >
      </v-row>
      <v-divider thickness="1" class="my-1"></v-divider>
      <template v-for="part in participants">
        <v-row no-gutters class="my-2">
          <v-col align-self="center" :cols="hideOptionsAndAides ? 8 : 3">
            {{ part.Label }} <br />
            <span class="text-grey">
              {{ part.Camp }}
            </span>
          </v-col>
          <v-col
            align-self="center"
            cols="2"
            class="text-center"
            v-if="!hideOptionsAndAides"
            >{{ Formatters.montant(part.Prix) }}</v-col
          >
          <v-col
            v-if="!hideOptionsAndAides"
            align-self="center"
            cols="2"
            class="text-center"
            >{{ Formatters.montant(part.Bilan.AvecOption) }} <br />
            <span class="text-grey">
              {{ part.Bilan.AvecOptionDescription }}
            </span>
          </v-col>
          <v-col
            align-self="center"
            cols="3"
            class="text-center"
            v-if="!hideOptionsAndAides"
          >
            <span v-if="!part.Bilan.Aides?.length">-</span>
            <v-chip v-for="aide in part.Bilan.Aides" label size="small">
              {{ aide.Structure }} : {{ Formatters.montant(aide.Montant) }}
            </v-chip>
          </v-col>
          <v-col
            align-self="center"
            :cols="hideOptionsAndAides ? 4 : 2"
            class="text-right"
          >
            {{ part.Bilan.AvecAides }}
          </v-col>
        </v-row>

        <v-row v-if="hasRemise(part.Bilan)" no-gutters class="my-1">
          <v-col cols="10" class="text-right">
            <i>
              {{ formatRemises(part.Bilan.Remises) }}
            </i>
          </v-col>
          <v-col cols="2" class="text-right">
            {{ part.Bilan.Net }}
          </v-col>
        </v-row>

        <v-divider thickness="1"></v-divider>
      </template>

      <!-- Participant non validés -->
      <template v-if="dossier.Bilan.DemandeEnAttenteValidation">
        <v-row class="mt-0 text-grey">
          <v-col cols="auto">
            <v-icon>mdi-information-outline</v-icon>
          </v-col>
          <v-col>Séjours en attente de validation</v-col>
          <v-col cols="4" class="text-right">
            {{ dossier.Bilan.DemandeEnAttenteValidation }}
          </v-col>
        </v-row>

        <v-divider thickness="1" class="my-2"></v-divider>
      </template>

      <!-- Total avant paiements -->
      <v-row>
        <v-col cols="10" class="text-right">Total demandé</v-col>
        <v-col cols="2" class="text-right">{{
          props.dossier.Bilan.Demande
        }}</v-col>
      </v-row>

      <v-divider thickness="1" class="my-2"></v-divider>

      <!-- Paiements -->
      <v-row v-if="!paiements.length">
        <v-col class="text-right">
          <i>Aucun paiement pour l'instant.</i>
        </v-col>
      </v-row>
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

      <v-divider thickness="1" class="my-2"></v-divider>

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
import { Camps, Formatters, Participants, Personnes } from "@/utils";
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
      Prix: p.Camp.Prix, // prix de base du séjour
      Bilan: props.dossier.Bilan.Inscrits![p.Participant.Id], // licite pour un Inscrit
    }))
);

function hasRemise(b: BilanParticipantPub) {
  return !!(
    b.Remises.Famille ||
    b.Remises.Equipiers ||
    b.Remises.Speciale.Cent
  );
}

function formatRemises(remises: Remises) {
  const chunks = [];
  if (remises.Famille) chunks.push(`Remise famille : ${remises.Famille}%`);
  if (remises.Equipiers)
    chunks.push(`Remise équipiers : ${remises.Equipiers}%`);
  if (remises.Speciale.Cent)
    chunks.push(`Remise : ${Formatters.montant(remises.Speciale)}`);
  return chunks.join("  ;  ");
}

const paiements = computed(() => {
  const out = Object.values(props.dossier.Paiements || {});
  out.sort((a, b) => new Date(a.Time).valueOf() - new Date(b.Time).valueOf());
  return out;
});

// hide the column if all the values between base and net are the same
const hideOptionsAndAides = computed(() =>
  participants.value.every(
    (part) =>
      Participants.montantEquals(part.Prix, part.Bilan.AvecOption) &&
      !part.Bilan.Aides?.length
  )
);
</script>
