<template>
  <v-card title="Modifier les participants">
    <v-dialog
      v-if="aideToCreate != null"
      :model-value="aideToCreate != null"
      @update:model-value="aideToCreate = null"
      max-width="400px"
    >
      <v-card title="Ajouter une aide extérieure">
        <v-card-text>
          <v-select
            variant="outlined"
            density="comfortable"
            label="Structure"
            :items="structureItems"
            :model-value="
              aideToCreate.IdStructure == 0 ? null : aideToCreate.IdStructure
            "
            @update:model-value="(v) => (aideToCreate.IdStructure = v)"
          ></v-select>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn
            color="green"
            @click="
              emit('createAide', aideToCreate);
              aideToCreate = null;
            "
            :disabled="aideToCreate.IdStructure == 0"
            >Créer</v-btn
          >
        </v-card-actions>
      </v-card>
    </v-dialog>

    <v-card-text>
      <ParticipantRow
        v-for="participant in props.dossier.Participants"
        :participant="participant"
        :aides="(props.dossier.Aides || {})[participant.Participant.Id]"
        :structures="props.structures"
        :has-many-participants="(props.dossier.Participants?.length || 0) > 1"
        @create-aide="
          aideToCreate = {
            IdParticipant: participant.Participant.Id,
            IdStructure: 0 as Int,
          }
        "
      ></ParticipantRow>
    </v-card-text>
    <!-- <v-card-actions>
        <v-spacer></v-spacer>
        <v-btn @click="emit('save', innerDossier)">Enregistrer</v-btn>
      </v-card-actions> -->
  </v-card>
</template>

<script setup lang="ts">
import {
  type AidesCreateIn,
  type DossierExt,
  type Int,
  type Structureaides,
} from "@/clients/backoffice/logic/api";
import { computed, ref } from "vue";
import ParticipantRow from "./ParticipantRow.vue";

const props = defineProps<{
  dossier: DossierExt;
  structures: NonNullable<Structureaides>;
}>();

const emit = defineEmits<{
  (e: "createAide", args: AidesCreateIn): void;
}>();

const aideToCreate = ref<AidesCreateIn | null>(null);

const structureItems = computed(() =>
  Object.values(props.structures).map((s) => ({ value: s.Id, title: s.Nom }))
);
</script>
