<template>
  <v-row>
    <!-- identité -->
    <v-col align-self="center" cols="2">
      <v-list-item-title>
        {{ Personnes.label(props.participant.Personne) }}
      </v-list-item-title>
      <v-list-item-subtitle>
        {{ Camps.label(props.participant.Camp) }}
      </v-list-item-subtitle>
    </v-col>
    <!-- options sur le prix -->
    <v-col align-self="center" cols="3" class="text-center">
      <v-menu :disabled="campNoOption" :close-on-content-click="false">
        <template v-slot:activator="{ props: menuProps }">
          <v-chip label v-bind="menuProps" :elevation="campNoOption ? 0 : 1">
            {{
              props.participant.Participant.QuotientFamilial
                ? `QF ${props.participant.Participant.QuotientFamilial}.`
                : "Pas de QF."
            }}
            {{ formatOption(props.participant.Participant.OptionPrix) }}
          </v-chip>
        </template>
        <ParticipantOptionPrix
          :camp="props.participant.Camp"
          v-model:options="props.participant.Participant.OptionPrix"
          v-model:quotient-familial="
            props.participant.Participant.QuotientFamilial
          "
        ></ParticipantOptionPrix>
      </v-menu>
    </v-col>
    <!-- aides -->
    <v-col align-self="center" class="text-center">
      <v-chip
        v-for="aide in sortedAides"
        prepend-icon="mdi-cash-plus"
        color="green"
      >
        {{ Formatters.montant(aide.Valeur) }}
      </v-chip>
    </v-col>
    <!-- remises -->
    <v-col align-self="center" class="text-center" cols="4">
      <RemisesChip
        v-model="props.participant.Participant.Remises"
      ></RemisesChip>
    </v-col>
    <!-- actions -->
    <v-col align-self="center" cols="auto">
      <v-menu location="left">
        <template v-slot:activator="{ props: menuProps }">
          <v-btn
            v-bind="menuProps"
            icon="mdi-dots-vertical"
            size="small"
          ></v-btn>
        </template>
        <v-list density="comfortable">
          <v-list-item
            v-if="props.hasManyParticipants"
            title="Appliquer au dossier"
            subtitle="Dupliquer les remises et options"
          ></v-list-item>
          <v-divider></v-divider>
          <v-list-item
            title="Supprimer"
            prepend-icon="mdi-delete"
          ></v-list-item>
        </v-list>
      </v-menu>
    </v-col>
  </v-row>
</template>

<script setup lang="ts">
import {
  Currency,
  OptionPrixKind,
  type Aide,
  type Aides,
  type Dossier,
  type Int,
  type OptionPrixParticipant,
  type ParticipantExt,
} from "@/clients/backoffice/logic/api";
import { Camps, copy, Formatters, FormRules, Personnes } from "@/utils";
import { computed, ref } from "vue";
import ParticipantOptionPrix from "./ParticipantOptionPrix.vue";
import RemisesChip from "./RemisesChip.vue";

const props = defineProps<{
  participant: ParticipantExt;
  aides: Aides;
  hasManyParticipants: boolean;
}>();

const emit = defineEmits<{
  (e: "save", dossier: Dossier): void;
}>();

const aide: Aide = {
  Id: 1 as Int,
  IdStructureaide: 1 as Int,
  IdParticipant: 1 as Int,
  Valide: true,
  ParJour: false,
  NbJoursMax: 0 as Int,
  Valeur: { Cent: 1000 as Int, Currency: Currency.Euros },
};
const aides: Aides = { [1 as Int]: aide };

const sortedAides = computed(() => {
  const out = Object.values(aides || {});
  out.sort((a, b) => a.Id - b.Id);
  return out;
});

const campNoOption = computed(
  () => props.participant.Camp.OptionPrix.Active == OptionPrixKind.NoOption
);

function formatOption(opt: OptionPrixParticipant) {
  switch (props.participant.Camp.OptionPrix.Active) {
    case OptionPrixKind.NoOption:
      return "Camp sans option";
    case OptionPrixKind.PrixJour:
      const count = opt.Jour?.length || 0;
      if (count == 0) return "Présent tous les jours";
      return `Présent ${count} jour${count > 1 ? "s" : ""}`;
    case OptionPrixKind.PrixStatut:
      const statut = props.participant.Camp.OptionPrix.Statuts?.find(
        (s) => s.Id == opt.IdStatut
      );
      return statut ? `Option ${statut.Label}` : "Pas d'option.";
  }
}
</script>
