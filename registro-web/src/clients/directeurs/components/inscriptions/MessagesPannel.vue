<template>
  <v-card v-if="data != null" title="Messages des familles">
    <template #append>
      <v-btn-toggle color="secondary" class="mx-1" v-model="mode" rounded>
        <v-btn size="small" prepend-icon="mdi-clock"> Trier par date</v-btn>
        <v-btn size="small" append-icon="mdi-account"> Trier par inscrit</v-btn>
      </v-btn-toggle>

      <v-divider vertical thickness="1" class="mx-1"> </v-divider>
      <v-btn
        icon
        size="small"
        class="mx-1"
        :variant="showOnlyNew ? 'tonal' : 'elevated'"
        @click="showOnlyNew = !showOnlyNew"
      >
        <v-icon color="pink">mdi-message-badge</v-icon>
        <v-tooltip activator="parent"> Filtrer les messages non lu </v-tooltip>
      </v-btn>
    </template>
    <!-- par date -->
    <v-card-text v-if="mode == 0">
      <MessageRow
        v-for="message in filteredMessages"
        :message="message"
        @set-seen="(s) => setMessageSeen(message.Event.Id, s)"
        @start-reply="createMessageTo = message.Event.IdDossier"
        :show-reply="true"
      ></MessageRow>
    </v-card-text>

    <!-- par dossier -->
    <v-card-text v-else>
      <v-card
        v-for="dossier in byDossiers"
        :title="dossier.Dossier.Responsable"
        :subtitle="(dossier.Dossier.Participants || []).join(', ')"
        class="my-2"
      >
        <template #append>
          <v-btn
            size="small"
            @click="createMessageTo = dossier.IdDossier"
            prepend-icon="mdi-reply"
          >
            Répondre</v-btn
          >
        </template>
        <v-card-text>
          <MessageRow
            v-for="message in dossier.Messages"
            :message="message"
            @set-seen="(s) => setMessageSeen(message.Event.Id, s)"
            @start-reply="createMessageTo = message.Event.IdDossier"
            :show-reply="false"
          ></MessageRow>
        </v-card-text>
      </v-card>
    </v-card-text>

    <!-- new message -->
    <v-dialog
      :model-value="createMessageTo !== null"
      @update:model-value="createMessageTo = null"
      max-width="600px"
    >
      <v-card
        v-if="createMessageTo"
        title="Nouveau message"
        :subtitle="(data.Dossiers || {})[createMessageTo].Responsable"
      >
        <v-card-text>
          <v-textarea
            autofocus
            placeholder="Rédigez votre message..."
            v-model="newMessage"
            rows="10"
          ></v-textarea>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn
            :disabled="!newMessage.length"
            @click="sendMessage"
            prepend-icon="mdi-send"
          >
            Envoyer</v-btn
          >
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-card>
  <v-skeleton-loader v-else type="card"></v-skeleton-loader>
</template>

<script lang="ts" setup>
import { ref, onMounted, computed } from "vue";
import { controller, isMessageNew } from "../../logic/logic";
import type {
  EventExt_MessageEvt,
  IdDossier,
  IdEvent,
  Messages,
} from "../../logic/api";
import MessageRow from "./MessageRow.vue";

const props = defineProps<{}>();

onMounted(loadMessages);

const mode = ref<0 | 1>(0); // time / inscrit
const showOnlyNew = ref(false);

const filteredMessages = computed(() =>
  (data.value?.Messages || []).filter(
    (m) => !showOnlyNew.value || isMessageNew(m)
  )
);

const byDossiers = computed(() => {
  const out = new Map<IdDossier, EventExt_MessageEvt[]>();
  for (const element of filteredMessages.value) {
    out.set(
      element.Event.IdDossier,
      (out.get(element.Event.IdDossier) || []).concat(element)
    );
  }
  return Array.from(out.entries())
    .map((dossier) => ({
      IdDossier: dossier[0],
      Dossier: (data.value?.Dossiers || {})[dossier[0]],
      Messages: dossier[1].reverse(),
    }))
    .sort((a, b) => a.Dossier.Responsable.localeCompare(b.Dossier.Responsable));
});

const data = ref<Messages | null>(null);
async function loadMessages() {
  const res = await controller.ParticipantsMessagesLoad();
  if (res === undefined) return;
  data.value = res || [];
}

async function setMessageSeen(idEvent: IdEvent, seen: boolean) {
  if (!data.value) return;
  const res = await controller.ParticipantsMessageSetSeen({ idEvent, seen });
  if (res === undefined) return;
  const index = data.value.Messages?.findIndex((m) => m.Event.Id == idEvent)!;
  data.value.Messages![index] = res;
  controller.showMessage(
    seen ? "Message marqué comme lu." : "Message marqué comme non lu."
  );
}

const createMessageTo = ref<IdDossier | null>(null);
const newMessage = ref("");
async function sendMessage() {
  const idDossier = createMessageTo.value;
  if (!data.value || !idDossier) return;
  createMessageTo.value = null;
  const res = await controller.ParticipantsMessagesCreate({
    Contenu: newMessage.value,
    IdDossier: idDossier,
  });
  if (res === undefined) return;
  controller.showMessage("Message envoyé avec succès.");
  data.value.Messages = [res].concat(data.value.Messages || []);
}
</script>
