<template>
  <v-app>
    <v-main>
      <v-app-bar rounded elevation="4" color="primary">
        <v-app-bar-title>
          <v-row>
            <v-col align-self="center" cols="auto">
              <v-img width="150" :src="logo" />
            </v-col>
            <v-col align-self="center"> Portail des inscriptions </v-col>
          </v-row>
        </v-app-bar-title>

        <template #append v-if="smAndUp">
          <v-btn :href="vitrineURL" class="mr-2">Site de l'association</v-btn>
        </template>
      </v-app-bar>

      <v-alert
        v-model="showPreinscription"
        closable
        title="Inscription rapide"
        class="ma-1"
      >
        <div v-if="mailFound === null">
          <v-row>
            <v-col cols="12" sm="6">
              Vous avez déjà participé à un de nos séjours ? Pré-remplissez ce
              formulaire en fournissant votre adresse e-mail !
            </v-col>
            <v-col align-self="center" cols="6" sm="3">
              <v-text-field
                variant="outlined"
                density="compact"
                label="Mail"
                hide-details
                v-model="search"
                @keydown.enter.prevent="
                  search.length < 3 ? null : searchHistory()
                "
              >
              </v-text-field>
            </v-col>
            <v-col align-self="center" cols="6" sm="auto">
              <v-btn
                variant="outlined"
                :disabled="search.length < 3"
                @click="searchHistory"
              >
                <template #prepend v-if="smAndUp">
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
          <b>lien sécurisé</b> vous y a été envoyé ({{ search }}). <br /><br />

          <div class="text-grey font-italic">
            Vous pouvez quitter cette page, vous y serez ramené par le lien.
          </div>
        </div>
        <div v-else-if="mailFound === false">
          Votre adresse n'a pas été trouvée. Aucun problème, vous pouvez suivre
          le parcours d'inscription standard ci-dessous !
        </div>
      </v-alert>

      <v-container style="min-height: 92%" v-if="isLoading">
        <v-skeleton-loader type="card"></v-skeleton-loader>
      </v-container>
      <v-container class="fill-height" v-else-if="!camps.length">
        <v-responsive>
          <v-alert class="text-center">
            Aucun camp n'est encore ouvert aux inscriptions.
          </v-alert>
        </v-responsive>
      </v-container>
      <v-container
        class="py-2"
        style="min-height: 92%"
        v-else-if="data != null"
      >
        <!-- confirmation après API call; hiding the form -->
        <v-card v-if="showInscriptionSaved != null" title="Confirmation">
          <v-card-text>
            Merci pour votre demande d'inscription ! <br />
            <br />
            Par mesure de sécurité, nous devons vérifier votre adresse mail. Un
            mail de confirmation a été envoyé à
            <b>{{ showInscriptionSaved.Responsable.Mail }}</b> : veuillez
            valider votre demande en suivant le lien que vous y trouverez.
            <br />
            <br />
            <div class="text-grey font-italic">
              Vous pouvez désormais quitter cette page.
            </div>
          </v-card-text>
        </v-card>

        <InscriptionPannel
          v-else
          :camps="camps"
          :preselected="preselected"
          :data="data"
          @done="(insc) => (showInscriptionSaved = insc)"
        ></InscriptionPannel>
      </v-container>
      <v-container class="fill-height" fluid v-else>
        <CampsList :camps="camps" @goTo="initWithCamp"></CampsList>
      </v-container>

      <v-footer color="primary">
        <v-row no-gutters class="my-1" justify="space-between">
          <v-col
            >{{ asso }} -
            <a href="/cgu" class="text-text">Mentions légales et CGU</a></v-col
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
import type { CampExt, Data, IdCamp, Inscription } from "./logic/api";
import { useDisplay } from "vuetify";
import CampsList from "./components/CampsList.vue";

const message = ref("");
const messageColor = ref("primary");

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
const vitrineURL = import.meta.env.VITE_ASSO_URL;
const version = `v${VITE_APP_VERSION}`;

onMounted(onLoad);

const { smAndUp } = useDisplay();

const preselected = ref<IdCamp>(0 as IdCamp);
const isLoading = ref(true);
async function onLoad() {
  // 3 cases :
  //    - preinscription : skip landing page and init with no preselection
  //    - preselection : skip landing page and init
  //    - nothing : just show landing page
  const query = new URLSearchParams(window.location.search);
  const preselectedS = query.get("preselected") || "";
  const preinscriptionS = query.get("preinscription") || "";

  // in any case, we need open camps
  await fetchCamps();

  // validate preselected
  const preselectedCamp = camps.value.find((c) => c.Slug == preselectedS);
  const preselectedId =
    preselectedCamp === undefined ? null : preselectedCamp.Id;

  if (preinscriptionS || preselectedId != null) {
    // init inscription
    await initInscription(preinscriptionS, preselectedId);
  }

  isLoading.value = false;
  preselected.value = preselectedId || (0 as IdCamp);
}

// camps ouverts aux inscriptions
const camps = ref<CampExt[]>([]);
async function fetchCamps() {
  const res = await controller.GetCamps();
  if (res === undefined) return;
  // ignore camp InscriptionExterne
  camps.value = (res || []).filter((c) => !c.InscriptionExterne);
}

async function initWithCamp(id: IdCamp) {
  preselected.value = id;
  initInscription("", id);
}

// inscription and settings
const data = ref<Data | null>(null);
async function initInscription(
  preinscription: string,
  preselected: IdCamp | null
) {
  const res = await controller.InitInscription({ preinscription });
  if (res === undefined) return;
  // vue reactivity does not work if Participants is null
  res.InitialInscription.Participants =
    res.InitialInscription.Participants || [];
  // apply preselected
  if (preselected !== null) {
    res.InitialInscription.Participants.forEach(
      (p) => (p.IdCamp = preselected)
    );
  }
  data.value = res;
  showPreinscription.value =
    res.Settings.ShowInscriptionRapide && preinscription == "";
}

// preinscription form
const showPreinscription = ref(false);
const search = ref("");
const mailFound = ref<boolean | null>(null);
async function searchHistory() {
  const res = await controller.SearchHistory({ mail: search.value });
  if (res === undefined) return;
  mailFound.value = res.MailFound;
}

const showInscriptionSaved = ref<Inscription | null>(null);
</script>
