<template>
  <v-menu location="bottom center">
    <template #activator="{ props: menuProps }">
      <v-chip v-bind="menuProps" :color="color" class="my-1" size="small">
        {{ DemandeStateLabels[props.documents.State] }}
        {{ noteCasierJudiciaire ? "*" : "" }}
      </v-chip>
    </template>
    <v-card
      title="Demande de document"
      :subtitle="
        noteCasierJudiciaire
          ? `L'équipier étant Suisse, ce document n'est pas requis.`
          : ''
      "
    >
      <v-card-text>
        <v-btn-toggle
          density="comfortable"
          rounded="lg"
          variant="outlined"
          color="primary"
          mandatory
          :model-value="props.documents.State"
          @update:model-value="(v) => emit('updateState', v)"
        >
          <v-btn :value="DemandeState.NonDemande">Non requis</v-btn>
          <v-btn :value="DemandeState.Optionnelle">{{
            DemandeStateLabels[DemandeState.Optionnelle]
          }}</v-btn>
          <v-btn :value="DemandeState.Obligatoire">{{
            DemandeStateLabels[DemandeState.Obligatoire]
          }}</v-btn>
        </v-btn-toggle>
        <template v-if="props.documents.Files?.length">
          <v-divider thickness="2" class="mt-2"></v-divider>
          <v-row>
            <v-col v-for="file in props.documents.Files" cols="4">
              <FileCard :file="file"></FileCard>
            </v-col>
          </v-row>
        </template>
      </v-card-text>
    </v-card>
  </v-menu>
</template>

<script lang="ts" setup>
import { computed } from "vue";
import {
  Categorie,
  DemandeState,
  DemandeStateLabels,
  type Demande,
  type EquipierDemande,
} from "../../logic/api";

const props = defineProps<{
  isSuisse: boolean;
  demande: Demande;
  documents: EquipierDemande;
}>();

const emit = defineEmits<{
  (e: "updateState", s: DemandeState): void;
}>();

const color = computed(() => {
  const hasFile = !!props.documents.Files?.length;
  switch (props.documents.State) {
    case DemandeState.NonDemande:
      return undefined;
    case DemandeState.Optionnelle:
      return hasFile ? "green" : "yellow-darken-2";
    case DemandeState.Obligatoire:
      return hasFile ? "green" : "red";
  }
});

const noteCasierJudiciaire = computed(
  () =>
    props.demande.Categorie == Categorie.ExtraitCasierJudiciaire &&
    props.isSuisse
);
</script>
