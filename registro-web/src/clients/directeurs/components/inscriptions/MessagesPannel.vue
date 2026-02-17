<template>
  <v-card v-if="data != null" title="Messages des familles">
    <template #append>
      <v-btn-toggle
        color="primary"
        class="mx-1"
        v-model="sortBy"
        rounded
        mandatory
      >
        <v-btn
          icon="mdi-sort-clock-ascending-outline"
          title="Trier par date du dernier message"
        >
        </v-btn>
        <v-btn icon="mdi-sort-alphabetical-ascending" title="Trier par nom">
        </v-btn>
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

    <v-card-text>
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

const sortBy = ref<0 | 1>(0); // time / inscrit
const showOnlyNew = ref(false);

const filteredMessages = computed(() =>
  (data.value?.Messages || []).filter(
    (m) => !showOnlyNew.value || isMessageNew(m)
  )
);

function timeLastMessage(l: EventExt_MessageEvt[]) {
  let ti = new Date(l[0].Event.Created);
  for (const event of l) {
    const itemTime = new Date(event.Event.Created);
    if (itemTime.valueOf() > ti.valueOf()) {
      ti = itemTime;
    }
  }
  return ti;
}

const byDossiers = computed(() => {
  const tmp = new Map<IdDossier, EventExt_MessageEvt[]>();
  for (const element of filteredMessages.value) {
    tmp.set(
      element.Event.IdDossier,
      (tmp.get(element.Event.IdDossier) || []).concat(element)
    );
  }
  const out = Array.from(tmp.entries()).map((dossier) => ({
    IdDossier: dossier[0],
    Dossier: (data.value?.Dossiers || {})[dossier[0]],
    Messages: dossier[1].reverse(),
  }));
  if (sortBy.value == 0) {
    // by time, new comes first
    out.sort(
      (a, b) =>
        timeLastMessage(b.Messages).valueOf() -
        timeLastMessage(a.Messages).valueOf()
    );
  } else {
    out.sort((a, b) =>
      a.Dossier.Responsable.localeCompare(b.Dossier.Responsable)
    );
  }
  return out;
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
