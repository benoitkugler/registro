<template>
  <v-card title="Modifier le dossier">
    <template v-slot:append>
      <v-btn
        @click="
          participantToCreate = {
            IdCamp: 0 as IdCamp,
            IdPersonne: 0 as IdPersonne,
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
          >
          </DossierEditCard>
        </v-col>
        <v-divider thickness="4" vertical></v-divider>
        <v-col v-if="!isLoading">
          <div class="my-1"></div>
          <DossierParticipantRow
            v-for="participant in props.dossier.Dossier.Participants"
            :participant="participant"
            :aides="
              (props.dossier.Dossier.Aides || {})[participant.Participant.Id]
            "
            :aides-files="props.dossier.Dossier.AidesFiles || {}"
            :has-many-participants="
              (props.dossier.Dossier.Participants?.length || 0) > 1
            "
            v-model:structures="structures"
            @create-aide="(v) => emit('createAide', v)"
            @delete-aide="(v) => emit('deleteAide', v)"
            @update-aide="(v) => emit('updateAide', v)"
            @deleteFileAide="(v) => emit('deleteFileAide', v)"
            @uploadFileAide="(f, v) => emit('uploadFileAide', f, v)"
            @update="(p) => emit('updateParticipant', p)"
            @delete="emit('deleteParticipant', participant.Participant.Id)"
            @expand="emit('expandParticipant', participant.Participant)"
          ></DossierParticipantRow>
        </v-col>
      </v-row>
    </v-card-text>
  </v-card>
</template>

<script setup lang="ts">
import { onMounted, ref } from "vue";
import {
  type Aide,
  type AidesCreateIn,
  type CampItem,
  type Dossier,
  type DossierDetails,
  type IdAide,
  type IdCamp,
  type IdParticipant,
  type IdPersonne,
  type Participant,
  type ParticipantsCreateIn,
  type Structureaides,
} from "../../../../logic/api";
import DossierEditCard from "./DossierEditCard.vue";
import DossierParticipantRow from "./DossierParticipantRow.vue";
import { controller } from "@/clients/backoffice/logic/logic";

const props = defineProps<{
  dossier: DossierDetails;
  camps: CampItem[];
}>();

const emit = defineEmits<{
  (e: "updateDossier", dossier: Dossier): void;
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
}>();

const participantToCreate = ref<ParticipantsCreateIn | null>(null);

onMounted(loadStructureaides);

// ensure [structures] is properly loaded
const isLoading = ref(true);
const structures = ref<NonNullable<Structureaides>>({});
async function loadStructureaides() {
  const res = await controller.StructureaidesGet();
  if (res === undefined) return;
  isLoading.value = false;
  structures.value = res || {};
}
</script>
