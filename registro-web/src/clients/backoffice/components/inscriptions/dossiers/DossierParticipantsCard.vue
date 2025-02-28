<template>
  <v-card title="Modifier les participants">
    <v-card-text>
      <ParticipantRow
        v-for="participant in props.dossier.Participants"
        :participant="participant"
        :aides="(props.dossier.Aides || {})[participant.Participant.Id]"
        :has-many-participants="(props.dossier.Participants?.length || 0) > 1"
      ></ParticipantRow>
    </v-card-text>
    <!-- <v-card-actions>
        <v-spacer></v-spacer>
        <v-btn @click="emit('save', innerDossier)">Enregistrer</v-btn>
      </v-card-actions> -->
  </v-card>
</template>

<script setup lang="ts">
import { type Dossier, type DossierExt } from "@/clients/backoffice/logic/api";
import { copy } from "@/utils";
import { ref } from "vue";
import ParticipantRow from "./ParticipantRow.vue";

const props = defineProps<{
  dossier: DossierExt;
}>();

const emit = defineEmits<{
  (e: "save", dossier: Dossier): void;
}>();

const innerDossier = ref(copy(props.dossier));
</script>
