<template>
  <NavBar title="Bienvenue sur votre espace de suivi"> </NavBar>

  <v-skeleton-loader type="card" v-if="data == null"></v-skeleton-loader>
  <v-container class="fill-height" fluid v-else>
    <v-row>
      <v-col> </v-col>
      <v-col cols="7">
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
            <div class="overflow-y-auto" style="height: 75vh">
              <v-timeline side="end" class="mt-4" density="compact">
                <EventSwitch v-for="event in events" :event="event">
                </EventSwitch>
              </v-timeline>
            </div>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
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
</template>

<script lang="ts" setup>
import { useRouter } from "vue-router";
import { computed, onMounted, ref } from "vue";
import NavBar from "../components/NavBar.vue";
import { controller } from "../logic/logic";
import type { Data } from "../logic/api";
import { buildPseudoEvents } from "@/utils";

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
</script>
