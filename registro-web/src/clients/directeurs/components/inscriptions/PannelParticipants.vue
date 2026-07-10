<template>
  <v-card v-if="data != null">
    <template #title>
      <v-btn flat icon>
        <v-icon>mdi-information</v-icon>
        <v-tooltip activator="parent" content-class="ma-0 pa-0">
          <StatistiquesCard
            :participants="data.Participants || []"
            :statistiques="data.Statistiques"
          ></StatistiquesCard>
        </v-tooltip>
      </v-btn>
      Participants
    </template>
    <template #append>
      <v-btn @click="showMessages = true" class="mx-1" prepend-icon="mdi-email">
        Messages</v-btn
      >
      <v-btn
        @click="showDocuments = true"
        class="mx-1"
        prepend-icon="mdi-folder"
      >
        Documents</v-btn
      >
      <v-btn
        @click="showGroupes = true"
        class="mx-1"
        prepend-icon="mdi-account-group"
      >
        Groupes</v-btn
      >

      <v-divider thickness="1" vertical class="mx-1"></v-divider>

      <v-btn icon flat>
        <v-icon>mdi-dots-vertical</v-icon>
        <v-menu activator="parent">
          <v-list density="compact">
            <v-list-item
              title="Trier par moment d'inscription"
              @click="sortByTime = !sortByTime"
              prepend-icon="mdi-sort-calendar-ascending"
              :append-icon="sortByTime ? 'mdi-check' : ''"
            >
            </v-list-item>

            <v-divider thickness="1"></v-divider>

            <v-list-item
              prepend-icon="mdi-file-excel"
              title="Exporter"
              subtitle="au format Excel"
              link
              :href="controller.ParticipantsDownloadListe(controller.authToken)"
            ></v-list-item>
            <v-divider thickness="1"></v-divider>

            <v-list-item
              prepend-icon="mdi-pill"
              title="Fiches sanitaires"
              subtitle="Analyser le contenu des fiches sanitaires"
              @click="showFichesSanitaires = true"
            ></v-list-item>
            <v-list-item
              prepend-icon="mdi-currency-eur"
              title="Suivi du règlement"
              @click="showReglements = true"
            ></v-list-item>
          </v-list>
        </v-menu>
      </v-btn>
    </template>
    <v-card-text class="mt-4">
      <v-skeleton-loader type="table" v-if="isLoading"></v-skeleton-loader>
      <div v-else>
        <ParticipantRowHeader>
          <template #append v-if="hasGroupes">
            <v-col class="font-weight-bold text-center" align-self="center"
              >Groupe</v-col
            >
          </template>
        </ParticipantRowHeader>

        <div class="text-center font-italic mt-4" v-if="!participants.length">
          Aucun participant n'est encore inscrit sur ce séjour.
        </div>

        <ParticipantRow
          v-for="(p, index) in participants"
          :participant="p.participant"
          :index="index"
          :groupe="p.groupe"
        >
          <template #append v-if="hasGroupes">
            <v-col align-self="center" class="text-center">
              <v-chip
                :color="p.groupe ? p.groupe.Couleur : ''"
                label
                elevation="1"
              >
                {{ p.groupe ? p.groupe.Nom : "Aucun groupe" }}
                <v-menu activator="parent">
                  <v-list density="compact">
                    <v-list-item
                      v-for="groupe in Object.values(groupes.Groupes || {})"
                      :title="groupe.Nom"
                      @click.once="
                        setParticipantGroupe(p.participant, groupe.Id)
                      "
                    >
                      <template #prepend>
                        <v-badge :color="groupe.Couleur" inline></v-badge>
                      </template>
                    </v-list-item>
                    <v-divider></v-divider>
                    <v-list-item
                      title="Retirer"
                      @click.once="
                        setParticipantGroupe(p.participant, -1 as IdGroupe)
                      "
                      prepend-icon="mdi-close"
                    ></v-list-item>
                  </v-list>
                </v-menu>
              </v-chip>
            </v-col>
          </template>
          <template #actions>
            <v-list density="comfortable">
              <v-list-item
                :disabled="
                  p.participant.Participant.Statut != StatutParticipant.AStatuer
                "
                title="Aller à l'inscription"
                prepend-icon="mdi-arrow-left-top-bold"
                @click="
                  emit('goToInscription', p.participant.Participant.IdDossier)
                "
              ></v-list-item>
              <v-list-item
                title="Détails"
                subtitle="Responsable, ..."
                prepend-icon="mdi-account-details"
                @click="participantDetailsToShow = p"
              ></v-list-item>
              <v-divider horizontal thickness="1"></v-divider>
              <v-list-item
                title="Modifier"
                prepend-icon="mdi-pencil"
                @click="toEdit = p.participant"
              ></v-list-item>
            </v-list>
          </template>
        </ParticipantRow>
      </div>
    </v-card-text>

    <!-- edit participant -->
    <v-dialog
      v-if="toEdit != null"
      :model-value="toEdit != null"
      @update:model-value="toEdit = null"
      max-width="800px"
    >
      <ParticipantEdit
        :participant="toEdit.Participant"
        :personne="toEdit.Personne"
        :api="{
          SelectPersonne: controller.SelectPersonne.bind(controller),
        }"
        hide-personne-dossier
        readonly-statut
        @save="updateParticipant"
      ></ParticipantEdit>
    </v-dialog>

    <!-- fiches sanitaires -->
    <v-dialog v-model="showFichesSanitaires">
      <FichesSanitairesPannel></FichesSanitairesPannel>
    </v-dialog>

    <!-- documents demandés -->
    <v-dialog v-model="showDocuments" max-width="800px">
      <DocumentsPannel></DocumentsPannel>
    </v-dialog>

    <!-- messages -->
    <v-dialog v-model="showMessages" max-width="1200px">
      <MessagesPannel></MessagesPannel>
    </v-dialog>

    <!-- règlement -->
    <v-dialog v-model="showReglements" max-width="700px">
      <ReglementsCard :data="data"></ReglementsCard>
    </v-dialog>

    <!-- groupes -->
    <v-dialog v-model="showGroupes" max-width="700px">
      <GroupesPannel
        :groupes="groupes"
        :participants="data.Participants || []"
        @refresh="loadGroupes"
      ></GroupesPannel>
    </v-dialog>

    <!-- détails responsable -->
    <v-dialog
      :model-value="participantDetailsToShow != null"
      @update:model-value="participantDetailsToShow = null"
      max-width="600px"
    >
      <v-card title="Détails" v-if="participantDetailsToShow != null">
        <v-card-text>
          <v-card class="my-2" subtitle="Participant">
            <v-card-text>
              <v-row>
                <v-col cols="7">
                  <v-chip prepend-icon="mdi-account">{{
                    Personnes.NOMPrenom(
                      participantDetailsToShow.participant.Personne
                    )
                  }}</v-chip>
                </v-col>

                <v-col
                  class="text-right"
                  v-if="
                    participantDetailsToShow.participant.Personne.Nationnalite
                      .IsSuisse
                  "
                >
                  <v-chip prepend-icon="mdi-passport"
                    >Nationnalité Suisse
                  </v-chip>
                </v-col>
              </v-row>
            </v-card-text>
          </v-card>

          <v-card subtitle="Responsable" class="my-2">
            <v-card-text>
              <v-row>
                <v-col cols="7">
                  <v-chip prepend-icon="mdi-account">{{
                    Personnes.NOMPrenom(
                      participantDetailsToShow.dossier.Responsable
                    )
                  }}</v-chip></v-col
                >
                <v-col class="text-right">
                  <v-chip
                    prepend-icon="mdi-passport"
                    v-if="
                      participantDetailsToShow.dossier.Responsable.Nationnalite
                        .IsSuisse
                    "
                    >Nationnalité Suisse
                  </v-chip>
                </v-col>
              </v-row>

              <v-row
                ><v-col>
                  <v-chip prepend-icon="mdi-email">{{
                    participantDetailsToShow.dossier.Responsable.Mail
                  }}</v-chip>
                </v-col></v-row
              >
              <v-row
                ><v-col>
                  <v-chip prepend-icon="mdi-phone">
                    {{
                      Formatters.tels(
                        participantDetailsToShow.dossier.Responsable.Tels
                      )
                    }}
                  </v-chip></v-col
                ></v-row
              >
              <v-row
                ><v-col>
                  <v-chip prepend-icon="mdi-home-city">
                    {{ participantDetailsToShow.dossier.Responsable.Adresse }}
                    {{
                      participantDetailsToShow.dossier.Responsable.CodePostal
                    }}
                    {{ participantDetailsToShow.dossier.Responsable.Ville }}
                    {{ participantDetailsToShow.dossier.Responsable.Pays }}
                  </v-chip></v-col
                ></v-row
              >
            </v-card-text>
          </v-card>

          <v-card subtitle="Préférences">
            <v-card-text>
              <v-chip
                :append-icon="
                  participantDetailsToShow.dossier.Dossier.PartageAdressesOK
                    ? 'mdi-check'
                    : 'mdi-close'
                "
                :color="
                  participantDetailsToShow.dossier.Dossier.PartageAdressesOK
                    ? 'green'
                    : 'red'
                "
              >
                Diffusion des coordonnées pour covoiturage
              </v-chip>
            </v-card-text>
          </v-card>
        </v-card-text>
      </v-card>
    </v-dialog>
  </v-card>
  <v-skeleton-loader v-else type="card"></v-skeleton-loader>
</template>

<script lang="ts" setup>
import { ref, onMounted, computed } from "vue";
import { controller } from "../../logic/logic";
import {
  StatutParticipant,
  type Date_,
  type DossierRecap,
  type Groupe,
  type GroupeParticipant,
  type GroupesOut,
  type IdDossier,
  type IdGroupe,
  type Participant,
  type ParticipantExt,
  type ParticipantsOut,
  type Personne,
} from "../../logic/api";
import { Formatters, Participants, Personnes } from "@/utils";
import FichesSanitairesPannel from "./FichesSanitairesPannel.vue";
import DocumentsPannel from "./DocumentsPannel.vue";
import MessagesPannel from "./MessagesPannel.vue";
import StatistiquesCard from "./StatistiquesCard.vue";
import GroupesPannel from "./GroupesPannel.vue";

const props = defineProps<{}>();

const emit = defineEmits<{
  (e: "goToInscription", dossier: IdDossier): void;
}>();

onMounted(() => {
  loadGroupes();
  loadParticipants();
});

defineExpose({ loadParticipants });

const isLoading = ref(false);

type participantData = {
  participant: ParticipantExt;
  dossier: DossierRecap;
  groupe: Groupe | null;
};

// with sort and groupe
const sortByTime = ref(false);
const participants = computed(() => {
  const d = data.value;
  if (!d) return [];
  const g = groupes.value;
  const dossiers = d.Dossiers || {};
  const out = (d.Participants || []).map((p) => {
    const link: GroupeParticipant | undefined = (g.ParticipantsToGroupe || {})[
      p.Participant.Id
    ];
    return {
      participant: p,
      dossier: dossiers[p.Participant.IdDossier],
      groupe: link ? (g.Groupes || {})[link.IdGroupe] : null,
    } satisfies participantData;
  });

  out.sort((a, b) =>
    Participants.cmp(a.participant, b.participant, sortByTime.value)
  );
  return out;
});
const hasGroupes = computed(
  () => Object.values(groupes.value.Groupes || {}).length != 0
);

const data = ref<ParticipantsOut | null>(null);
async function loadParticipants() {
  isLoading.value = true;
  const res = await controller.ParticipantsGet();
  isLoading.value = false;
  if (res === undefined) return;
  data.value = res;
}

const groupes = ref<GroupesOut>({
  Groupes: {},
  ParticipantsToGroupe: {},
  MinHint: "" as Date_,
  MaxHint: "" as Date_,
});
async function loadGroupes() {
  const res = await controller.GroupesGet();
  if (res === undefined) return;
  groupes.value = res;
}

const toEdit = ref<ParticipantExt | null>(null);
async function updateParticipant(p: Participant) {
  if (toEdit.value == null || data.value == null) return;
  const res = await controller.ParticipantsUpdate(p);
  toEdit.value = null;
  if (res === undefined) return;
  controller.showMessage("Participant modifié avec succès.");
  // reload the list
  loadParticipants();
}

async function setParticipantGroupe(
  participant: ParticipantExt,
  idGroupe: IdGroupe
) {
  const res = await controller.ParticipantSetGroupe({
    idParticipant: participant.Participant.Id,
    idGroupe,
  });
  if (res === undefined) return;
  controller.showMessage("Groupe attribué avec succès.");
  loadGroupes();
}

const showFichesSanitaires = ref(false);

const showDocuments = ref(false);

const showMessages = ref(false);

const showReglements = ref(false);

const showGroupes = ref(false);

const participantDetailsToShow = ref<participantData | null>(null);
</script>
