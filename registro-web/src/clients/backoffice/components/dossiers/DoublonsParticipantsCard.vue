<template>
  <v-card
    title="Doublons"
    subtitle="Détecte les participants inscrits sur plusieurs séjours."
  >
    <v-skeleton-loader type="table" v-if="!data"></v-skeleton-loader>
    <v-card-text v-else>
      <v-card
        v-for="group in data.Participants"
        class="my-2"
        :title="`${Personnes.label(group![0])}`"
        :subtitle="`${SexeLabels[group![0].Sexe]} - ${Formatters.dateNaissance(group![0].DateNaissance)}`"
      >
        <v-card-text>
          <v-list>
            <v-list-item
              v-for="participant in group"
              :title="Camps.label(data.Camps![participant.IdCamp])"
              :subtitle="`ID : ${participant.Id} ; ID inscription : ${participant.IdInscription} ; ${Formatters.time(data.Inscriptions![participant.IdInscription].DateHeure)}`"
            >
              <template #append>
                <v-row no-gutters>
                  <v-col align-self="center" class="mr-2">
                    <v-btn
                      size="small"
                      icon="mdi-account-details"
                      @click="
                        toShowDetails =
                          data.Inscriptions![participant.IdInscription]
                      "
                    ></v-btn>
                  </v-col>
                  <v-col align-self="center">
                    <v-btn size="small" @click="toMarkDoublon = participant"
                      >Marquer comme doublon</v-btn
                    >
                  </v-col>
                </v-row>
              </template>
            </v-list-item>
          </v-list>
        </v-card-text>
      </v-card>
    </v-card-text>

    <!-- détails -->
    <v-dialog
      :model-value="toShowDetails != null"
      @update:model-value="toShowDetails = null"
      max-width="800px"
    >
      <v-card title="Données de l'inscription" v-if="toShowDetails">
        <v-card-text>
          <pre style="font-size: 10pt">{{
            JSON.stringify(toShowDetails, null, 2)
          }}</pre>
        </v-card-text>
      </v-card>
    </v-dialog>

    <!-- confirme doublons -->
    <v-dialog
      :model-value="toMarkDoublon != null"
      @update:model-value="toMarkDoublon = null"
      max-width="800px"
    >
      <v-card title="Confirmation" v-if="toMarkDoublon">
        <v-card-text>
          Cette inscription (ID : {{ toMarkDoublon.Id }}) sera marquée comme
          <b>doublon</b> et n'apparaîtra plus dans cette liste. <br /><br />
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn @click="markDoublon">Confirmer</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-card>
</template>

<script setup lang="ts">
import {
  SexeLabels,
  type IdPersonne,
  type Inscription,
  type InscriptionParticipant,
  type InscriptionsDoublonsOut,
} from "@/clients/backoffice/logic/api";
import { controller } from "@/clients/backoffice/logic/logic";
import { Camps, Formatters, Personnes } from "@/utils";
import { onMounted, ref } from "vue";

const props = defineProps<{
  //   dossier: DossierExt;
}>();

const emit = defineEmits<{
  (e: "create", idResponsable: IdPersonne): void;
}>();

onMounted(fetchDoublons);

const data = ref<InscriptionsDoublonsOut | null>(null);
async function fetchDoublons() {
  const res = await controller.InscriptionsSearchDoublons();
  if (res === undefined) return;
  data.value = res;
}

const toShowDetails = ref<Inscription | null>(null);

const toMarkDoublon = ref<InscriptionParticipant | null>(null);
async function markDoublon() {
  const v = toMarkDoublon.value;
  if (v == null) return;
  toMarkDoublon.value = null;
  const res = await controller.InscriptionsMarkDoublon({
    id: v.Id,
  });
  if (res === undefined) return;
  fetchDoublons();
  controller.showMessage("Inscription marquée comme doublon.");
}
</script>
