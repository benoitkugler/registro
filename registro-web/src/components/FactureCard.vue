<template>
  <v-card
    title="Détails du réglement"
    :subtitle="
      listeAttente.length
        ? `Les participants en liste d'attente (${listeAttente.join(
            ', ',
          )}) ne sont pas pris en compte.`
        : ''
    "
    :min-width="hideOptionsAndAides ? '500px' : '900px'"
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
          >Aide extérieure</v-col
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
            >{{ Formatters.montant(part.Bilan.PrixCamp) }}</v-col
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
            <span v-if="!part.Bilan.Aide.Valid">-</span>
            <v-chip v-else label size="small">
              {{ part.Bilan.Aide.Structure }} :
              {{ Formatters.montant(part.Bilan.Aide.Montant) }}
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
      <v-row class="mt-0">
        <v-col cols="10" class="text-right">Total demandé</v-col>
        <v-col cols="2" class="text-right">{{
          props.dossier.Bilan.Demande
        }}</v-col>
      </v-row>

      <v-divider thickness="1" class="my-2"></v-divider>
      <!-- Justificatifs en attente de validation -->
      <template v-if="showJustificatifsAttente && pendingJustificatifs.length">
        <v-list-subheader>
          Justificatifs en attente de validation
        </v-list-subheader>
        <v-list density="compact">
          <v-list-item
            v-for="pending in pendingJustificatifs"
            :key="pending.file.Key"
            density="compact"
            :title="justificatifLabel(pending.id)"
            :subtitle="pending.file.NomClient"
          >
            <template #append>
              <ReviewStatusChip
                :status="pending.file.ReviewStatus"
              ></ReviewStatusChip>
              <v-btn
                icon="mdi-download"
                size="small"
                variant="text"
                :href="endpoints.LoadDocument(pending.file.Key)"
                @click.stop=""
                target="_blank"
              ></v-btn>
              <v-btn
                v-if="canDeleteJustificatifsAttente"
                icon="mdi-delete"
                size="small"
                variant="text"
                color="red"
                @click.stop="emit('deleteJustificatif', pending.id)"
              ></v-btn>
            </template>
          </v-list-item>
        </v-list>
      </template>

      <!-- Attestation de quotient familial -->
      <template v-if="props.dossier.HasQuotientFamilialFile">
        <v-divider thickness="1" class="my-2"></v-divider>
        <v-list-subheader>Attestation de quotient familial</v-list-subheader>
        <v-list density="compact">
          <v-list-item
            density="compact"
            :title="props.dossier.QuotientFamilialFile.NomClient"
            :subtitle="`${Formatters.size(props.dossier.QuotientFamilialFile.Taille)} - Ajouté ${Formatters.date(props.dossier.QuotientFamilialFile.Uploaded)}`"
          >
            <template #append>
              <ReviewStatusChip
                :status="props.dossier.QuotientFamilialFile.ReviewStatus"
              ></ReviewStatusChip>

              <v-btn
                icon="mdi-download"
                size="small"
                variant="text"
                :href="
                  endpoints.LoadDocument(props.dossier.QuotientFamilialFile.Key)
                "
                @click.stop=""
                target="_blank"
              ></v-btn>
              <v-tooltip
                v-if="canDeleteQuotientFamilial"
                :disabled="!quotientFamilialValidated"
                :text="quotientFamilialValidatedMessage"
              >
                <template #activator="{ props: tooltipProps }">
                  <span v-bind="tooltipProps">
                    <v-btn
                      icon="mdi-delete"
                      size="small"
                      variant="text"
                      color="red"
                      :disabled="quotientFamilialValidated"
                      @click.stop="emit('deleteQuotientFamilial')"
                    ></v-btn>
                  </span>
                </template>
              </v-tooltip>
            </template>
          </v-list-item>
        </v-list>
      </template>

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
  FileReviewStatus,
  FileReviewStatusLabels,
  StatutParticipant,
  type BilanParticipantPub,
  type DossierExt,
  type IdParticipant,
  type PublicFile,
  type Remises,
} from "@/clients/backoffice/logic/api";
import { Camps, endpoints, Formatters, Participants, Personnes } from "@/utils";
import { computed } from "vue";

const props = defineProps<{
  dossier: DossierExt;
  showJustificatifsAttente: boolean;
  canDeleteJustificatifsAttente?: boolean;
  canDeleteQuotientFamilial?: boolean;
}>();

const emit = defineEmits<{
  (e: "deleteJustificatif", id: IdParticipant): void;
  (e: "deleteQuotientFamilial"): void;
}>();

const listeAttente = computed(() =>
  (props.dossier.Participants || [])
    .filter((p) => p.Participant.Statut != StatutParticipant.Inscrit)
    .map((p) => Personnes.label(p.Personne)),
);

const quotientFamilialValidatedMessage =
  "Attestation validée : demandez au backoffice de la refuser avant d'en envoyer une nouvelle.";
const quotientFamilialValidated = computed(
  () =>
    props.dossier.HasQuotientFamilialFile &&
    props.dossier.QuotientFamilialFile.ReviewStatus ===
      FileReviewStatus.FileReviewValidated,
);

const participants = computed(() =>
  (props.dossier.Participants || [])
    .filter((p) => p.Participant.Statut == StatutParticipant.Inscrit)
    .map((p) => ({
      Label: Personnes.label(p.Personne),
      Camp: Camps.label(p.Camp),
      Bilan: props.dossier.Bilan.Inscrits![p.Participant.Id], // licite pour un Inscrit
    })),
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

const pendingJustificatifs = computed<
  { id: IdParticipant; file: PublicFile }[]
>(() => {
  const out: { id: IdParticipant; file: PublicFile }[] = [];
  for (const [id, file] of Object.entries(
    props.dossier.AidesFilesAttente || {},
  )) {
    out.push({ id: Number(id) as unknown as IdParticipant, file });
  }
  return out;
});

function justificatifLabel(id: IdParticipant): string {
  const p = (props.dossier.Participants || []).find(
    (p) => p.Participant.Id === id,
  );
  if (!p) return "";
  return `${Personnes.label(p.Personne)} - ${Camps.label(p.Camp)}`;
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
      Participants.montantEquals(part.Bilan.PrixCamp, part.Bilan.AvecOption) &&
      !part.Bilan.Aide.Valid,
  ),
);
</script>
