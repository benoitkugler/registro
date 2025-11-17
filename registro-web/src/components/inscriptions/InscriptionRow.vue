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
      <v-btn
        v-else
        :disabled="!allIdentified || !props.inscription.Participants?.length"
        @click="emit('valide')"
        prepend-icon="mdi-check-all"
      >
        Tout valider</v-btn
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
            prepend-icon="mdi-file-move"
            @click="emit('merge')"
            title="Fusionner vers ..."
          ></v-list-item>
          <v-divider></v-divider>
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
              :acteur="acteur"
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
              :acteur="acteur"
              @identifie="(v) => emit('identifie', v)"
            ></InscriptionEtatcivilCols>
            <v-col align-self="center" cols="3" class="text-center">
              {{ Camps.label(part.Camp) }}
            </v-col>
            <v-spacer></v-spacer>

            <v-col align-self="center" cols="1" class="pr-2 text-center">
              <!-- participant déjà validé -->
              <template
                v-if="part.Participant.Statut != StatutParticipant.AStatuer"
              >
                <v-icon
                  :color="
                    Formatters.statutParticipant(part.Participant.Statut).color
                  "
                  :icon="
                    Formatters.statutParticipant(part.Participant.Statut).icon
                  "
                >
                </v-icon>
              </template>
              <template v-else-if="isValidable(part.Participant)">
                <v-btn
                  size="small"
                  title="Valider ce participant..."
                  @click="emit('valide', part.Participant.Id)"
                >
                  <v-icon
                    :icon="
                      Formatters.statutParticipant(StatutParticipant.AStatuer)
                        .icon
                    "
                  ></v-icon>
                  <v-icon>mdi-arrow-right</v-icon>
                  <v-icon
                    :icon="
                      Formatters.statutParticipant(
                        (props.inscription.StatutHints || {})[
                          part.Participant.Id
                        ].Statut
                      ).icon
                    "
                    :color="
                      Formatters.statutParticipant(
                        (props.inscription.StatutHints || {})[
                          part.Participant.Id
                        ].Statut
                      ).color
                    "
                  >
                  </v-icon>
                </v-btn>
              </template>
            </v-col>
            <v-col align-self="center" cols="auto" v-if="!props.hideDelete">
              <v-btn
                icon
                size="x-small"
                @click="toDelete = part"
                title="Supprimer ce participant..."
              >
                <v-icon color="red">mdi-delete</v-icon>
              </v-btn>
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

    <!-- confirme delete -->
    <v-dialog
      :model-value="toDelete != null"
      @update:model-value="toDelete = null"
      max-width="600px"
    >
      <v-card title="Confirmation" v-if="toDelete">
        <v-card-text>
          Confirmez-vous la suppression du participant
          <b>{{ Personnes.label(toDelete.Personne) }}</b> ?

          <br /><br />
          Attention, cette opération est irréversible.
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn
            color="red"
            @click="
              emit('deleteParticipant', toDelete.Participant.Id);
              toDelete = null;
            "
            >Supprimer</v-btn
          >
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-card>
</template>

<script setup lang="ts">
import { Camps, Formatters, Personnes } from "@/utils";
import {
  Acteur,
  StatutParticipant,
  type IdCamp,
  type IdentTarget,
  type IdParticipant,
  type Inscription,
  type Participant,
  type ParticipantCamp,
} from "../../clients/backoffice/logic/api";
import InscriptionEtatcivilCols from "./InscriptionEtatcivilCols.vue";
import { computed, ref } from "vue";
import type { SimilairesAPI } from "../types";
const props = defineProps<{
  inscription: Inscription;
  api: SimilairesAPI;
  user: IdCamp | null; // null for admin
  hideDelete?: boolean;
  alreadyValidated?: boolean;
}>();

const emit = defineEmits<{
  (e: "identifie", params: IdentTarget): void;
  (e: "valide", idParticipant?: IdParticipant): void;
  (e: "merge"): void;
  (e: "delete"): void;
  (e: "deleteParticipant", id: IdParticipant): void;
}>();

const acteur = computed(() =>
  props.user == null ? Acteur.Backoffice : Acteur.Directeur
);

const allIdentified = computed(
  () =>
    !props.inscription.Responsable.IsTemp &&
    !!props.inscription.Participants?.every((pr) => !pr.Personne.IsTemp)
);

function isValidable(part: Participant) {
  return (
    allIdentified.value && (props.user === null || props.user == part.IdCamp)
  );
}

const toDelete = ref<ParticipantCamp | null>(null);
</script>
