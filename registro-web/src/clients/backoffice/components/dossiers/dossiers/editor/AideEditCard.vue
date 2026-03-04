<template>
  <v-card
    title="Modifier l'aide extérieure"
    :subtitle="`pour ${Personnes.label(
      props.participant.Personne
    )} - ${Camps.label(props.participant.Camp)}`"
  >
    <template #append>
      <v-btn @click="emit('showStructureaides')">
        <v-icon>mdi-view-list</v-icon>
        Structures</v-btn
      >
    </template>
    <v-card-text>
      <v-row>
        <v-col>
          <v-select
            variant="outlined"
            density="comfortable"
            label="Structure"
            :items="structureItems"
            v-model="inner.IdStructureaide"
          ></v-select>
        </v-col>
        <v-col>
          <BoolField
            v-model="inner.Valide"
            label="Aide validée"
            hide-details
          ></BoolField>
        </v-col>
      </v-row>
      <v-row>
        <v-col align-self="center">
          <MontantField
            v-model="inner.Valeur"
            label="Montant"
            hide-details
          ></MontantField>
        </v-col>
        <v-col align-self="center" cols="auto">
          <BoolField hide-details label="Par jour" v-model="inner.ParJour">
          </BoolField>
        </v-col>
        <v-col align-self="center">
          <IntField
            label="Limite"
            suffix="jours"
            v-model="inner.NbJoursMax"
            :disabled="!inner.ParJour"
            hint="0 pour ne pas limiter"
            :min="0 as Int"
            persistent-hint
          ></IntField>
        </v-col>
      </v-row>
      <FilesRow
        class="mt-4"
        :files="props.file ? [props.file] : []"
        title="Document justificatif"
        :max-docs="1"
        :in-upload="false"
        @delete="emit('deleteFile')"
        @upload="(f) => emit('uploadFile', f)"
      ></FilesRow>
    </v-card-text>
    <v-card-actions>
      <v-spacer></v-spacer>
      <v-btn @click="emit('save', inner)">Enregistrer</v-btn>
    </v-card-actions>
  </v-card>
</template>

<script setup lang="ts">
import {
  type Aide,
  type Int,
  type ParticipantCamp,
  type PublicFile,
  type Structureaides,
} from "@/clients/backoffice/logic/api";
import { Camps, copy, Personnes } from "@/utils";
import { computed, ref } from "vue";

const props = defineProps<{
  aide: Aide;
  file: PublicFile | undefined;
  structures: NonNullable<Structureaides>;
  participant: ParticipantCamp;
}>();

const emit = defineEmits<{
  (e: "save", aide: Aide): void;
  (e: "deleteFile"): void;
  (e: "uploadFile", f: File): void;
  (e: "showStructureaides"): void;
}>();

const inner = ref(copy(props.aide));

const structureItems = computed(() =>
  Object.values(props.structures).map((s) => ({ value: s.Id, title: s.Nom }))
);
</script>
