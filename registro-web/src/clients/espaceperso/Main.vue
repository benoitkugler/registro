<template>
  <NavBar title="Espace de suivi de votre inscription">
    <v-btn @click="showSondages = 0 as IdCamp">
      <template #prepend>
        <v-icon>mdi-comment-quote</v-icon>
      </template>
      Avis
    </v-btn>
    <v-btn @click="showPhotos = true">
      <template #prepend>
        <v-icon>mdi-image-album</v-icon>
      </template>
      Album photos
    </v-btn>
    <v-menu>
      <template #activator="{ props: menuProps }">
        <v-btn v-bind="menuProps" :disabled="!enableJustificatifs">
          <template #prepend>
            <v-icon>mdi-download</v-icon>
          </template>
          Justificatifs
        </v-btn>
      </template>
      <v-list>
        <v-list-item
          title="Attestation de présence"
          subtitle="Télécharger au format .pdf"
          prepend-icon="mdi-file-document-check"
          link
          :href="endpoints.DownloadAttestationPresence(token)"
        ></v-list-item>
        <v-list-item
          title="Facture"
          subtitle="Télécharger au format .pdf"
          prepend-icon="mdi-invoice-list"
          :href="endpoints.DownloadFacture(token)"
        ></v-list-item>
      </v-list>
    </v-menu>
  </NavBar>

  <v-skeleton-loader type="card" v-if="data == null"></v-skeleton-loader>
  <v-container class="fill-height" fluid v-else>
    <v-row>
      <!-- participants  et finances -->
      <v-col align-self="center" cols="5">
        <v-card subtitle="Participants">
          <template #append>
            <v-btn size="small" class="mr-1">
              <template #prepend>
                <v-icon>mdi-folder</v-icon>
              </template>
              <v-badge
                :color="allDocumentsToFillCount ? 'pink' : 'transparent'"
                :content="allDocumentsToFillCount || ''"
                floating
              >
                Documents
              </v-badge>

              <v-menu activator="parent">
                <v-list>
                  <v-list-item
                    title="Documents du séjour"
                    subtitle="Lettre du directeur, ..."
                    prepend-icon="mdi-mail"
                    @click="showDocuments = true"
                  >
                    <template #append v-if="data.DocumentsToReadOrFillCount">
                      <v-badge
                        color="pink"
                        :content="data.DocumentsToReadOrFillCount"
                        inline
                      >
                      </v-badge>
                    </template>
                  </v-list-item>
                  <v-list-item
                    title="Fiches sanitaires"
                    prepend-icon="mdi-pill"
                    @click="showFichesantaires = true"
                  >
                    <template #append v-if="data.FichesanitaireToFillCount">
                      <v-badge
                        color="pink"
                        :content="data.FichesanitaireToFillCount"
                        inline
                      >
                      </v-badge>
                    </template>
                  </v-list-item>
                </v-list>
              </v-menu>
            </v-btn>

            <!-- TODO -->
            <!-- <v-btn @click="showEditOptions = true" size="small">
              <template #prepend>
                <v-icon>mdi-pencil</v-icon>
              </template>
              Options
            </v-btn> -->
          </template>
          <v-card-text>
            <v-list class="pa-0">
              <v-list-item
                v-for="participant in data.Dossier.Participants"
                :title="Personnes.label(participant.Personne)"
                :subtitle="Camps.label(participant.Camp)"
                rounded
                :class="{
                  'my-2': true,
                  [participantColorClass(participant)]: true,
                }"
              >
                <template #append>
                  <v-chip
                    variant="elevated"
                    :color="
                      Formatters.statutParticipant(
                        participant.Participant.Statut
                      ).color
                    "
                    :prepend-icon="
                      Formatters.statutParticipant(
                        participant.Participant.Statut
                      ).icon
                    "
                  >
                    {{
                      StatutParticipantLabels[participant.Participant.Statut]
                    }}
                  </v-chip>
                </template>
              </v-list-item>
            </v-list>
          </v-card-text>
        </v-card>

        <FinancesCard
          :token="token"
          :dossier="data.Dossier"
          :is-paiement-open="data.IsPaiementOpen"
          :settings="data.PaiementSettings"
          @refresh="fetchData"
          @show-reglement="showReglement = true"
        ></FinancesCard>
      </v-col>

      <!-- fil des messages -->
      <v-col align-self="center">
        <v-card subtitle="Fil de suivi de votre inscription">
          <template #append>
            <v-btn
              @click="showCreateMessage = { content: '', toFondSoutien: false }"
            >
              <template #prepend>
                <v-icon>mdi-email</v-icon>
              </template>
              Nous écrire</v-btn
            >
          </template>
          <v-card-text>
            <div class="overflow-y-auto" style="max-height: 75vh">
              <v-timeline side="end" class="mt-4" density="compact">
                <EventSwitch
                  v-for="event in events"
                  :event="event"
                  @go-to-sondage="(id) => (showSondages = id)"
                  @go-to-documents="showDocuments = true"
                  @go-to-validation="showValidation = true"
                  @accept-place-liberee="(id) => handleFromEvent(id)"
                  @reply-fond-soutien="
                    showCreateMessage = { content: '', toFondSoutien: true }
                  "
                >
                </EventSwitch>
              </v-timeline>
            </div>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>

    <!-- option edit -->
    <v-dialog v-model="showEditOptions">
      <ParticipantsEditCard
        :participants="data.Dossier.Participants || []"
        @save="updateParticipants"
      ></ParticipantsEditCard>
    </v-dialog>
  </v-container>

  <!-- new message -->
  <v-dialog
    :model-value="showCreateMessage != null"
    @update:model-value="showCreateMessage = null"
    max-width="600px"
  >
    <v-card
      v-if="showCreateMessage"
      title="Nouveau message"
      :subtitle="
        showCreateMessage.toFondSoutien
          ? 'Ce message ne sera visible que par le fonds de soutien.'
          : 'Ce message sera visible par le centre et les directeurs.'
      "
    >
      <v-card-text>
        <v-textarea
          autofocus
          placeholder="Rédigez votre message..."
          v-model="showCreateMessage.content"
          rows="10"
        ></v-textarea>
      </v-card-text>
      <v-card-actions>
        <v-spacer></v-spacer>
        <v-btn
          :disabled="!showCreateMessage.content.length"
          @click="sendMessage"
        >
          <template #prepend>
            <v-icon>mdi-send</v-icon>
          </template>
          Envoyer</v-btn
        >
      </v-card-actions>
    </v-card>
  </v-dialog>

  <v-dialog
    :model-value="showSondages != null"
    @update:model-value="showSondages = null"
  >
    <SondagesCard
      :token="token"
      :initial-camp="showSondages"
      v-if="showSondages != null"
    ></SondagesCard>
  </v-dialog>

  <v-dialog v-model="showPhotos">
    <JoomeoCard :token="token"></JoomeoCard>
  </v-dialog>

  <v-dialog v-model="showFichesantaires">
    <FichessanitairesCard
      :token="token"
      @close="showFichesantaires = false"
      @update-notifs="v => data!.FichesanitaireToFillCount = v"
    ></FichessanitairesCard>
  </v-dialog>

  <v-dialog v-model="showDocuments" max-width="800px">
    <DocumentsCard
      :token="token"
      @close="showDocuments = false"
      @update-notifs="v => data!.DocumentsToReadOrFillCount = v"
    ></DocumentsCard>
  </v-dialog>

  <!-- reglement -->
  <v-dialog v-if="data" v-model="showReglement" max-width="800px">
    <FinancesReglementCard
      :token="token"
      :dossier="data.Dossier"
      :settings="data.PaiementSettings"
    ></FinancesReglementCard>
  </v-dialog>

  <!-- présentation initiale -->
  <v-dialog v-model="showPresentation" max-width="800px">
    <PresentationCard
      v-if="data"
      :mail-centre="data.MailCentre"
    ></PresentationCard>
  </v-dialog>

  <!-- présentation initiale -->
  <v-dialog v-model="showValidation" max-width="800px">
    <ValidationInscriptionCard v-if="data" :dossier="data.Dossier">
    </ValidationInscriptionCard>
  </v-dialog>

  <!-- confirme accept place -->
  <v-dialog
    :model-value="showConfirmAccept != null"
    @update:model-value="showConfirmAccept = null"
    max-width="600px"
  >
    <v-card title="Accepter la place" v-if="showConfirmAccept">
      <v-card-text>
        Confirmez-vous l'inscription de
        {{ showConfirmAccept.Content.Data.ParticipantLabel }}
        sur le séjour
        {{ showConfirmAccept.Content.Data.CampLabel }} ? <br /><br />
        <span class="text-grey"
          >Un mail de notification sera envoyé automatiquement au centre.</span
        >
      </v-card-text>
      <v-card-actions>
        <v-spacer></v-spacer>
        <v-btn color="success" @click="acceptePlaceLiberee">Confirmer</v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script lang="ts" setup>
import { computed, onMounted, ref } from "vue";
import NavBar from "./components/NavBar.vue";
import { controller } from "./logic/logic";
import {
  Acteur,
  EventContentKind,
  StatutParticipant,
  StatutParticipantLabels,
  type Data,
  type Event,
  type IdCamp,
  type IdEvent,
  type Participant,
  type ParticipantCamp,
  type PlaceLiberee,
  type SendMessageIn,
} from "./logic/api";
import { buildPseudoEvents, Camps, Formatters, Personnes } from "@/utils";
import ParticipantsEditCard from "./components/ParticipantsEditCard.vue";
import FinancesCard from "./components/FinancesCard.vue";
import JoomeoCard from "./components/JoomeoCard.vue";
import FichessanitairesCard from "./components/FichessanitairesCard.vue";
import SondagesCard from "./components/SondagesCard.vue";
import DocumentsCard from "./components/DocumentsCard.vue";
import { endpoints } from "@/clients/directeurs/logic/logic";
import FinancesReglementCard from "./components/FinancesReglementCard.vue";
import PresentationCard from "./components/PresentationCard.vue";
import ValidationInscriptionCard from "./components/ValidationInscriptionCard.vue";

// id token
const token = ref("");

onMounted(async () => {
  // process query params
  const query = new URLSearchParams(window.location.search);
  const tokenS = query.get("token") || "";
  const fromInscription = !!query.get("from-inscription");
  const fromIdEvent = Number(query.get("idEvent") || 0) as IdEvent;

  // always store the token and load the data
  token.value = tokenS;
  await fetchData();
  if (!data.value) return;

  // then, handle from inscription
  if (fromInscription) {
    showPresentation.value = true;
    return;
  }

  // and the possibly event linking to the page
  handleFromEvent(fromIdEvent);
});

const data = ref<Data | null>(null);
async function fetchData() {
  const res = await controller.Load({ token: token.value });
  if (res === undefined) return;
  data.value = res;
}

const events = computed(() =>
  data.value == null
    ? []
    : buildPseudoEvents(data.value.Dossier, Acteur.Espaceperso)
);

const allDocumentsToFillCount = computed(() =>
  data.value
    ? data.value.DocumentsToReadOrFillCount +
      data.value.FichesanitaireToFillCount
    : 0
);

function handleFromEvent(fromIdEvent: IdEvent) {
  if (!data.value) return;
  const event = (data.value.Dossier.Events || []).find(
    (ev) => ev.Id == fromIdEvent
  );
  if (!event) return;
  switch (event.Content.Kind) {
    case EventContentKind.Validation:
      showValidation.value = true;
      return;
    case EventContentKind.Facture:
      showReglement.value = true;
      return;
    case EventContentKind.PlaceLiberee:
      showConfirmAccept.value = event as EventPlaceLiberee;
      return;
    case EventContentKind.CampDocs:
      showDocuments.value = true;
      return;
    case EventContentKind.Sondage:
      showSondages.value = event.Content.Data.IdCamp;
      return;
    default:
      // TODO
      return;
  }
}

const enableJustificatifs = computed(() => {
  if (data.value == null) return false;
  return (
    (data.value.Dossier.Participants || []).find(
      (p) => p.Participant.Statut == StatutParticipant.Inscrit
    ) != undefined
  );
});

const showCreateMessage = ref<{
  content: string;
  toFondSoutien: boolean;
} | null>(null);
async function sendMessage() {
  if (!showCreateMessage.value || !data.value) return;
  const args: SendMessageIn = {
    Token: token.value,
    Message: showCreateMessage.value.content,
    OnlyToFondSoutien: showCreateMessage.value.toFondSoutien,
  };
  showCreateMessage.value = null;
  const res = await controller.SendMessage(args);
  if (res === undefined) return;
  controller.showMessage("Message envoyé avec succès.");
  data.value.Dossier.Events = (data.value.Dossier.Events || []).concat(res);
}

function participantColorClass(p: ParticipantCamp) {
  if (p.Participant.Statut == StatutParticipant.Inscrit) {
    return "bg-lime-lighten-2";
  } else if (p.Participant.Statut == StatutParticipant.AStatuer) {
    return "bg-grey-lighten-1";
  }
  return "bg-orange-lighten-3";
}

const showEditOptions = ref(false);
async function updateParticipants(params: Participant[]) {
  showEditOptions.value = false;
  const res = await controller.UpdateParticipants({
    Token: token.value,
    Participants: params,
  });
  if (res === undefined) return;
  controller.showMessage("Modifications enregistrées avec succès. Merci !");
}

const showSondages = ref<IdCamp | null>(null);

const showPhotos = ref(false);

const showFichesantaires = ref(false);

const showDocuments = ref(false);

const showReglement = ref(false);

// mail reçu après avoir validé (tout ou partie)
const showValidation = ref(false);

// redirection après l'inscription (avant toute validation)
const showPresentation = ref(false);

type EventPlaceLiberee = Event & {
  Content: { Kind: "PlaceLiberee"; Data: PlaceLiberee };
};

const showConfirmAccept = ref<EventPlaceLiberee | null>(null);
async function acceptePlaceLiberee() {
  const event = showConfirmAccept.value;
  if (!event) return;
  showConfirmAccept.value = null;
  const res = await controller.AcceptePlaceLiberee({
    token: token.value,
    idEvent: event.Id,
  });
  if (res === undefined) return;
  fetchData();
  controller.showMessage("Place acceptée avec succès.");
}
</script>
