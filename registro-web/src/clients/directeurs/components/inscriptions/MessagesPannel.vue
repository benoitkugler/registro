<template>
  <v-card v-if="data != null" title="Communication avec les familles">
    <template #append>
      <v-row>
        <v-col align-self="center">
          <v-btn icon size="small">
            <v-icon>mdi-filter</v-icon>
            <v-menu activator="parent" :close-on-content-click="false">
              <v-list density="compact">
                <v-radio-group hide-details v-model="sortBy">
                  <v-list-item
                    append-icon="mdi-sort-clock-ascending-outline"
                    title="Trier par date du dernier message"
                    :active="sortBy == 0"
                    color="info"
                    @click="sortBy = 0"
                  >
                    <template #prepend>
                      <v-list-item-action start>
                        <v-radio density="compact" :value="0"></v-radio>
                      </v-list-item-action>
                    </template>
                  </v-list-item>
                  <v-list-item
                    append-icon="mdi-sort-alphabetical-ascending"
                    title="Trier par nom du responsable"
                    :active="sortBy == 1"
                    color="info"
                    @click="sortBy = 1"
                  >
                    <template #prepend>
                      <v-list-item-action start>
                        <v-radio density="compact" :value="1"></v-radio>
                      </v-list-item-action>
                    </template>
                  </v-list-item>
                </v-radio-group>
                <v-divider></v-divider>
                <v-list-item
                  color="pink"
                  append-icon="mdi-message-badge"
                  title="Filtrer les messages non lu"
                  :active="showOnlyNew"
                  @click.stop="showOnlyNew = !showOnlyNew"
                >
                  <template #prepend>
                    <v-list-item-action start>
                      <v-checkbox-btn v-model="showOnlyNew"></v-checkbox-btn>
                    </v-list-item-action>
                  </template>
                </v-list-item>
              </v-list>
            </v-menu>
          </v-btn>
        </v-col>
        <v-divider thickness="1" vertical></v-divider>
        <v-col align-self="center">
          <v-btn
            prepend-icon="mdi-plus"
            color="success"
            @click="
              createMessageTo = null;
              showCreateMessage = true;
            "
            >Nouveau message</v-btn
          >
        </v-col>
      </v-row>
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
            @click="
              () => {
                createMessageTo = dossier.IdDossier;
                showCreateMessage = true;
              }
            "
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
          ></MessageRow>
        </v-card-text>
      </v-card>
    </v-card-text>

    <!-- new message -->
    <v-dialog v-model="showCreateMessage" max-width="600px">
      <NewMessageCard
        :dossiers="data.Dossiers"
        :initial-destinataire="createMessageTo"
        @send="sendMessage"
      ></NewMessageCard>
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
import NewMessageCard from "./NewMessageCard.vue";

const props = defineProps<{}>();

onMounted(loadMessages);

const sortBy = ref<0 | 1>(0); // time / inscrit
const showOnlyNew = ref(false);

const filteredMessages = computed(() =>
  (data.value?.Messages || []).filter(
    (m) => !showOnlyNew.value || isMessageNew(m),
  ),
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
      (tmp.get(element.Event.IdDossier) || []).concat(element),
    );
  }
  const out = Array.from(tmp.entries()).map((dossier) => ({
    IdDossier: dossier[0],
    Dossier: (data.value?.Dossiers || {})[dossier[0]],
    Messages: dossier[1],
  }));
  if (sortBy.value == 0) {
    // by time, new comes first
    out.sort(
      (a, b) =>
        timeLastMessage(b.Messages).valueOf() -
        timeLastMessage(a.Messages).valueOf(),
    );
  } else {
    out.sort((a, b) =>
      a.Dossier.Responsable.localeCompare(b.Dossier.Responsable),
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
    seen ? "Message marqué comme lu." : "Message marqué comme non lu.",
  );
}

const showCreateMessage = ref(false);
const createMessageTo = ref<IdDossier | null>(null);

// async function sendMessage(destinataire: IdDossier, message: string) {
//   const idDossier = createMessageTo.value;
//   if (!data.value || !idDossier) return;
//   createMessageTo.value = null;
//   const res = await controller.ParticipantsMessagesCreate({
//     Contenu: newMessage.value,
//     IdDossier: idDossier,
//   });
//   if (res === undefined) return;
//   controller.showMessage("Message envoyé avec succès.");
//   data.value.Messages = [res].concat(data.value.Messages || []);
// }

async function sendMessage(destinataire: IdDossier, message: string) {
  showCreateMessage.value = false;
  const res = await controller.ParticipantsMessagesCreate({
    Contenu: message,
    IdDossier: destinataire,
  });
  if (res === undefined) return;
  controller.showMessage("Message envoyé avec succès.");
  if (!data.value) return;
  data.value.Messages = [res].concat(data.value.Messages || []);
}
</script>
