<template>
  <v-card title="Nouveau message" :subtitle="hint">
    <v-card-text>
      <v-row>
        <v-col>
          <v-select
            density="comfortable"
            variant="outlined"
            :items="items"
            label="Destinataire"
            :readonly="props.toFondSoutien"
            v-model="selectedDestinataire"
          >
          </v-select>
        </v-col>
      </v-row>
      <v-textarea
        autofocus
        placeholder="Rédigez votre message..."
        v-model="message"
        rows="10"
      ></v-textarea>
    </v-card-text>
    <v-card-actions>
      <v-spacer></v-spacer>
      <v-btn
        :disabled="!message.length || selectedDestinataire == null"
        @click="emit('send', message, selectedDestinataire!)"
        prepend-icon="mdi-send"
      >
        Envoyer</v-btn
      >
    </v-card-actions>
  </v-card>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import { type CampItem, type IdCamp } from "../logic/api.ts";
import { Camps } from "@/utils.ts";

const props = defineProps<{
  camps: CampItem[];
  allowFondSoutien: boolean;
  toFondSoutien: boolean; // if true, destinataire is read-only
}>();

type destinataire = "fond-soutien" | IdCamp | "all";

const emit = defineEmits<{
  (e: "send", contenu: string, destinataire: destinataire): void;
}>();

const selectedDestinataire = ref<destinataire | null>(
  props.toFondSoutien ? "fond-soutien" : "all",
);
const message = ref("");

const hint = computed(() => {
  if (selectedDestinataire.value == "fond-soutien")
    return "Ce message ne sera visible que par le fonds de soutien.";
  if (selectedDestinataire.value == null) return "";
  if (selectedDestinataire.value == "all")
    return "Ce message sera visible par le centre et les directeurs.";
  return "Ce message sera visible par le centre et la direction du séjour.";
});

const items = computed(() => {
  const fs = {
    value: "fond-soutien" as destinataire,
    title: "Fonds de soutien",
  };
  if (props.toFondSoutien) return [fs]; // readonly anyway
  const out = [
    {
      value: "all" as destinataire,
      title: "Tous les séjours",
    },
    { type: "divider" } as const,
  ].concat(...props.camps.map((c) => ({ value: c.Id, title: Camps.label(c) })));
  if (props.allowFondSoutien) {
    out.push({ type: "divider" }, fs);
  }
  return out;
});
</script>
