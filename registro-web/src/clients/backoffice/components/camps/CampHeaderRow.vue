<template>
  <v-card :color="periodeColor" no-gutters @click="emit('click')" class="my-1">
    <v-row
      :style="{ 'background-color': periodeColor }"
      class="px-1"
      no-gutters
      justify="space-between"
    >
      <v-col>
        <v-list-item
          :title="Camps.label(props.camp.Camp.Camp)"
          :subtitle="`ID : ${props.camp.Camp.Camp.Id}`"
        >
        </v-list-item>
      </v-col>
      <v-col align-self="center">
        {{ Camps.formatPlage(camp.Camp.Camp) }}</v-col
      >
      <v-col cols="2" align-self="center">
        <v-tooltip width="450px" location="left" content-class="pa-0">
          <template #activator="{ props: innerProps }">
            <v-progress-linear
              class="bg-white border-md border-primary"
              v-bind="innerProps"
              :max="props.camp.Camp.Camp.Places"
              :model-value="props.camp.Stats.Valides"
              height="36"
              rounded
              color="primary"
            >
              <strong
                >{{ props.camp.Stats.Valides }}/{{
                  props.camp.Camp.Camp.Places
                }}
                <span v-if="allAttente > 0">
                  (+
                  {{ allAttente }})</span
                >
              </strong>
              <v-icon class="ml-2">mdi-account-multiple</v-icon>
            </v-progress-linear>
          </template>
          <CampStats :stats="props.camp.Stats"></CampStats>
        </v-tooltip>
      </v-col>
      <v-col cols="1" align-self="center" class="px-2">
        <v-tooltip width="600" location="left" content-class="pa-0">
          <template #activator="{ props: innerProps }">
            <v-progress-linear
              class="bg-white border-md border-primary"
              v-bind="innerProps"
              :max="100"
              :model-value="fileUploadProgress"
              height="36"
              rounded
              color="primary"
            >
              <strong>{{ fileUploadProgress }} % </strong>
              <v-icon class="ml-2">mdi-file-upload</v-icon>
            </v-progress-linear>
          </template>
          <v-card title="Pièces justificatives">
            <v-card-text>
              <v-list>
                <v-list-item
                  v-for="demande in props.camp.ParticipantsFiles"
                  :title="demande.Title"
                >
                  <template #append>
                    {{ demande.UploadedCount }} /
                    {{ demande.InscritsCount }} ({{
                      Formatters.pourcent(
                        demande.UploadedCount,
                        demande.InscritsCount
                      )
                    }}
                    %)
                  </template>
                </v-list-item>
              </v-list>
            </v-card-text>
          </v-card>
        </v-tooltip>
      </v-col>
      <v-col cols="1"></v-col>
      <v-col cols="auto" align-self="center">
        <v-menu>
          <template #activator="{ props: innerProps }">
            <v-btn
              v-bind="innerProps"
              size="x-small"
              class="mx-1"
              icon="mdi-dots-vertical"
            ></v-btn>
          </template>
          <v-list density="compact">
            <v-list-item prepend-icon="mdi-pencil" @click="emit('edit')"
              >Modifier les paramètres</v-list-item
            >

            <v-list-item
              prepend-icon="mdi-currency-eur"
              @click="emit('edit-taux')"
              >Taux de conversion</v-list-item
            >

            <v-divider thickness="1"></v-divider>
            <v-list-item
              prepend-icon="mdi-mail"
              @click="emit('show-documents')"
              title="Documents"
            ></v-list-item>

            <v-divider thickness="1"></v-divider>
            <v-list-item
              prepend-icon="mdi-plus"
              @click="showAddDirecteur = true"
              title="Ajouter un directeur..."
              :disabled="props.camp.HasDirecteur"
            ></v-list-item>
            <v-divider thickness="1"></v-divider>

            <v-list-item
              title="Envoyer le sondage"
              subtitle="de fin de séjour"
              prepend-icon="mdi-comment-quote"
              @click="emit('send-sondage')"
            >
            </v-list-item>

            <v-divider thickness="1"></v-divider>
            <v-list-item
              prepend-icon="mdi-delete"
              :disabled="props.camp.Stats.Inscriptions > 0"
              @click="showDelete = true"
              >Supprimer</v-list-item
            >
          </v-list>
        </v-menu>
      </v-col>
    </v-row>
  </v-card>

  <v-dialog v-model="showDelete" max-width="600">
    <v-card title="Confirmer la suppression">
      <v-card-text>
        Etes vous certain de supprimer le séjour
        <b>{{ Camps.label(props.camp.Camp.Camp) }}</b> ? <br />
        <br />
        Les éventuels <i>équipiers</i> déclarés sur ce séjour seront aussi
        supprimés.

        <br /><br />
        Attention, cette opération est irréversible.
      </v-card-text>
      <v-card-actions>
        <v-spacer></v-spacer>
        <v-btn
          color="red"
          @click="
            showDelete = false;
            emit('delete');
          "
          >Supprimer</v-btn
        >
      </v-card-actions>
    </v-card>
  </v-dialog>

  <v-dialog v-model="showAddDirecteur" max-width="400px">
    <v-card title="Ajouter un directeur">
      <v-card-text>
        <SelectPersonne
          initial-personne=""
          label="Directeur"
          v-model="idDirecteur"
          :api="controller"
        ></SelectPersonne>
      </v-card-text>
      <v-card-actions>
        <v-spacer></v-spacer>
        <v-btn
          :disabled="!idDirecteur"
          @click="
            showAddDirecteur = false;
            emit('add-directeur', idDirecteur);
          "
          >Ajouter</v-btn
        >
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
import { ref, computed } from "vue";
import type {
  CampHeader,
  Camp,
  IdPersonne,
  DemandeStat,
} from "@/clients/backoffice/logic/api";
import CampStats from "./CampStats.vue";
import { Camps, Formatters } from "@/utils";
import { controller } from "../../logic/logic";
const props = defineProps<{
  camp: CampHeader;
}>();

const emit = defineEmits<{
  (e: "click"): void;
  (e: "edit"): void;
  (e: "edit-taux"): void;
  (e: "delete"): void;
  (e: "show-documents"): void;
  (e: "send-sondage"): void;
  (e: "add-directeur", id: IdPersonne): void;
}>();

const allAttente = computed(
  () => props.camp.Stats.Inscriptions - props.camp.Stats.Valides
);

const showDelete = ref(false);

/** renvoie la couleur de la période du séjour */
const periodeColor = computed(() => {
  const month = new Date(props.camp.Camp.Camp.DateDebut).getUTCMonth() + 1;
  switch (month) {
    case 7:
    case 8: // Ete
      return "rgb(45, 185, 187)";
    case 9:
    case 10:
    case 11: // Automne
      return "rgb(190, 150, 60)";
    case 12:
    case 1:
    case 2:
    case 3: // Hiver
      return "rgb(240, 240, 240)";
    case 4:
    case 5:
    case 6:
    default: // Printemps
      return "rgb(190, 228, 100)";
  }
});

const showAddDirecteur = ref(false);
const idDirecteur = ref(0 as IdPersonne);

const fileUploadProgress = computed(() => {
  const files = props.camp.ParticipantsFiles || [];
  if (files.length == 0) return 0;
  let p = 0;
  for (const file of files) {
    if (file.InscritsCount == 0) continue;
    p += file.UploadedCount / file.InscritsCount;
  }
  return Math.round((p / files.length) * 100);
});
</script>
