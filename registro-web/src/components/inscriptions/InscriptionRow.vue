<template>
  <v-card
    :subtitle="
      Formatters.time(props.inscription.Dossier.MomentInscription, true, true)
    "
  >
    <template #append>
      <v-tooltip
        v-if="alreadyValidated"
        text="En attente de validation par les autres séjours"
      >
        <template #activator="{ props: tooltipProps }">
          <v-icon v-bind="tooltipProps" color="green"> mdi-check </v-icon>
        </template>
      </v-tooltip>
      <v-btn v-else :disabled="!allIdentified" @click="emit('valide')">
        <template #prepend>
          <v-icon>mdi-check</v-icon>
        </template>
        Valider</v-btn
      >
      <v-menu v-if="!props.hideDelete">
        <template #activator="{ props: menuProps }">
          <v-btn
            v-bind="menuProps"
            variant="flat"
            icon="mdi-dots-vertical"
            size="small"
            class="ml-1"
          ></v-btn>
        </template>
        <v-list density="compact">
          <v-list-item
            title="Supprimer"
            prepend-icon="mdi-delete"
            @click="emit('delete')"
          ></v-list-item>
        </v-list>
      </v-menu>
    </template>
    <v-card-text>
      <v-row>
        <v-col cols="7">
          <v-row no-gutters class="my-0 py-1">
            <InscriptionEtatcivilCols
              :api="props.api"
              :personne="props.inscription.Responsable"
              @identifie="(v) => emit('identifie', v)"
            ></InscriptionEtatcivilCols>
          </v-row>
          <v-divider thickness="1"></v-divider>
          <v-row
            no-gutters
            class="my-0 py-1"
            v-for="(part, i) in props.inscription.Participants"
            :key="i"
          >
            <InscriptionEtatcivilCols
              :api="props.api"
              :personne="part.Personne"
              @identifie="(v) => emit('identifie', v)"
            ></InscriptionEtatcivilCols>
            <v-col align-self="center" cols="3" class="text-center">
              {{ Camps.label(part.Camp) }}
            </v-col>
          </v-row>
        </v-col>
        <v-col cols="5">
          <v-row no-gutters>
            <v-col cols="auto" class="px-2">
              <v-icon
                :color="props.inscription.Message.length ? 'secondary' : 'grey'"
                >mdi-message</v-icon
              >
            </v-col>
            <v-col>
              <div
                v-for="(line, i) in props.inscription.Message.split('\n')"
                :key="i"
              >
                {{ line }}
              </div>
            </v-col>
          </v-row>
        </v-col>
      </v-row>
    </v-card-text>
  </v-card>
</template>

<script setup lang="ts">
import { Camps, Formatters } from "@/utils";
import {
  type IdentTarget,
  type Inscription,
} from "../../clients/backoffice/logic/api";
import InscriptionEtatcivilCols from "./InscriptionEtatcivilCols.vue";
import { computed } from "vue";
import type { SimilairesAPI } from "./types";
const props = defineProps<{
  inscription: Inscription;
  api: SimilairesAPI;
  hideDelete?: boolean;
  alreadyValidated?: boolean;
}>();

const emit = defineEmits<{
  (e: "identifie", params: IdentTarget): void;
  (e: "valide"): void;
  (e: "delete"): void;
}>();

const allIdentified = computed(
  () =>
    !props.inscription.Responsable.IsTemp &&
    !!props.inscription.Participants?.every((pr) => !pr.Personne.IsTemp)
);
</script>
