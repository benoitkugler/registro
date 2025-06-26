<template>
  <v-card title="Envoyer une relance de paiement">
    <v-card-text>
      <v-row>
        <v-col align-self="center" cols="5">
          <SelectCamp
            :camps="props.camps"
            label="Séjour"
            v-model="idCamp"
            @update:model-value="fetchPreview"
          ></SelectCamp>
        </v-col>
        <v-col align-self="center">
          <v-card v-if="preview" subtitle="Dossiers à relancer">
            <template #append v-if="preview.length">
              <v-checkbox
                label="Tout sélectionner"
                density="compact"
                hide-details
                :indeterminate="
                  !!selected.length && selected.length != preview.length
                "
                :model-value="selected.length == preview.length"
                @update:model-value="
                  selected = allSelected ? [] : preview.map((v) => v.Id)
                "
              ></v-checkbox>
            </template>
            <v-card-text>
              <v-list
                selectable
                v-model:selected="selected"
                select-strategy="leaf"
                density="comfortable"
              >
                <v-list-item
                  v-for="(dossier, index) in preview"
                  :key="dossier.Id"
                  :title="dossier.Responsable"
                  :subtitle="
                    isDateZero(dossier.LastEventFacture)
                      ? ''
                      : `Dernière demande : ${Formatters.time(
                          dossier.LastEventFacture
                        )}`
                  "
                  :value="dossier.Id"
                  rounded
                  class="my-1"
                >
                  <template #append="{ isSelected, select }">
                    <v-row
                      ><v-col align-self="center">
                        <v-chip
                          class="mr-2"
                          :color="
                            Formatters.colorStatutPaiement(dossier.Bilan.Statut)
                          "
                        >
                          {{ dossier.Bilan.Recu }} /
                          {{ dossier.Bilan.Demande }}</v-chip
                        >
                      </v-col>
                      <v-col align-self="center">
                        <v-checkbox-btn
                          density="compact"
                          hide-details
                          :model-value="isSelected"
                          @update:model-value="select"
                        ></v-checkbox-btn> </v-col
                    ></v-row>
                  </template>
                </v-list-item>
              </v-list>
            </v-card-text>
          </v-card>
        </v-col>
      </v-row>
    </v-card-text>
    <v-card-actions>
      <v-spacer></v-spacer>
      <v-btn
        :disabled="!idCamp || !selected.length"
        @click="emit('send', idCamp, selected)"
      >
        <template #append>
          <v-icon>mdi-send</v-icon>
        </template>
        Envoyer
      </v-btn>
    </v-card-actions>
  </v-card>
</template>

<script setup lang="ts">
import {
  type CampItem,
  type IdCamp,
  type IdDossier,
  type PreviewRelance,
} from "@/clients/backoffice/logic/api";
import { controller } from "@/clients/backoffice/logic/logic";
import { isDateZero } from "@/components/date";
import { Formatters } from "@/utils";
import { computed, ref } from "vue";

const props = defineProps<{
  camps: CampItem[];
}>();

const emit = defineEmits<{
  (e: "send", idCamp: IdCamp, idDossiers: IdDossier[]): void;
}>();

const idCamp = ref(0 as IdCamp);

const preview = ref<PreviewRelance[]>([]);
async function fetchPreview() {
  const res = await controller.EventsSendRelancePaiementPreview({
    idCamp: idCamp.value,
  });
  if (res === undefined) return;
  preview.value = res || [];
  // select all by default
  selected.value = (res || []).map((v) => v.Id);
}

const selected = ref<IdDossier[]>([]);
const allSelected = computed(
  () => selected.value.length == preview.value.length
);
</script>
