<template>
  <v-card title="Modifier les options">
    <v-card-text>
      <v-row>
        <v-col v-for="(participant, index) in props.participants" cols="6">
          <v-card
            :title="Personnes.label(participant.Personne)"
            :subtitle="Camps.label(participant.Camp)"
          >
            <v-card-text>
              <v-row>
                <v-col> TODO: option prix </v-col>
              </v-row>
            </v-card-text>
          </v-card>
        </v-col>
      </v-row>
    </v-card-text>
    <v-card-actions>
      <v-spacer></v-spacer>
      <v-btn
        @click="emit('save', inner)"
        color="green"
        prepend-icon="mdi-content-save-cog"
      >
        Enregistrer</v-btn
      >
    </v-card-actions>
  </v-card>
</template>

<script setup lang="ts">
import { ref } from "vue";
import { type Participant } from "@/clients/backoffice/logic/api";
import type { PublicParticipantCamp } from "../logic/api";
import { Camps, Personnes } from "@/utils";
const props = defineProps<{
  participants: PublicParticipantCamp[];
}>();
const emit = defineEmits<{
  (e: "save", participants: Participant[]): void;
}>();

const inner = ref(props.participants.map((p) => p.Participant));
</script>
