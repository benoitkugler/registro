<template>
  <v-card title="Ajouter une aide extérieure">
    <v-card-text>
      <v-row>
        <v-col>
          <v-select
            variant="outlined"
            density="comfortable"
            label="Participant"
            v-model="aide.IdParticipant"
            :items="participantsItems"
          >
            <template #item="{ item, props: itemProps }">
              <v-list-item
                v-bind="itemProps"
                :title="item.raw.title"
                :subtitle="item.raw.subtitle"
              ></v-list-item>
            </template>
          </v-select>
        </v-col>
        <v-col>
          <v-select
            variant="outlined"
            density="comfortable"
            label="Structure"
            :model-value="zeroableToNullable(aide.IdStructureaide)"
            @update:model-value="
              (v) => (aide.IdStructureaide = nullableToZeroable(v))
            "
            :items="structures"
            no-data-text="Aucune structure n'est disponible."
          ></v-select>
        </v-col>
      </v-row>
      <v-row>
        <v-col align-self="center" cols="5">
          <MontantField
            label="Valeur"
            v-model="aide.Valeur"
            hide-details
          ></MontantField>
        </v-col>
        <v-col align-self="center">
          <v-checkbox
            v-model="aide.ParJour"
            label="Montant par jour"
            hide-details
          ></v-checkbox>
        </v-col>
        <v-col align-self="center">
          <IntField
            v-if="aide.ParJour"
            label="Limite sur le nombre de jours"
            v-model="aide.NbJoursMax"
            :min="0 as Int"
            hint="0 pour ne pas limiter"
            persistent-hint
          ></IntField>
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
        :disabled="!isAideValide"
        @click="emit('save', aide, file!)"
      >
        Ajouter
      </v-btn>
    </v-card-actions>
  </v-card>
</template>

<script setup lang="ts">
import {
  Currency,
  type Aide,
  type DossierExt,
  type IdAide,
  type IdStructureaide,
  type Int,
  type Structureaides,
} from "../logic/api";
import { computed, ref } from "vue";
import { controller } from "../logic/logic";
import {
  Camps,
  nullableToZeroable,
  Personnes,
  zeroableToNullable,
} from "@/utils";
const props = defineProps<{
  dossier: DossierExt;
  structureaides: Structureaides;
}>();

const emit = defineEmits<{
  (e: "save", aide: Aide, file: File): void;
}>();

const aide = ref<Aide>({
  Id: 0 as IdAide,
  IdStructureaide: 0 as IdStructureaide,
  IdParticipant: (props.dossier.Participants || [])[0].Participant.Id,
  Valide: false,
  Valeur: { Currency: Currency.Euros, Cent: 0 as Int },
  ParJour: false,
  NbJoursMax: 0 as Int,
});
const file = ref<File | null>(null);
const isAideValide = computed(() => {
  if (!aide.value) return false;
  return (
    file.value != null &&
    aide.value.IdParticipant != 0 &&
    aide.value.IdStructureaide != 0 &&
    aide.value.Valeur.Cent != 0
  );
});

const participantsItems = computed(() =>
  (props.dossier.Participants || []).map((p) => ({
    title: Personnes.label(p.Personne),
    subtitle: Camps.label(p.Camp),
    value: p.Participant.Id,
  }))
);

const structures = computed(() =>
  Object.values(props.structureaides || {}).map((s) => ({
    value: s.Id,
    title: s.Nom,
  }))
);
</script>
