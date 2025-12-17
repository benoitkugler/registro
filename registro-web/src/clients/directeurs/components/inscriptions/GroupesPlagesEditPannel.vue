<template>
  <v-card
    title="Modifier les plages"
    subtitle="Vous pouvez ajuster la plage de date de naissance de chaque groupe."
  >
    <v-card-text>
      <v-alert v-if="sizes.isMissing" type="warning">
        Attention, les plages définies ne prennent pas en compte tout les
        inscrits.
      </v-alert>
      <v-list>
        <v-list-item
          v-for="groupe in Object.values(inner)"
          :title="groupe.Nom"
          :subtitle="plages[groupe.Id]"
        >
          <template #prepend>
            <v-badge inline :color="groupe.Couleur"></v-badge>
          </template>
          <template #append>
            <v-row style="width: 300px">
              <v-col align-self="center" cols="8" class="my-4">
                <GroupesDateSlider
                  :groupes="props.groupes"
                  v-model="groupe.Fin"
                ></GroupesDateSlider>
              </v-col>
              <v-col align-self="center" cols="4">
                <v-chip prepend-icon="mdi-account-multiple">
                  {{ sizes.sizes[groupe.Id] || 0 }}
                </v-chip>
              </v-col>
            </v-row>
          </template>
        </v-list-item>
      </v-list>
    </v-card-text>
    <v-card-actions>
      <v-btn
        prepend-icon="mdi-check"
        @click="
          () =>
            hasManuel() ? (showConfirme = true) : emit('save', inner, true)
        "
        >Enregistrer et appliquer</v-btn
      >
    </v-card-actions>

    <v-dialog v-model="showConfirme" max-width="600px">
      <v-card title="Mise à jour des groupes">
        <v-card-text>
          Certains participants ont déjà un groupe, affecté manuellement. <br />
          Vous pouvez conserver ce choix ou le remplacer par celui définit par
          les dates de naissance.
        </v-card-text>
        <v-card-actions>
          <v-btn @click="emit('save', inner, false)"
            >Conserver les affectations manuelles</v-btn
          >
          <v-spacer></v-spacer>
          <v-btn @click="emit('save', inner, true)">Tout mettre à jour</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-card>
</template>

<script lang="ts" setup>
import { computed, reactive, ref } from "vue";
import {
  type Groupes,
  type GroupesOut,
  type ParticipantExt,
} from "../../logic/api";
import { groupesPlages, groupesSizes } from "./groupes";
import { copy } from "@/utils";
import GroupesDateSlider from "./GroupesDateSlider.vue";

const props = defineProps<{
  groupes: GroupesOut;
  participants: ParticipantExt[];
}>();

const emit = defineEmits<{
  (e: "save", groupes: Groupes, erase: boolean): void;
}>();

const inner = reactive(copy(props.groupes.Groupes || {}));

const plages = computed(() => groupesPlages(inner));
const sizes = computed(() => groupesSizes(inner, props.participants));

function hasManuel() {
  for (const element of Object.values(
    props.groupes.ParticipantsToGroupe || {}
  )) {
    if (element.Manuel) return true;
  }
  return false;
}
const showConfirme = ref(false);
</script>
