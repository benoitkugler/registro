<template>
  <v-card
    title="Envoyer un justificatif d'aide"
    subtitle="Ce justificatif sera vérifié par le centre."
  >
    <v-card-text>
      <v-row>
        <v-col>
          <v-select
            variant="outlined"
            density="comfortable"
            label="Participant concerné"
            :model-value="zeroableToNullable(idParticipant)"
            @update:model-value="
              (id) => (idParticipant = nullableToZeroable(id))
            "
            :items="participantsItems"
          >
            <template #item="{ item, props: itemProps }">
              <v-list-item
                v-bind="itemProps"
                :title="item.raw.title"
                :subtitle="item.raw.subtitle"
                :disabled="!item.raw.enabled"
              ></v-list-item>
            </template>
          </v-select>
        </v-col>
      </v-row>
      <v-row>
        <v-col>
          <FileInput
            label="Pièce justificative"
            @update="(f) => (file = f)"
          ></FileInput>
        </v-col>
      </v-row>
    </v-card-text>
    <v-card-actions>
      <v-spacer></v-spacer>
      <v-btn
        color="green"
        :disabled="!isValid"
        @click="emit('save', idParticipant, file!)"
      >
        Envoyer pour vérification
      </v-btn>
    </v-card-actions>
  </v-card>
</template>

<script setup lang="ts">
import {
  StatutParticipant,
  type DossierExt,
  type IdParticipant,
} from "../logic/api";
import { computed, ref } from "vue";
import {
  Camps,
  nullableToZeroable,
  Personnes,
  zeroableToNullable,
} from "@/utils";

const props = defineProps<{
  dossier: DossierExt;
}>();

const emit = defineEmits<{
  (e: "save", idParticipant: IdParticipant, file: File): void;
}>();

const idParticipant = ref<IdParticipant>(0 as IdParticipant);
const file = ref<File | null>(null);

const isValid = computed(() => file.value != null && idParticipant.value != 0);

const participantsItems = computed(() =>
  (props.dossier.Participants || []).map((p) => ({
    title: Personnes.label(p.Personne),
    subtitle: Camps.label(p.Camp),
    value: p.Participant.Id,
    enabled:
      p.Participant.Statut == StatutParticipant.Inscrit &&
      !(props.dossier.Aides || {})[p.Participant.Id],
  })),
);
</script>
