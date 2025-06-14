<template>
  <v-card
    :title="`Dossier de ${props.dossier.Dossier.Responsable}`"
    class="ml-2"
  >
    <template #subtitle>
      <span class="mr-1" v-for="p in props.dossier.Dossier.Participants">
        <a href="/camps" @click.prevent="goToParticipant(p.Participant)">
          {{ Personnes.label(p.Personne) }}
        </a>
      </span>
    </template>

    <!-- actions -->
    <template #append>
      <v-btn
        icon="mdi-pencil"
        @click="showEditDialog = true"
        class="mx-1"
      ></v-btn>

      <v-menu>
        <template #activator="{ props: menuProps }">
          <v-btn
            v-bind="menuProps"
            icon="mdi-dots-vertical"
            size="small"
            class="mx-1"
          >
          </v-btn>
        </template>
        <!-- actions -->
        <v-list density="compact">
          <v-list-item
            prepend-icon="mdi-email"
            @click="showMessage = true"
            title="Envoyer un message"
            subtitle="Notification par email"
          >
          </v-list-item>
          <v-list-item
            prepend-icon="mdi-invoice-send"
            @click="emit('sendFacture')"
            title="Envoyer une demande de règlement"
            subtitle="Notification par email"
          >
          </v-list-item>
          <v-divider></v-divider>
          <v-list-item
            prepend-icon="mdi-cash-plus"
            @click="emit('createPaiement')"
            title="Ajouter un paiement"
          >
          </v-list-item>
          <v-divider></v-divider>
          <v-list-item
            prepend-icon="mdi-link"
            @click="showLinks = true"
            title="Afficher les identifiants"
          ></v-list-item>
          <v-divider></v-divider>
          <v-list-item
            prepend-icon="mdi-file-move"
            @click="showMergeCard = true"
            title="Fusionner vers ..."
          ></v-list-item>
          <v-divider></v-divider>
          <v-list-item @click="showDeleteDialog = true" title="Supprimer">
            <template #prepend>
              <v-icon color="red">mdi-delete</v-icon>
            </template>
          </v-list-item>
        </v-list>
      </v-menu>
    </template>

    <v-card-text>
      <!-- récap financier -->
      <v-row>
        <v-col class="ml-2 mb-1 text-center">
          <v-menu>
            <template #activator="{ props: menuProps }">
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
      <div class="overflow-y-auto" style="max-height: 65vh">
        <v-timeline side="end" class="mt-4" density="compact">
          <EventSwitch
            :event="event"
            v-for="(event, i) in events"
            :key="i"
            @edit-paiement="(p) => (paiementToUpdate = p)"
            @delete-message="(m) => emit('deleteMessage', m)"
          >
          </EventSwitch>
        </v-timeline>
      </div>
    </v-card-text>

    <!-- confirme delete dialog -->
    <v-dialog v-model="showDeleteDialog" max-width="600px">
      <v-card title="Supprimer le dossier">
        <v-card-text>
          Confirmez-vous la suppression du dossier et de ses participants ?
          <br /><br />

          Attention, cette opération est irréversible.
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn
            color="red"
            @click="
              emit('deleteDossier');
              showDeleteDialog = false;
            "
            >Supprimer</v-btn
          >
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- participant create dialog -->
    <v-dialog
      v-if="participantToCreate != null"
      :model-value="participantToCreate != null"
      @update:model-value="participantToCreate = null"
      max-width="600px"
    >
      <v-card title="Ajouter un participant">
        <v-card-text>
          <v-row>
            <v-col cols="12">
              <SelectCamp
                label="Camp"
                :camps="props.camps"
                v-model="participantToCreate.IdCamp"
              ></SelectCamp>
            </v-col>
            <v-col cols="12">
              <SelectPersonne
                label="Personne"
                initial-personne=""
                v-model="participantToCreate.IdPersonne"
                :api="{
                  SelectPersonne: controller.SelectPersonne.bind(controller),
                }"
              ></SelectPersonne>
            </v-col>
          </v-row>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn
            color="green"
            :disabled="
              participantToCreate.IdPersonne == 0 ||
              participantToCreate.IdCamp == 0
            "
            @click="
              emit('createParticipant', participantToCreate);
              participantToCreate = null;
            "
            >Ajouter</v-btn
          >
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- dossier editor -->
    <v-dialog v-model="showEditDialog">
      <DossierEditor
        :dossier="props.dossier"
        :camps="props.camps"
        @updateDossier="(v) => emit('updateDossier', v)"
        @createParticipant="(v) => emit('createParticipant', v)"
        @updateParticipant="(v) => emit('updateParticipant', v)"
        @deleteParticipant="(v) => emit('deleteParticipant', v)"
        @expandParticipant="(v) => emit('expandParticipant', v)"
        @createAide="(v) => emit('createAide', v)"
        @updateAide="(v) => emit('updateAide', v)"
        @deleteAide="(v) => emit('deleteAide', v)"
        @deleteFileAide="(v) => emit('deleteFileAide', v)"
        @uploadFileAide="(f, v) => emit('uploadFileAide', f, v)"
      >
      </DossierEditor>
    </v-dialog>

    <!-- paiement editor -->
    <v-dialog
      v-if="paiementToUpdate != null"
      :model-value="paiementToUpdate != null"
      @update:model-value="paiementToUpdate = null"
      max-width="600px"
    >
      <PaiementEditCard
        :paiement="paiementToUpdate"
        @update="
          (p) => {
            emit('updatePaiement', p);
            paiementToUpdate = null;
          }
        "
        @delete="
          emit('deletePaiement', paiementToUpdate);
          paiementToUpdate = null;
        "
      ></PaiementEditCard>
    </v-dialog>

    <!-- identifiants dialog -->
    <v-dialog v-model="showLinks" max-width="800px">
      <v-card title="Liens et identifiants">
        <v-card-text>
          <v-row>
            <v-col cols="3">Espace personnel</v-col>
            <v-col class="text-truncate">
              <a :href="props.dossier.EspacepersoURL" target="_blank"
                >{{ props.dossier.EspacepersoURL }}
              </a>
            </v-col>
            <v-col align-self="center" cols="auto">
              <v-btn
                size="small"
                icon="mdi-content-copy"
                @click="copyEspacepersoURL"
              ></v-btn>
            </v-col>
          </v-row>
          <v-row v-for="account in props.dossier.BankAccounts">
            <v-col cols="3">RIB ({{ asso }})</v-col>
            <v-col cols="3">{{ account[0] }}</v-col>
            <v-col>{{ account[1] }}</v-col>
            <v-col align-self="center" cols="auto">
              <v-btn
                size="small"
                icon="mdi-content-copy"
                @click="() => copyBankIBAN(account[1])"
              ></v-btn>
            </v-col>
          </v-row>
          <v-row>
            <v-col cols="3">Identification de virement</v-col>
            <v-col>
              {{ props.dossier.VirementCode }}
            </v-col>
            <v-col align-self="center" cols="auto">
              <v-btn
                size="small"
                icon="mdi-content-copy"
                @click="copyVirementCode"
              ></v-btn>
            </v-col>
          </v-row>
        </v-card-text>
      </v-card>
    </v-dialog>

    <!-- merge dialog -->
    <v-dialog v-model="showMergeCard" max-width="600px">
      <MergeCard
        :from="props.dossier.Dossier.Dossier.Id"
        @merge="
          (args) => {
            emit('mergeDossier', args);
            showMergeCard = false;
          }
        "
      ></MergeCard>
    </v-dialog>

    <!-- message dialog -->
    <v-dialog v-model="showMessage" max-width="600px">
      <v-card
        title="Nouveau message"
        subtitle="Envoie un message sur le fil de suivi et une notification par mail."
      >
        <v-card-text>
          <v-row>
            <v-col>
              <v-textarea
                rows="5"
                variant="outlined"
                v-model="messageContenu"
              ></v-textarea>
            </v-col>
          </v-row>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn
            :disabled="!messageContenu.length"
            @click="
              emit('sendMessage', messageContenu);
              showMessage = false;
            "
            >Envoyer</v-btn
          >
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-card>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import {
  StatutPaiement,
  type Aide,
  type AidesCreateIn,
  type CampItem,
  type Dossier,
  type DossierDetails,
  type DossiersMergeIn,
  type Event,
  type IdAide,
  type IdParticipant,
  type Paiement,
  type Participant,
  type ParticipantsCreateIn,
} from "../../../logic/api";
import { buildPseudoEvents, copyToClipboard, Personnes } from "@/utils";
import FactureCard from "./FactureCard.vue";
import PaiementEditCard from "./PaiementEditCard.vue";
import { controller } from "@/clients/backoffice/logic/logic";
import { goToParticipant } from "@/clients/backoffice/plugins/router";
import MergeCard from "../MergeCard.vue";
import DossierEditor from "./editor/DossierEditor.vue";

const props = defineProps<{
  dossier: DossierDetails;
  camps: CampItem[];
}>();

const emit = defineEmits<{
  (e: "updateDossier", dossier: Dossier): void;
  (e: "deleteDossier"): void;
  (e: "mergeDossier", args: DossiersMergeIn): void;
  // participants
  (e: "createParticipant", participant: ParticipantsCreateIn): void;
  (e: "updateParticipant", participant: Participant): void;
  (e: "deleteParticipant", id: IdParticipant): void;
  (e: "expandParticipant", participant: Participant): void;
  // aides
  (e: "createAide", args: AidesCreateIn): void;
  (e: "updateAide", aide: Aide): void;
  (e: "deleteAide", id: IdAide): void;
  (e: "deleteFileAide", id: IdAide): void;
  (e: "uploadFileAide", f: File, id: IdAide): void;
  // paiements
  (e: "createPaiement"): void;
  (e: "updatePaiement", paiement: Paiement): void;
  (e: "deletePaiement", paiement: Paiement): void;
  // events
  (e: "sendMessage", contenu: string): void;
  (e: "deleteMessage", event: Event): void;
  (e: "sendFacture"): void;
}>();

defineExpose({ showEditPaiement, showEditDossier });

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

const events = computed(() =>
  buildPseudoEvents(props.dossier.Dossier, "backoffice")
);

const showDeleteDialog = ref(false);

const showEditDialog = ref(false);
function showEditDossier() {
  showEditDialog.value = true;
}
const participantToCreate = ref<ParticipantsCreateIn | null>(null);

const paiementToUpdate = ref<Paiement | null>(null);
function showEditPaiement(paiement: Paiement) {
  paiementToUpdate.value = paiement;
}

const asso = import.meta.env.VITE_ASSO_TITLE;
const showLinks = ref(false);
async function copyEspacepersoURL() {
  await copyToClipboard(props.dossier.EspacepersoURL);
  controller.showMessage("Lien vers l'espace personnel copié.");
}
async function copyVirementCode() {
  await copyToClipboard(props.dossier.VirementCode);
  controller.showMessage("Identifiant de virement copié.");
}
async function copyBankIBAN(iban: string) {
  await copyToClipboard(iban);
  controller.showMessage("IBAN copié.");
}

const showMergeCard = ref(false);

const showMessage = ref(false);
const messageContenu = ref("");
</script>
