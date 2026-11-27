<template>
  <v-card subtitle="Paiement" class="mt-2">
    <template #append>
      <v-tooltip
        v-if="hasQuotientFamilialOption"
        :disabled="!qfValidated"
        :text="qfValidatedMessage"
      >
        <template #activator="{ props: tooltipProps }">
          <span v-bind="tooltipProps">
            <v-btn
              size="small"
              class="mr-1"
              :disabled="qfValidated"
              @click.stop="showQuotientFamilialUpload = true"
            >
              <template #prepend>
                <v-icon color="green">mdi-plus</v-icon>
              </template>
              Attestation QF
            </v-btn>
          </span>
        </template>
      </v-tooltip>
      <v-btn
        size="small"
        :disabled="!props.dossier.Participants?.length"
        @click.stop="showCreateAide = true"
        v-if="props.supportsAidesExt"
      >
        <template #prepend>
          <v-icon color="green">mdi-plus</v-icon>
        </template>
        Aide extérieure
      </v-btn>
    </template>

    <v-card-text class="pb-0">
      <v-card @click="showDetails = true" :elevation="0">
        <v-card-text class="pa-0">
          <v-list>
            <v-list-item title="Prix des séjours" density="compact">
              <template #append> {{ dossier.Bilan.Demande }} </template>
            </v-list-item>
            <v-list-item
              class="text-grey"
              title="Séjours en attente de validation"
              density="compact"
              v-if="dossier.Bilan.DemandeEnAttenteValidation"
            >
              <template #append>
                {{ dossier.Bilan.DemandeEnAttenteValidation }}
              </template>
            </v-list-item>

            <v-divider thickness="1"></v-divider>

            <v-list-item title="Paiements" density="compact">
              <template #append>
                {{ dossier.Bilan.Recu }}
              </template>
            </v-list-item>

            <template
              v-if="props.mode == RequiredPaiementState.PaiementAcompte"
            >
              <v-divider thickness="1"></v-divider>

              <v-list-item title="Acompte à régler" density="compact">
                <template #append>
                  <span class="font-weight-bold text-orange">
                    {{ dossier.Bilan.AcompteRestant }}
                  </span>
                </template>
              </v-list-item>
            </template>

            <v-divider thickness="1"></v-divider>

            <v-list-item title="Montant restant à régler" density="compact">
              <template #append>
                <span
                  :class="
                    dossier.Bilan.Statut == StatutPaiement.Complet
                      ? 'font-weight-bold text-green'
                      : 'font-weight-bold text-orange'
                  "
                >
                  {{ dossier.Bilan.Restant }}
                </span>
              </template>
            </v-list-item>

            <template v-if="props.supportsAidesExt && pendingFiles.length">
              <v-divider thickness="1"></v-divider>
              <v-list-subheader>
                Justificatifs en attente de validation
              </v-list-subheader>
              <v-list-item
                v-for="pending in pendingFiles"
                :key="pending.file.Key"
                density="compact"
                :title="participantLabel(pending.id)"
                :subtitle="pending.file.NomClient"
              >
                <template #append>
                  <v-btn
                    icon="mdi-download"
                    size="small"
                    variant="text"
                    :href="endpoints.LoadDocument(pending.file.Key)"
                    @click.stop=""
                    target="_blank"
                  ></v-btn>
                  <v-btn
                    icon="mdi-delete"
                    size="small"
                    variant="text"
                    color="red"
                    @click.stop="deleteAideJustificatif(pending.id)"
                  ></v-btn>
                </template>
              </v-list-item>
            </template>

            <template v-if="props.dossier.HasQuotientFamilialFile">
              <v-divider thickness="1"></v-divider>
              <v-list-subheader
                >Attestation de quotient familial</v-list-subheader
              >
              <v-list-item
                density="compact"
                :title="props.dossier.QuotientFamilialFile.NomClient"
                :subtitle="`${Formatters.size(props.dossier.QuotientFamilialFile.Taille)} - Ajouté ${Formatters.date(props.dossier.QuotientFamilialFile.Uploaded)}`"
              >
                <template #append>
                  <v-chip
                    class="mr-2"
                    size="small"
                    :color="
                      Formatters.reviewStatus(
                        props.dossier.QuotientFamilialFile.ReviewStatus,
                      ).color
                    "
                    :prepend-icon="
                      Formatters.reviewStatus(
                        props.dossier.QuotientFamilialFile.ReviewStatus,
                      ).icon
                    "
                  >
                    {{
                      FileReviewStatusLabels[
                        props.dossier.QuotientFamilialFile.ReviewStatus
                      ]
                    }}
                  </v-chip>
                  <v-btn
                    icon="mdi-download"
                    size="small"
                    variant="text"
                    :href="
                      endpoints.LoadDocument(
                        props.dossier.QuotientFamilialFile.Key,
                      )
                    "
                    @click.stop=""
                    target="_blank"
                  ></v-btn>
                  <v-tooltip
                    :disabled="!qfValidated"
                    :text="qfValidatedMessage"
                  >
                    <template #activator="{ props: tooltipProps }">
                      <span v-bind="tooltipProps">
                        <v-btn
                          icon="mdi-delete"
                          size="small"
                          variant="text"
                          color="red"
                          :disabled="qfValidated"
                          @click.stop="deleteQuotientFamilial"
                        ></v-btn>
                      </span>
                    </template>
                  </v-tooltip>
                </template>
              </v-list-item>
            </template>
          </v-list>
        </v-card-text>
      </v-card>
    </v-card-text>

    <v-dialog v-model="showQuotientFamilialUpload" max-width="450px">
      <v-card title="Attestation de quotient familial">
        <v-card-text>
          Téléversez une attestation CAF au format PDF. Un nouvel envoi remplace
          l'attestation précédente.
          <FileInput
            class="mt-2"
            label="Attestation"
            accept=".pdf"
            @update="quotientFamilialFile = $event"
          />
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn
            color="green"
            :disabled="!quotientFamilialFile"
            @click="uploadQuotientFamilial"
          >
            Envoyer
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <v-dialog v-model="showCreateAide" max-width="450px">
      <AideCard :dossier="props.dossier" @save="createAide"></AideCard>
    </v-dialog>

    <v-dialog v-model="showDetails">
      <FactureCard :dossier="dossier"></FactureCard>
    </v-dialog>

    <v-card-actions>
      <v-btn
        block
        variant="outlined"
        class="mr-1"
        @click.stop="emit('showReglement')"
        :disabled="props.mode == RequiredPaiementState.PaiementDesactive"
        prepend-icon="mdi-cash-plus"
      >
        {{
          props.mode == RequiredPaiementState.PaiementAcompte
            ? "Payer l'acompte"
            : "Payer"
        }}...
      </v-btn>
    </v-card-actions>
  </v-card>
</template>

<script setup lang="ts">
import {
  FileReviewStatus,
  FileReviewStatusLabels,
  RequiredPaiementState,
  StatutPaiement,
  type DossierExt,
  type IdParticipant,
  type PublicFile,
} from "../logic/api";
import { computed, ref } from "vue";
import AideCard from "./AideCard.vue";
import FileInput from "@/components/files/FileInput.vue";
import { controller } from "../logic/logic";
import { Camps, endpoints, Formatters, Personnes } from "@/utils";
const props = defineProps<{
  token: string;
  dossier: DossierExt;
  supportsAidesExt: boolean;
  mode: RequiredPaiementState;
}>();

const emit = defineEmits<{
  (e: "refresh"): void;
  (e: "showReglement"): void;
}>();

const showCreateAide = ref(false);
const showQuotientFamilialUpload = ref(false);
const quotientFamilialFile = ref<File | null>(null);

const hasQuotientFamilialOption = computed(
  () =>
    props.dossier.Participants?.some((p) =>
      Camps.isQuotientFamilialActive(p.CampQuotientFamilial),
    ) ?? false,
);

const qfValidatedMessage =
  "Attestation validée : demandez au backoffice de la refuser avant d'en envoyer une nouvelle.";
const qfValidated = computed(
  () =>
    props.dossier.HasQuotientFamilialFile &&
    props.dossier.QuotientFamilialFile.ReviewStatus ===
      FileReviewStatus.FileReviewValidated,
);

const pendingFiles = computed(() => {
  const out: { id: IdParticipant; file: PublicFile }[] = [];
  for (const [id, file] of Object.entries(
    props.dossier.AidesFilesAttente || {},
  )) {
    out.push({ id: Number(id) as unknown as IdParticipant, file });
  }
  return out;
});

function participantLabel(id: IdParticipant): string {
  const p = props.dossier.Participants?.find((p) => p.Participant.Id === id);
  if (!p) return "";
  return `${Personnes.label(p.Personne)} - ${Camps.label(p.Camp)}`;
}

async function createAide(idParticipant: IdParticipant, file: File) {
  showCreateAide.value = false;
  const res = await controller.UploadAideJustificatif(file, {
    token: props.token,
    idParticipant,
  });
  if (res === undefined) return;
  controller.showMessage(
    "Justificatif envoyé avec succès (validation à venir). Merci !",
  );
  emit("refresh");
}

async function uploadQuotientFamilial() {
  if (!hasQuotientFamilialOption.value) return;
  if (qfValidated.value) {
    controller.showMessage(qfValidatedMessage);
    return;
  }
  if (!quotientFamilialFile.value) return;
  const res = await controller.UploadQuotientFamilial(
    quotientFamilialFile.value,
    {
      token: props.token,
    },
  );
  if (res === undefined) return;
  controller.showMessage(
    "Attestation envoyée avec succès (validation à venir). Merci !",
  );
  quotientFamilialFile.value = null;
  showQuotientFamilialUpload.value = false;
  emit("refresh");
}

async function deleteQuotientFamilial() {
  if (qfValidated.value) {
    controller.showMessage(qfValidatedMessage);
    return;
  }
  const res = await controller.DeleteQuotientFamilial({ token: props.token });
  if (res === undefined) return;
  controller.showMessage("Attestation supprimée avec succès.");
  emit("refresh");
}

async function deleteAideJustificatif(idParticipant: IdParticipant) {
  const res = await controller.DeleteAideJustificatif({
    token: props.token,
    idParticipant,
  });
  if (res === undefined) return;
  controller.showMessage("Justificatif supprimé avec succès.");
  emit("refresh");
}

const showDetails = ref(false);
</script>
