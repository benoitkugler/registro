<template>
  <v-card
    title="Envoyer une relance de confirmation"
    subtitle="Sélectionner les inscriptions à relancer"
  >
    <template #append>
      <v-checkbox
        label="Tout sélectionner"
        density="compact"
        hide-details
        :indeterminate="
          !!selected.length && selected.length != props.inscriptions.length
        "
        :model-value="selected.length == props.inscriptions.length"
        @update:model-value="
          selected = allSelected
            ? [] // remove all
            : props.inscriptions.map((v) => v.Inscription.Id) // select all
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
          v-for="(inscription, index) in props.inscriptions"
          :key="inscription.Inscription.Id"
          :title="Personnes.label(inscription.Inscription.Responsable)"
          :subtitle="Formatters.time(inscription.Inscription.DateHeure)"
          :value="inscription.Inscription.Id"
          rounded
          class="my-1"
        >
          <template #append="{ isSelected, select }">
            <v-row>
              <v-col>
                {{ inscription.Inscription.Responsable.Mail }}
              </v-col>
              <v-col>
                <v-checkbox-btn
                  density="compact"
                  hide-details
                  :model-value="isSelected"
                  @update:model-value="select"
                ></v-checkbox-btn>
              </v-col>
            </v-row>
          </template>
        </v-list-item>
      </v-list>
    </v-card-text>
    <v-card-actions>
      <v-spacer></v-spacer>
      <v-btn
        :disabled="!selected.length"
        @click="emit('send', selected)"
        prepend-icon="mdi-send"
      >
        Envoyer
      </v-btn>
    </v-card-actions>
  </v-card>
</template>

<script setup lang="ts">
import {
  type IdInscription,
  type PendingInscription,
} from "@/clients/backoffice/logic/api";
import { Formatters, Personnes } from "@/utils";
import { computed, ref } from "vue";

const props = defineProps<{
  inscriptions: PendingInscription[];
}>();

const emit = defineEmits<{
  (e: "send", ids: IdInscription[]): void;
}>();

const selected = ref<IdInscription[]>([]);
const allSelected = computed(
  () => selected.value.length == props.inscriptions.length
);
</script>
