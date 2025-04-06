<template>
  <v-app>
    <v-main>
      <v-app-bar rounded elevation="4" color="primary">
        <v-app-bar-title>
          <v-row>
            <v-col align-self="center" cols="auto">
              <v-img width="60" :src="logo" />
            </v-col>
            <v-col align-self="center">
              <template v-if="smAndUp">
                Bienvenue sur le Portail des inscriptions
              </template>
              <template v-else> Portail des inscriptions </template>
            </v-col>
          </v-row>
        </v-app-bar-title>
      </v-app-bar>

      <v-dialog v-model="showPreinscription" max-width="600px">
        <v-card title="Inscription rapide">
          <v-card-text>
            <div v-if="mailFound === null">
              Vous avez déjà participé à un de nos séjours ? Pré-remplissez ce
              formulaire en fournissant votre adresse e-mail !
              <v-row justify="space-between" class="my-2">
                <v-col align-self="center">
                  <v-text-field
                    variant="outlined"
                    density="comfortable"
                    label="Mail"
                    hide-details
                    v-model="search"
                  >
                  </v-text-field>
                </v-col>
                <v-col align-self="center" cols="auto">
                  <v-btn
                    color="secondary"
                    :disabled="search.length < 3"
                    @click="searchHistory"
                  >
                    <template #prepend>
                      <v-icon>mdi-magnify</v-icon>
                    </template>
                    Rechercher</v-btn
                  >
                </v-col>
              </v-row>
            </div>
            <div v-else-if="mailFound === true">
              Votre adresse mail a bien été utilisée sur un de nos séjours.
              <br />
              Afin de protéger vos données personnelles, un
              <b>lien sécurisé</b> vous y a été envoyé ({{ search }}).
              <br /><br />

              <div class="text-grey font-italic">
                Vous pouvez quitter cette page, vous y serez ramené par le lien.
              </div>
            </div>
            <div v-else-if="mailFound === false">
              Votre adresse n'a pas été trouvée. Aucun problème, vous pouvez
              reprendre l'inscription standard !
            </div>
          </v-card-text>
          <v-card-actions v-if="mailFound === null">
            <v-btn @click="showPreinscription = false">Ignorer</v-btn>
          </v-card-actions>
        </v-card>
      </v-dialog>

      <v-container style="min-height: 92%" v-if="data == null">
        <v-skeleton-loader type="card"></v-skeleton-loader>
      </v-container>
      <v-container class="fill-height" v-else-if="!data.Camps?.length">
        <v-responsive>
          <v-alert class="text-center">
            Aucun camp n'est encore ouvert aux inscriptions.
          </v-alert>
        </v-responsive>
      </v-container>
      <v-container class="py-2" style="min-height: 92%" v-else>
        <InscriptionPannel :data="data"></InscriptionPannel>
      </v-container>

      <v-footer color="secondary">
        <v-row no-gutters class="my-1" justify="space-between">
          <v-col
            >{{ asso }} -
            <a href="/cgu" class="text-black">Mentions légales et CGU</a></v-col
          >
          <v-col class="text-right">{{ version }}</v-col>
        </v-row>
      </v-footer>

      <v-snackbar
        style="z-index: 10000"
        app
        :model-value="message != ''"
        @update:model-value="message = ''"
        :timeout="4000"
        :color="messageColor"
        location="bottom left"
        close-on-content-click
      >
        {{ message }}
      </v-snackbar>

      <v-snackbar
        app
        :model-value="errorKind != ''"
        @update:model-value="errorKind = ''"
        :timeout="4000"
        color="red"
      >
        <b>{{ errorKind }}</b>
        <div v-html="errorHtml"></div>
      </v-snackbar>
    </v-main>
  </v-app>
</template>

<script lang="ts" setup>
import { onMounted, ref } from "vue";
import { controller } from "./logic/logic";
import InscriptionPannel from "./components/InscriptionPannel.vue";
import type { Data, IdCamp } from "./logic/api";
import { useDisplay } from "vuetify";

const message = ref("");
const messageColor = ref("secondary");

const errorKind = ref("");
const errorHtml = ref("");

controller.onError = (s, m) => {
  errorKind.value = s;
  errorHtml.value = m;
};

controller.showMessage = (s, color) => {
  message.value = s;
  messageColor.value = color || "success";
};

const logo = `${import.meta.env.BASE_URL}${import.meta.env.VITE_ASSO}/logo.png`;
const asso = import.meta.env.VITE_ASSO_TITLE;
const version = `v${VITE_APP_VERSION}`;

const data = ref<Data | null>(null);

onMounted(fetchData);

const { smAndUp } = useDisplay();

async function fetchData() {
  // forward url params
  const query = new URLSearchParams(window.location.search);
  const preselected = query.get("preselected") || "";
  const preinscription = query.get("preinscription") || "";
  const res = await controller.LoadData({
    preselected: (preselected ? Number(preselected) : 0) as IdCamp,
    preinscription,
  });
  if (res === undefined) return;
  // vue reactivity does not work if Participants is null
  res.InitialInscription.Participants =
    res.InitialInscription.Participants || [];
  data.value = res;
}

// preinscription form
const showPreinscription = ref(true);
const search = ref("");
const mailFound = ref<boolean | null>(null);
async function searchHistory() {
  const res = await controller.SearchHistory({ mail: search.value });
  if (res === undefined) return;
  mailFound.value = res.MailFound;
}
</script>
