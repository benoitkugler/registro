<template>
  <v-card
    title="Essais d'inscription"
    subtitle="Adresse mail non confirmée"
    class="ma-2"
  >
    <template #append>
      <v-btn @click="showRelance = true" prepend-icon="mdi-send"
        >Envoyer une relance...</v-btn
      >
    </template>
    <v-card-text>
      <v-skeleton-loader v-if="data == null"></v-skeleton-loader>
      <div
        class="text-center font-italic"
        v-else-if="!data.Inscriptions?.length"
      >
        Il n'y a aucune inscription en attente de confirmation.
      </div>
      <div v-else>
        <v-row v-for="insc in data.Inscriptions">
          <v-col align-self="center" cols="1" class="text-center">
            <v-chip size="small" label>ID : {{ insc.Inscription.Id }}</v-chip>
          </v-col>
          <v-col align-self="center" cols="3">
            <v-list-item-title>{{
              Personnes.label(insc.Inscription.Responsable)
            }}</v-list-item-title>
            <v-list-item-subtitle>{{
              Formatters.time(insc.Inscription.DateHeure)
            }}</v-list-item-subtitle>
          </v-col>
          <v-col align-self="center" class="text-center" cols="5">
            <v-chip v-for="participant in insc.Participants" class="my-1">
              {{ Personnes.label(participant) }} :
              {{ Camps.label((data.Camps || {})[participant.IdCamp]) }}
            </v-chip>
          </v-col>
          <v-col align-self="center" class="text-center">
            {{ insc.Inscription.Responsable.Mail }}
          </v-col>
          <v-col align-self="center" cols="auto">
            <v-btn icon size="small" flat>
              <v-icon>mdi-dots-vertical</v-icon>
              <v-menu activator="parent">
                <v-list>
                  <v-list-item
                    prepend-icon="mdi-pencil"
                    title="Modifier..."
                    @click="toEdit = copy(insc.Inscription)"
                  ></v-list-item>
                  <v-list-item
                    prepend-icon="mdi-delete"
                    title="Supprimer"
                    @click="toDelete = insc.Inscription"
                  ></v-list-item>
                  <v-list-item
                    prepend-icon="mdi-account-details"
                    title="Voir les détails"
                    @click="toShowDetails = insc"
                  ></v-list-item>
                </v-list>
              </v-menu>
            </v-btn>
          </v-col>
        </v-row>
      </div>
    </v-card-text>

    <!-- edit -->
    <v-dialog
      :model-value="toEdit != null"
      @update:model-value="toEdit = null"
      max-width="800px"
    >
      <v-card title="Modifier l'inscription" v-if="toEdit">
        <v-card-text>
          <v-row>
            <v-col>
              <v-text-field
                autofocus
                label="Mail"
                variant="outlined"
                density="compact"
                v-model="toEdit.Responsable.Mail"
                :rules="[FormRules.validMail()]"
              ></v-text-field>
            </v-col>
          </v-row>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn :disabled="!toEdit.Responsable.Mail" @click="updateInscription"
            >Enregistrer</v-btn
          >
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- confirmation -->
    <v-dialog
      :model-value="toDelete != null"
      @update:model-value="toDelete = null"
      max-width="600"
    >
      <v-card title="Confirmer la suppression" v-if="toDelete">
        <v-card-text>
          Etes vous certain de supprimer l'inscription effectuée par
          <b>{{ Personnes.label(toDelete.Responsable) }}</b> (ID :
          {{ toDelete.Id }}) ?
          <br />
          <br />
          Attention, cette opération est irréversible.
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="red" @click="deleteInscription">Supprimer</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

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

    <!-- relance -->
    <v-dialog v-model="showRelance" max-width="800px">
      <SendRelanceConfirmation
        :inscriptions="data?.Inscriptions || []"
        @send="sendRelance"
      ></SendRelanceConfirmation>
    </v-dialog>

    <!-- relance monitor -->
    <v-dialog
      :model-value="sendProgress != null"
      v-if="sendProgress"
      max-width="600px"
    >
      <RequestProgressCard
        title="Envoi des relances"
        :progress="sendProgress"
      ></RequestProgressCard>
    </v-dialog>
  </v-card>
</template>

<script setup lang="ts">
import { onMounted, ref } from "vue";
import { controller } from "../../logic/logic";
import {
  type IdInscription,
  type Inscription,
  type Int,
  type PendingInscription,
  type PendingInscriptionsOut,
  type SendProgress,
} from "../../logic/api";
import {
  Camps,
  copy,
  Formatters,
  FormRules,
  Personnes,
  readJSONStream,
} from "@/utils";
import SendRelanceConfirmation from "./pending/SendRelanceConfirmation.vue";
import RequestProgressCard from "@/components/RequestProgressCard.vue";

const props = defineProps<{}>();

const emit = defineEmits<{}>();

onMounted(fetchInscriptions);

const data = ref<PendingInscriptionsOut | null>(null);

async function fetchInscriptions() {
  const res = await controller.InscriptionsGetPending();
  if (res === undefined) return;
  data.value = res;
}

const toShowDetails = ref<PendingInscription | null>(null);

const toEdit = ref<Inscription | null>(null);
async function updateInscription() {
  const insc = toEdit.value;
  if (!insc || !data.value) return;
  toEdit.value = null;
  const res = await controller.InscriptionsUpdatePending({
    Id: insc.Id,
    Mail: insc.Responsable.Mail,
  });
  if (res === undefined) return;
  controller.showMessage("Adresse mail modifiée avec succès.");
  // update local view
  data.value.Inscriptions!.find(
    (item) => item.Inscription.Id == insc.Id
  )!.Inscription = insc;
}

const toDelete = ref<Inscription | null>(null);
async function deleteInscription() {
  const insc = toDelete.value;
  if (!insc || !data.value) return;
  toDelete.value = null;
  const res = await controller.InscriptionsDeletePending({ id: insc.Id });
  if (res === undefined) return;

  controller.showMessage("Inscription supprimée avec succès.");

  // delete from this view
  data.value.Inscriptions = (data.value.Inscriptions || []).filter(
    (val) => val.Inscription.Id != insc.Id
  );
}

const showRelance = ref(false);

const sendProgress = ref<SendProgress | null>(null);
async function sendRelance(ids: IdInscription[]) {
  showRelance.value = false;
  // start with initial 0 progress
  sendProgress.value = {
    Current: 0 as Int,
    Total: 10 as Int, // just a guess
  };
  const res = await controller.InscriptionsRelancePending({ Ids: ids });
  if (res === undefined) {
    sendProgress.value = null;
    return;
  }
  await readJSONStream(
    res,
    (v) => (sendProgress.value = v),
    (err) => controller.onError("Envoi du mail de relance", err)
  );
  sendProgress.value = null;
  controller.showMessage("Mails de relance envoyés avec succès.");
}
</script>
