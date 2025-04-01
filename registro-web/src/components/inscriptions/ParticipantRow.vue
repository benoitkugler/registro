<template>
  <v-row no-gutters :class="cl">
    <v-col align-self="center" cols="3">
      <v-list-item
        :title="Personnes.NOMPrenom(props.participant.Personne)"
        class="px-0"
      >
        <template #prepend>
          <v-tooltip
            v-if="
              props.participant.Participant.Statut != StatutParticipant.Inscrit
            "
            :text="
              StatutParticipantLabels[props.participant.Participant.Statut]
            "
          >
            <template #activator="{ props: tooltipProps }">
              <v-icon
                :color="
                  Formatters.statutParticipant(
                    props.participant.Participant.Statut
                  ).color
                "
                v-bind="tooltipProps"
              >
                {{
                  Formatters.statutParticipant(
                    props.participant.Participant.Statut
                  ).icon
                }}
              </v-icon>
            </template>
          </v-tooltip>
        </template>
      </v-list-item>
    </v-col>

    <v-col align-self="center" cols="1">
      <v-icon class="ma-2 mr-4">
        {{ Formatters.sexeIcon(props.participant.Personne.Sexe) }}
      </v-icon>
    </v-col>
    <v-col align-self="center" cols="1">
      <v-row no-gutters>
        <v-col align-self="end">
          {{
            Formatters.dateNaissance(props.participant.Personne.DateNaissance)
          }}
        </v-col>
        <v-col align-self="end" cols="auto">
          <v-tooltip
            v-if="props.participant.HasBirthday"
            :text="`${props.participant.Personne.Prenom} a son anniveraire pendant le camp !`"
          >
            <template #activator="{ props }">
              <v-icon v-bind="props" color="pink">mdi-cake-variant</v-icon>
            </template>
          </v-tooltip>
        </v-col>
      </v-row>
    </v-col>
    <v-col align-self="center" cols="1" class="text-center">
      {{ props.participant.Age }} ans
    </v-col>
    <v-col align-self="center" cols="1" class="text-center">
      {{
        props.participant.Participant.Navette == Navette.NoBus
          ? "-"
          : NavetteLabels[props.participant.Participant.Navette]
      }}
    </v-col>
    <v-col align-self="center">
      {{ props.participant.Participant.Details }}
    </v-col>
    <v-spacer></v-spacer>
    <v-col cols="2" align-self="center">
      {{ Formatters.time(props.participant.MomentInscription) }}
    </v-col>
    <v-col align-self="center" cols="auto">
      <v-menu>
        <template #activator="{ props: menuProps }">
          <v-btn
            v-bind="menuProps"
            icon="mdi-dots-vertical"
            size="x-small"
          ></v-btn>
        </template>

        <slot name="actions"></slot>
      </v-menu>
    </v-col>
  </v-row>
</template>

<script setup lang="ts">
import { Formatters, Personnes } from "@/utils";
import {
  StatutParticipant,
  StatutParticipantLabels,
  Navette,
  NavetteLabels,
  type ParticipantExt,
} from "../../clients/backoffice/logic/api";
import { computed } from "vue";
const props = defineProps<{
  participant: ParticipantExt;
  index: number;
}>();

const cl = computed(
  () => "mx-2 px-2 rounded " + (props.index % 2 == 0 ? "bg-grey-lighten-4" : "")
);
</script>
