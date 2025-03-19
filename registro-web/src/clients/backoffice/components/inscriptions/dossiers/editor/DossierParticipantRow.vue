<template>
  <v-row>
    <!-- identité -->
    <v-col align-self="center" cols="3">
      <v-list-item-title>
        <a
          href="/camps"
          @click.prevent="goToParticipant(props.participant.Participant)"
        >
          {{ Personnes.label(props.participant.Personne) }}
        </a>
      </v-list-item-title>
      <v-list-item-subtitle>
        {{ Camps.label(props.participant.Camp) }}
      </v-list-item-subtitle>
    </v-col>
    <!-- options sur le prix -->
    <v-col align-self="center" cols="2" class="text-center">
      <v-menu
        :disabled="campNoOption"
        :close-on-content-click="false"
        v-model="showMenuOption"
      >
        <template #activator="{ props: menuProps }">
          <v-card
            v-bind="menuProps"
            :elevation="campNoOption ? 0 : 1"
            color="grey-lighten-2"
          >
            <v-card-text class="py-1 px-1">
              <div v-if="campNoOption">Camp sans option</div>
              <div v-else>
                {{
                  props.participant.Participant.QuotientFamilial
                    ? `QF ${props.participant.Participant.QuotientFamilial}.`
                    : "Pas de QF."
                }}
                <br />
                {{ formatOption(props.participant.Participant.OptionPrix) }}
              </div>
            </v-card-text>
          </v-card>
        </template>
        <ParticipantOptionPrix
          :camp="props.participant.Camp"
          :option-and-qf="{
            options: props.participant.Participant.OptionPrix,
            quotientFamilial: props.participant.Participant.QuotientFamilial,
          }"
          @update="
            (oq) => {
              props.participant.Participant.OptionPrix = oq.options;
              props.participant.Participant.QuotientFamilial =
                oq.quotientFamilial;
              showMenuOption = false;
              emit('update', props.participant.Participant);
            }
          "
        ></ParticipantOptionPrix>
      </v-menu>
    </v-col>
    <!-- aides -->
    <v-col align-self="center" class="text-center">
      <v-chip v-if="!sortedAides.length" label>Aucune aide</v-chip>
      <v-chip
        v-for="aide in sortedAides"
        class="my-1"
        prepend-icon="mdi-cash-plus"
        :color="aide.Valide ? 'green' : 'orange'"
        @click="aideToUpdate = aide"
      >
        <template #close>
          <v-icon icon="mdi-close-circle" @click.stop="aideToRemove = aide" />
        </template>
        {{ props.structures[aide.IdStructureaide].Nom }} :
        {{ Formatters.montant(aide.Valeur) }}
      </v-chip>
    </v-col>
    <!-- remises -->
    <v-col align-self="center" class="text-center" cols="3">
      <RemisesChip
        :remises="props.participant.Participant.Remises"
        @update="
          (remises) => {
            props.participant.Participant.Remises = remises;
            emit('update', props.participant.Participant);
          }
        "
      ></RemisesChip>
    </v-col>
    <!-- actions -->
    <v-col align-self="center" cols="auto">
      <v-menu location="left">
        <template #activator="{ props: menuProps }">
          <v-btn
            v-bind="menuProps"
            icon="mdi-dots-vertical"
            size="small"
          ></v-btn>
        </template>
        <v-list density="comfortable">
          <v-list-item
            title="Ajouter une aide"
            prepend-icon="mdi-cash-plus"
            @click="
              aideToCreate = {
                IdParticipant: props.participant.Participant.Id,
                IdStructure: 0 as IdStructureaide,
              }
            "
          ></v-list-item>
          <v-divider></v-divider>
          <v-list-item
            v-if="props.hasManyParticipants"
            prepend-icon="mdi-arrow-expand"
            title="Étendre les options et remises"
            subtitle="... au reste du dossier"
            @click="emit('expand')"
          ></v-list-item>
          <v-divider></v-divider>
          <v-list-item
            title="Supprimer"
            prepend-icon="mdi-delete"
            @click="showConfirmeDelete = true"
          ></v-list-item>
        </v-list>
      </v-menu>
    </v-col>

    <!-- suppression participant -->
    <v-dialog v-model="showConfirmeDelete" max-width="600px">
      <v-card title="Supprimer le participant">
        <v-card-text>
          Confirmez-vous la suppression de ce participant ? <br />

          Attention, cette opération est irréversible.
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn
            color="red"
            @click="
              emit('delete');
              showConfirmeDelete = false;
            "
            >Supprimer</v-btn
          >
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- création aide -->
    <v-dialog
      v-if="aideToCreate != null"
      :model-value="aideToCreate != null"
      @update:model-value="aideToCreate = null"
      max-width="400px"
    >
      <v-card title="Ajouter une aide extérieure">
        <v-card-text>
          <v-select
            variant="outlined"
            density="comfortable"
            label="Structure"
            :items="structureItems"
            :model-value="
              aideToCreate.IdStructure == 0 ? null : aideToCreate.IdStructure
            "
            @update:model-value="(v) => (aideToCreate!.IdStructure = v)"
          ></v-select>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn
            color="green"
            @click="
              emit('createAide', aideToCreate);
              aideToCreate = null;
            "
            :disabled="aideToCreate.IdStructure == 0"
            >Créer</v-btn
          >
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- suppression aide -->
    <v-dialog
      v-if="aideToRemove != null"
      :model-value="aideToRemove != null"
      @update:model-value="aideToRemove = null"
      max-width="600px"
    >
      <v-card title="Supprimer l'aide">
        <v-card-text>
          Confirmez-vous la suppression de l'aide extérieure ? <br />

          Attention, cette opération est irréversible.
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn
            color="red"
            @click="
              emit('deleteAide', aideToRemove.Id);
              aideToRemove = null;
            "
            >Supprimer</v-btn
          >
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- edition aide -->
    <v-dialog
      v-if="aideToUpdate != null"
      :model-value="aideToUpdate != null"
      @update:model-value="aideToUpdate = null"
      max-width="800px"
    >
      <AideEditCard
        :aide="aideToUpdate"
        :structures="props.structures"
        :participant="props.participant"
        @save="
          (aide) => {
            emit('updateAide', aide);
            aideToUpdate = null;
          }
        "
      ></AideEditCard>
    </v-dialog>
  </v-row>
</template>

<script setup lang="ts">
import {
  OptionPrixKind,
  type Aide,
  type Aides,
  type AidesCreateIn,
  type IdAide,
  type IdStructureaide,
  type OptionPrixParticipant,
  type Participant,
  type ParticipantCamp,
  type Structureaides,
} from "@/clients/backoffice/logic/api";
import { Camps, Formatters, Personnes } from "@/utils";
import { computed, ref } from "vue";
import ParticipantOptionPrix from "./ParticipantOptionPrix.vue";
import RemisesChip from "./RemisesChip.vue";
import AideEditCard from "./AideEditCard.vue";
import { goToParticipant } from "@/clients/backoffice/plugins/router";

const props = defineProps<{
  participant: ParticipantCamp;
  aides: Aides;
  structures: NonNullable<Structureaides>;
  hasManyParticipants: boolean;
}>();

const emit = defineEmits<{
  (e: "update", participant: Participant): void;
  (e: "delete"): void;
  (e: "createAide", args: AidesCreateIn): void;
  (e: "updateAide", aide: Aide): void;
  (e: "deleteAide", id: IdAide): void;
  (e: "expand"): void;
}>();

const sortedAides = computed(() => {
  const out = Object.values(props.aides || {});
  out.sort((a, b) => a.Id - b.Id);
  return out;
});

const campNoOption = computed(
  () =>
    props.participant.Camp.OptionPrix.Active == OptionPrixKind.NoOption &&
    !Camps.isQuotientFamilialActive(
      props.participant.Camp.OptionQuotientFamilial
    )
);

function formatOption(opt: OptionPrixParticipant) {
  switch (props.participant.Camp.OptionPrix.Active) {
    case OptionPrixKind.NoOption:
      return "Séjour sans option";
    case OptionPrixKind.PrixJour:
      const count = opt.Jour?.length || 0;
      if (count == 0) return "Présent tous les jours";
      return `Présent ${count} jour${count > 1 ? "s" : ""}`;
    case OptionPrixKind.PrixStatut:
      const statut = props.participant.Camp.OptionPrix.Statuts?.find(
        (s) => s.Id == opt.IdStatut
      );
      return statut ? `Statut ${statut.Label}` : "Pas d'option.";
  }
}

const showMenuOption = ref(false);
const showConfirmeDelete = ref(false);

const structureItems = computed(() =>
  Object.values(props.structures).map((s) => ({ value: s.Id, title: s.Nom }))
);
const aideToCreate = ref<AidesCreateIn | null>(null);
const aideToUpdate = ref<Aide | null>(null);
const aideToRemove = ref<Aide | null>(null);
</script>
