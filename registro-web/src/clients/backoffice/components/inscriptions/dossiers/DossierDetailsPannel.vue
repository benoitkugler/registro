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
    <!-- actions -->
    <template #append>
      <v-btn
        icon="mdi-pencil"
        @click="showEditDialog = true"
        class="mx-1"
      ></v-btn>

      <v-tooltip text="Ajouter un paiement" location="top">
        <template #activator="{ props: tooltipProps }">
          <v-btn
            icon
            size="small"
            class="mx-1"
            v-bind="tooltipProps"
            @click="emit('createPaiement')"
          >
            <v-icon color="green">mdi-cash-plus</v-icon>
          </v-btn>
        </template>
      </v-tooltip>

      <v-tooltip text="Envoyer un message" location="top">
        <template #activator="{ props: tooltipProps }">
          <v-btn
            icon
            size="small"
            class="mx-1"
            v-bind="tooltipProps"
            @click="showMessage = true"
          >
            <v-icon color="green">mdi-email-plus</v-icon>
          </v-btn>
        </template>
      </v-tooltip>

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
        <v-list>
          <v-list-item
            prepend-icon="mdi-link"
            @click="showLinks = true"
            title="Afficher les identifiants"
          ></v-list-item>
          <v-divider></v-divider>
          <v-list-item
            prepend-icon="mdi-file-move"
            @click="
              mergeParams = {
                From: props.dossier.Dossier.Dossier.Id,
                To: 0 as Int,
                Notifie: true,
              }
            "
            title="Fusionner vers ..."
          ></v-list-item>
          <v-divider></v-divider>
          <v-list-item
            prepend-icon="mdi-delete"
            @click="showDeleteDialog = true"
            title="Supprimer"
          ></v-list-item>
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
      <div class="overflow-y-auto" style="max-height: 70vh">
        <v-timeline side="end" class="mt-4" density="compact">
          <EventSwitch
            :event="event"
            v-for="(event, i) in events"
            :key="i"
            @edit-paiement="(p) => (paiementToUpdate = p)"
            @delete-message="(m) => emit('deleteMessage', m)"
          ></EventSwitch>
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
                :model-value="zeroableToNullable(participantToCreate!.IdCamp)"
                @update:model-value="
                  (v) => (participantToCreate!.IdCamp = nullableToZeroable(v))
                "
              ></SelectCamp>
            </v-col>
            <v-col cols="12">
              <SelectPersonne
                label="Personne"
                initial-personne=""
                :model-value="zeroableToNullable(participantToCreate!.IdPersonne)"
                @update:model-value="
                  (v) => (participantToCreate!.IdPersonne = nullableToZeroable(v))
                "
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
      <v-card title="Modifier le dossier">
        <template v-slot:append>
          <v-btn
            @click="
              participantToCreate = {
                IdCamp: 0 as Int,
                IdPersonne: 0 as Int,
                IdDossier: props.dossier.Dossier.Dossier.Id,
              }
            "
          >
            <template v-slot:prepend>
              <v-icon color="green">mdi-plus</v-icon>
            </template>
            Ajouter un participant</v-btn
          >
        </template>
        <v-card-text>
          <v-row>
            <v-col cols="3">
              <DossierEditCard
                :responsable="props.dossier.Dossier.Responsable"
                :dossier="props.dossier.Dossier.Dossier"
                @save="(v) => emit('updateDossier', v)"
              ></DossierEditCard>
            </v-col>
            <v-divider thickness="4" vertical></v-divider>
            <v-col>
              <div class="my-1"></div>
              <DossierParticipantRow
                v-for="participant in props.dossier.Dossier.Participants"
                :participant="participant"
                :aides="
                  (props.dossier.Dossier.Aides || {})[
                    participant.Participant.Id
                  ]
                "
                :structures="props.structures"
                :has-many-participants="
                  (props.dossier.Dossier.Participants?.length || 0) > 1
                "
                @create-aide="(v) => emit('createAide', v)"
                @delete-aide="(v) => emit('deleteAide', v)"
                @update-aide="(v) => emit('updateAide', v)"
                @update="(p) => emit('updateParticipant', p)"
                @delete="emit('deleteParticipant', participant.Participant.Id)"
                @expand="emit('expandParticipant', participant.Participant)"
              ></DossierParticipantRow>
            </v-col>
          </v-row>
        </v-card-text>
      </v-card>
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
                icon="mdi-content-copy"
                @click="copyEspacepersoURL"
              ></v-btn>
            </v-col>
          </v-row>
          <v-row>
            <v-col cols="3">Identification de virement</v-col>
            <v-col>
              {{ props.dossier.VirementCode }}
            </v-col>
            <v-col align-self="center" cols="auto">
              <v-btn icon="mdi-content-copy" @click="copyVirementCode"></v-btn>
            </v-col>
          </v-row>
        </v-card-text>
      </v-card>
    </v-dialog>

    <!-- merge dialog -->
    <v-dialog
      v-if="mergeParams != null"
      :model-value="mergeParams != null"
      @update:model-value="mergeParams = null"
      max-width="1000px"
    >
      <v-card
        title="Fusionner le dossier vers"
        subtitle="Les participants, paiements et messages seront copiés vers le dossier cible."
      >
        <v-card-text>
          <v-row>
            <v-col cols="8">
              <DossierList
                v-model:query="mergeQuery"
                :camps="props.camps"
                @click="(v) => (mergeParams!.To = v.Id)"
              ></DossierList>
            </v-col>
            <v-col cols="4">
              <v-checkbox
                density="comfortable"
                label="Notifier par email"
                hint="Le reponsable du dossier absorbé sera averti du changement d'espace personnel."
                persistent-hint
                v-model="mergeParams.Notifie"
              ></v-checkbox>
            </v-col>
          </v-row>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn
            :disabled="
              mergeParams.To == 0 || mergeParams.To == mergeParams.From
            "
            @click="
              emit('mergeDossier', mergeParams);
              mergeParams = null;
            "
            >Fusionner</v-btn
          >
        </v-card-actions>
      </v-card>
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
  type Int,
  type Paiement,
  type Participant,
  type ParticipantsCreateIn,
  type Structureaides,
} from "../../../logic/api";
import {
  copyToClipboard,
  nullableToZeroable,
  Personnes,
  pseudoEventTime,
  zeroableToNullable,
  type PseudoEvent,
} from "@/utils";
import FactureCard from "./FactureCard.vue";
import DossierEditCard from "./editor/DossierEditCard.vue";
import DossierParticipantRow from "./editor/DossierParticipantRow.vue";
import PaiementEditCard from "./PaiementEditCard.vue";
import { controller, emptyQuery } from "@/clients/backoffice/logic/logic";
import DossierList from "./DossierList.vue";

const props = defineProps<{
  dossier: DossierDetails;
  structures: NonNullable<Structureaides>;
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
  // paiements
  (e: "createPaiement"): void;
  (e: "updatePaiement", paiement: Paiement): void;
  (e: "deletePaiement", paiement: Paiement): void;
  // events
  (e: "sendMessage", contenu: string): void;
  (e: "deleteMessage", event: Event): void;
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
  // last event last
  out.sort(
    (a, b) => pseudoEventTime(a).valueOf() - pseudoEventTime(b).valueOf()
  );
  return out;
});

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

const showLinks = ref(false);
async function copyEspacepersoURL() {
  await copyToClipboard(props.dossier.EspacepersoURL);
  controller.showMessage("Lien vers l'espace personnel copié.");
}
async function copyVirementCode() {
  await copyToClipboard(props.dossier.VirementCode);
  controller.showMessage("Identifiant de virement copié.");
}

const mergeQuery = ref(emptyQuery());
const mergeParams = ref<DossiersMergeIn | null>(null);

const showMessage = ref(false);
const messageContenu = ref("");
</script>
