<template>
  <v-card title="Nouveau message">
    <v-card-text>
      <v-row>
        <v-col>
          <v-select
            label="Destinataire"
            density="comfortable"
            variant="plain"
            hide-details
            :items="
              recordEntries(props.dossiers)
                .map((p) => ({
                  value: p[0],
                  title: p[1].Responsable,
                  subtitle: (p[1].Participants || []).join(', '),
                }))
                .sort((a, b) => a.title.localeCompare(b.title))
            "
            v-model="destinataire"
            :readonly="props.initialDestinataire !== null"
          >
            <template #item="{ item }">
              <v-list-item
                :title="item.title"
                :subtitle="item.raw.subtitle"
              ></v-list-item>
            </template>
          </v-select>
        </v-col>
      </v-row>
      <v-row>
        <v-col>
          <v-textarea
            autofocus
            placeholder="Rédigez votre message..."
            v-model="newMessage"
            rows="10"
          ></v-textarea>
        </v-col>
      </v-row>
    </v-card-text>
    <v-card-actions>
      <v-spacer></v-spacer>
      <v-btn
        :disabled="destinataire == null || !newMessage.length"
        @click="emit('send', destinataire!, newMessage)"
        prepend-icon="mdi-send"
      >
        Envoyer</v-btn
      >
    </v-card-actions>
  </v-card>
</template>

<script lang="ts" setup>
import { ref } from "vue";
import type { IdDossier, Messages } from "../../logic/api";
import { recordEntries } from "@/utils.ts";

const props = defineProps<{
  dossiers: Messages["Dossiers"];
  initialDestinataire: IdDossier | null;
}>();

const emit = defineEmits<{
  (e: "send", destinataire: IdDossier, message: string): void;
}>();

const destinataire = ref<IdDossier | null>(props.initialDestinataire);
const newMessage = ref("");
</script>
