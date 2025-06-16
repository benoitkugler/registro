<template>
  <v-card title="Envoyer les documents d'un séjour">
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
          <v-card v-if="preview" subtitle="Dossiers à contacter">
            <v-card-text>
              <v-list
                selectable
                v-model:selected="selected"
                select-strategy="leaf"
              >
                <v-list-item
                  v-for="dossier in preview.Dossiers"
                  :title="dossier.Responsable"
                  :subtitle="dossier.Participants"
                  :value="dossier.Id"
                  rounded
                  class="my-1"
                >
                  <template #prepend v-if="dossier.DocumentsSent">
                    <v-chip color="green" class="mr-2">Envoyés</v-chip>
                  </template>

                  <template #append="{ isSelected, select }">
                    <v-checkbox-btn
                      density="compact"
                      hide-details
                      :model-value="isSelected"
                      @update:model-value="select"
                    ></v-checkbox-btn>
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
        :disabled="!idCamp || !selected.length || isSending"
        @click="
          isSending = true;
          emit('send', idCamp, selected);
        "
      >
        <template #append>
          <v-progress-circular
            v-if="isSending"
            indeterminate
          ></v-progress-circular>
          <v-icon v-else>mdi-send</v-icon>
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
  type SendDocumentsCampPreview,
} from "@/clients/backoffice/logic/api";
import { controller } from "@/clients/backoffice/logic/logic";
import { ref } from "vue";

const props = defineProps<{
  camps: CampItem[];
}>();

const emit = defineEmits<{
  (e: "send", idCamp: IdCamp, idDossiers: IdDossier[]): void;
}>();

const idCamp = ref(0 as IdCamp);

const preview = ref<SendDocumentsCampPreview | null>(null);
async function fetchPreview() {
  const res = await controller.EventsSendDocumentsCampPreview({
    idCamp: idCamp.value,
  });
  if (res === undefined) return;
  preview.value = res;
  // by default, select the one not send yet
  selected.value = (res.Dossiers || [])
    .filter((d) => !d.DocumentsSent)
    .map((d) => d.Id);
}

const selected = ref<IdDossier[]>([]);

const isSending = ref(false);
</script>
