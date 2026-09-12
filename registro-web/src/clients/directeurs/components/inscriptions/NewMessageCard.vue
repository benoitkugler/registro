<template>
  <v-card title="Nouveau message">
    <v-card-text>
      <v-row>
        <v-col align-self="center">
          <v-select
            label="Destinataire"
            density="comfortable"
            variant="plain"
            hide-details
            :items="items"
            v-model:model-value="destinataire"
            :readonly="props.initialDestinataire !== null"
          >
            <template #item="{ props: itemProps, item }">
              <v-list-item
                v-bind="itemProps"
                :title="item.title"
                :subtitle="item.raw.subtitle"
                :prepend-icon="
                  item.raw.value == 'all' ? 'mdi-account-group' : undefined
                "
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
import { computed, ref } from "vue";
import type { IdDossier, Messages } from "../../logic/api";
import { recordEntries } from "@/utils.ts";

const props = defineProps<{
  dossiers: Messages["Dossiers"];
  initialDestinataire: IdDossier | null;
}>();

const emit = defineEmits<{
  (e: "send", destinataire: IdDossier | "all", message: string): void;
}>();

const destinataire = ref<IdDossier | "all" | null>(props.initialDestinataire);
const newMessage = ref("");

const items = computed(() => {
  const l = recordEntries(props.dossiers)
    .map((p) => ({
      value: p[0],
      title: p[1].Responsable,
      subtitle: (p[1].Participants || []).join(", "),
    }))
    .sort((a, b) => a.title.localeCompare(b.title));
  const all: { value: IdDossier | "all"; title: string; subtitle?: string }[] =
    [{ value: "all", title: "Tous les inscrits" }];
  return all.concat(...l);
});
</script>
