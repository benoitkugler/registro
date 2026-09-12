<template>
  <v-card
    title="Nouveau message"
    :subtitle="
      showCreateMessage.toFondSoutien
        ? 'Ce message ne sera visible que par le fonds de soutien.'
        : 'Ce message sera visible par le centre et les directeurs.'
    "
  >
    <v-card-text>
      <v-textarea
        autofocus
        placeholder="Rédigez votre message..."
        v-model="showCreateMessage.content"
        rows="10"
      ></v-textarea>
    </v-card-text>
    <v-card-actions>
      <v-spacer></v-spacer>
      <v-btn
        :disabled="!showCreateMessage.content.length"
        @click="sendMessage"
        prepend-icon="mdi-send"
      >
        Envoyer</v-btn
      >
    </v-card-actions>
  </v-card>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import {
  type ChampReponse,
  type ChampReponseQCM,
  type ChampReponseTexte,
  type FormReponses,
  type IdCamp,
} from "../logic/api.ts";
import { copy } from "@/utils";

const props = defineProps<{
  toFondSoutien: boolean; // if true, destinataire is read-only
  camps:
}>();

const emit = defineEmits<{
  (
    e: "send",
    contenu: string,
    destinataire: "fond-soutien" | IdCamp | null,
  ): void;
}>();
</script>
