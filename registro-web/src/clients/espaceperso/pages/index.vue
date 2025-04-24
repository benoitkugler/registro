<template>
  <NavBar title="Bienvenue sur votre espace de suivi">
    <v-btn @click="showPhotos = true">
      <template #prepend>
        <v-icon>mdi-image-album</v-icon>
      </template>
      Album photos
    </v-btn>
    <v-menu>
      <template #activator="{ props: menuProps }">
        <v-btn v-bind="menuProps">
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
        ></v-list-item>
        <v-list-item
          title="Facture"
          subtitle="Télécharger au format .pdf"
          prepend-icon="mdi-invoice-list"
        ></v-list-item>
      </v-list>
    </v-menu>
  </NavBar>

  <v-skeleton-loader type="card" v-if="data == null"></v-skeleton-loader>
  <v-container class="fill-height" fluid v-else>
    <v-row>
      <!-- participants  et finances -->
      <v-col align-self="center" cols="4">
        <v-card subtitle="Participants">
          <template #append>
            <v-btn @click="showEditOptions = true" size="small">
              <template #prepend>
                <v-icon>mdi-pencil</v-icon>
              </template>
              éditer
            </v-btn>
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
                  <v-tooltip
                    :text="
                      StatutParticipantLabels[participant.Participant.Statut]
                    "
                  >
                    <template #activator="{ props: tooltipProps }">
                      <v-icon
                        v-bind="tooltipProps"
                        :icon="
                          Formatters.statutParticipant(
                            participant.Participant.Statut
                          ).icon
                        "
                      ></v-icon>
                    </template>
                  </v-tooltip>
                </template>
              </v-list-item>
            </v-list>
          </v-card-text>
        </v-card>

        <FinancesCard
          :dossier="data.Dossier"
          :token="token"
          @refresh="fetchData"
        ></FinancesCard>
      </v-col>

      <!-- fil des messages -->
      <v-col align-self="center">
        <v-card subtitle="Suivi de votre inscription">
          <template #append>
            <v-btn @click="showCreateMessage = true">
              <template #prepend>
                <v-icon>mdi-email</v-icon>
              </template>
              Nous écrire</v-btn
            >
          </template>
          <v-card-text>
            <div class="overflow-y-auto" style="max-height: 75vh">
              <v-timeline side="end" class="mt-4" density="compact">
                <EventSwitch v-for="event in events" :event="event">
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
  <v-dialog v-model="showCreateMessage" max-width="600px">
    <v-card title="Nouveau message">
      <v-card-text>
        <v-textarea
          autofocus
          placeholder="Rédigez votre message..."
          v-model="createMessageContent"
          rows="10"
        ></v-textarea>
      </v-card-text>
      <v-card-actions>
        <v-spacer></v-spacer>
        <v-btn :disabled="!createMessageContent.length" @click="sendMessage">
          <template #prepend>
            <v-icon>mdi-send</v-icon>
          </template>
          Envoyer</v-btn
        >
      </v-card-actions>
    </v-card>
  </v-dialog>

  <v-dialog v-model="showPhotos">
    <JoomeoCard :token="token"></JoomeoCard>
  </v-dialog>
</template>

<script lang="ts" setup>
import { useRouter } from "vue-router";
import { computed, onMounted, ref } from "vue";
import NavBar from "../components/NavBar.vue";
import { controller } from "../logic/logic";
import {
  StatutParticipant,
  StatutParticipantLabels,
  type Data,
  type Participant,
  type ParticipantCamp,
} from "../logic/api";
import { buildPseudoEvents, Camps, Formatters, Personnes } from "@/utils";
import ParticipantsEditCard from "../components/ParticipantsEditCard.vue";
import FinancesCard from "../components/FinancesCard.vue";
import JoomeoCard from "../components/JoomeoCard.vue";
const router = useRouter();

// id token
const token = ref("");

onMounted(() => {
  // store ID token
  const query = new URLSearchParams(window.location.search);
  token.value = query.get("token") || "";
  fetchData();
});

const data = ref<Data | null>(null);
async function fetchData() {
  const res = await controller.Load({ token: token.value });
  if (res === undefined) return;
  data.value = res;
}

const events = computed(() =>
  data.value == null ? [] : buildPseudoEvents(data.value.Dossier, "espaceperso")
);

const showCreateMessage = ref(false);
const createMessageContent = ref("");
async function sendMessage() {
  if (!createMessageContent.value.length || !data.value) return;
  showCreateMessage.value = false;
  const res = await controller.SendMessage({
    Token: token.value,
    Message: createMessageContent.value,
  });
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

const showPhotos = ref(false);
</script>
