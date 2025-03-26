<template>
  <v-menu>
    <template #activator="{ props: menuProps }">
      <v-chip v-bind="menuProps" :color="color" class="my-1" size="small">
        {{ DemandeStateLabels[props.demande.State] }}
      </v-chip>
    </template>
    <v-card title="Documents">
      <v-card-text>
        <v-btn-toggle
          :model-value="props.demande.State"
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
        <template v-if="props.demande.Files?.length">
          <v-divider thickness="1"></v-divider>
          <v-row>
            <v-col v-for="file in props.demande.Files">
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
  DemandeState,
  DemandeStateLabels,
  type EquipierDemande,
} from "../../logic/api";

const props = defineProps<{
  demande: EquipierDemande;
}>();

const emit = defineEmits<{
  (e: "updateState", s: DemandeState): void;
}>();

const color = computed(() => {
  const hasFile = !!props.demande.Files?.length;
  switch (props.demande.State) {
    case DemandeState.NonDemande:
      return undefined;
    case DemandeState.Optionnelle:
      return hasFile ? "green" : "yellow-darken-2";
    case DemandeState.Obligatoire:
      return hasFile ? "green" : "red";
  }
});
</script>
