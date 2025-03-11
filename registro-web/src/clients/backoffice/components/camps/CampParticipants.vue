<template>
  <v-card
    v-if="data != null"
    title="Participants"
    :subtitle="Camps.label(data.Camp.Camp)"
  >
    <v-card-text class="mt-4">
      <v-skeleton-loader type="table" v-if="isLoading"></v-skeleton-loader>
      <div v-else>
        <div class="text-center font-italic" v-if="!participants.length">
          Aucun participant n'est encore inscrit sur ce séjour.
        </div>
        <v-row v-for="p in participants">
          <v-col align-self="center" cols="3">
            <v-list-item :title="Personnes.NOMPrenom(p.Personne)">
              <template #prepend>
                <v-tooltip
                  v-if="p.Participant.Statut != ListeAttente.Inscrit"
                  :text="ListeAttenteLabels[p.Participant.Statut]"
                >
                  <template #activator="{ props }">
                    <v-icon v-bind="props"> mdi-clock </v-icon>
                  </template>
                </v-tooltip>

                <v-tooltip
                  v-else-if="p.HasBirthday"
                  :text="`${p.Personne.Prenom} a son anniveraire pendant le camp !`"
                >
                  <template #activator="{ props }">
                    <v-icon v-bind="props" color="amber"
                      >mdi-cake-variant</v-icon
                    >
                  </template>
                </v-tooltip>
              </template>
            </v-list-item>
          </v-col>

          <v-col align-self="center" cols="auto">
            <v-icon>
              {{ Formatters.sexeIcon(p.Personne.Sexe) }}
            </v-icon>
          </v-col>
          <v-col align-self="center" cols="auto">
            {{ Formatters.dateNaissance(p.Personne.DateNaissance) }}
          </v-col>
          <v-col align-self="center" cols="1" class="text-center">
            {{
              ageFrom(
                p.Personne.DateNaissance,
                new Date(data.Camp.Camp.DateDebut)
              )
            }}
            ans
          </v-col>
          <v-col align-self="center">
            {{ NavettteLabels[p.Participant.Navette] }}
          </v-col>
          <v-col align-self="center">
            {{ p.Participant.Details }}
          </v-col>
        </v-row>
      </div>
    </v-card-text>
  </v-card>
  <v-skeleton-loader v-else type="card"></v-skeleton-loader>
</template>

<script lang="ts" setup>
import { ref, onMounted, computed } from "vue";
import { controller } from "@/clients/backoffice/logic/logic";
import {
  ListeAttente,
  ListeAttenteLabels,
  Navettte,
  NavettteLabels,
  type CampHeader,
  type CampsLoadOut,
  type IdCamp,
  type ParticipantPersonne,
} from "@/clients/backoffice/logic/api";
import { Camps, Formatters, Personnes } from "@/utils";
import { da } from "vuetify/locale";
import { ageFrom } from "@/components/date";

const props = defineProps<{
  id: IdCamp;
}>();

onMounted(loadCamp);

const isLoading = ref(false);

// with sort
const participants = computed(() => {
  const out = (data.value?.Participants || []).map((p) => p);
  out.sort((a, b) => {
    const sa = a.Participant.Statut;
    const sb = b.Participant.Statut;
    // By liste attente : Inscrit is higher
    if (sa != sb) return sb - sa;
    // By name :
    return Personnes.NOMPrenom(a.Personne).localeCompare(
      Personnes.NOMPrenom(b.Personne)
    );
  });
  return out;
});

const data = ref<CampsLoadOut | null>(null);
async function loadCamp() {
  isLoading.value = true;
  const res = await controller.CampsLoad({ idCamp: props.id });
  isLoading.value = false;
  if (res === undefined) return;
  data.value = res;
}
</script>
