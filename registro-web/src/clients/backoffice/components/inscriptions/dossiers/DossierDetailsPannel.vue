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
      ></PaiementEditCard>
    </v-dialog>

    <template #append>
      <v-btn icon="mdi-pencil" @click="showEditDialog = true"></v-btn>
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
        <v-col class="ml-2">
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
      <v-timeline side="end" class="mt-4" density="compact">
        <EventSwitch
          :event="event"
          v-for="(event, i) in events"
          :key="i"
          @edit-paiement="(p) => (paiementToUpdate = p)"
        ></EventSwitch>
      </v-timeline>
    </v-card-text>
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
  type IdAide,
  type IdParticipant,
  type Int,
  type Paiement,
  type Participant,
  type ParticipantsCreateIn,
  type Structureaides,
} from "../../../logic/api";
import {
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

const props = defineProps<{
  dossier: DossierDetails;
  structures: NonNullable<Structureaides>;
  camps: CampItem[];
}>();

const emit = defineEmits<{
  (e: "updateDossier", dossier: Dossier): void;
  (e: "deleteDossier"): void;
  (e: "createParticipant", participant: ParticipantsCreateIn): void;
  (e: "updateParticipant", participant: Participant): void;
  (e: "deleteParticipant", id: IdParticipant): void;
  (e: "expandParticipant", participant: Participant): void;
  (e: "createAide", args: AidesCreateIn): void;
  (e: "updateAide", aide: Aide): void;
  (e: "deleteAide", id: IdAide): void;
  (e: "createPaiement"): void;
  (e: "updatePaiement", paiement: Paiement): void;
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
  // last event first
  out.sort(
    (a, b) => pseudoEventTime(b).valueOf() - pseudoEventTime(a).valueOf()
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
</script>
