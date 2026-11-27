<template>
  <v-card subtitle="Paiement" class="mt-2">
    <template #append>
      <v-btn
        size="small"
        :disabled="!props.dossier.Participants?.length"
        @click.stop="showCreateAide = true"
        v-if="props.supportsAidesExt"
      >
        <template #prepend>
          <v-icon color="green">mdi-plus</v-icon>
        </template>
        Envoyer un justificatif
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
          </v-list>
        </v-card-text>
      </v-card>
    </v-card-text>

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
  RequiredPaiementState,
  StatutPaiement,
  type DossierExt,
  type IdParticipant,
  type PublicFile,
} from "../logic/api";
import { computed, ref } from "vue";
import AideCard from "./AideCard.vue";
import { controller } from "../logic/logic";
import { Camps, endpoints, Personnes } from "@/utils";
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
