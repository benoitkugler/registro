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
              :subtitle="`Inscription ${participant.IdInscription} ; ${Formatters.time(data.Inscriptions![participant.IdInscription].DateHeure)}`"
            >
              <template #append>
                <v-btn
                  size="small"
                  icon="mdi-account-details"
                  @click="
                    toShowDetails =
                      data.Inscriptions![participant.IdInscription]
                  "
                ></v-btn>
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
  </v-card>
</template>

<script setup lang="ts">
import {
  SexeLabels,
  type IdPersonne,
  type Inscription,
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
</script>
