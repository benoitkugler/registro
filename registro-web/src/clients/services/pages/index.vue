<template>
  <NavBar :title="`Portail de gestion - Association ${asso}`"> </NavBar>

  <v-container class="fill-height flex-column justify-space-evenly">
    <v-card class="border-primary border-lg" title="Inscriptions" width="100%">
      <v-card-text>
        <v-row>
          <v-col align-self="center">
            Vous cherchez plus d'informations sur notre association ? C'est par
            ici !
          </v-col>
          <v-col align-self="center" cols="auto">
            <v-btn color="accent" :href="vitrineURL">Site {{ asso }}</v-btn>
          </v-col>
        </v-row>
        <v-row>
          <v-col align-self="center">
            Vous souhaitez vous <b>inscrire</b> à un camp {{ asso }} ? C'est par
            là !
          </v-col>
          <v-col align-self="center" cols="auto">
            <v-btn color="accent" href="/inscription">Inscription</v-btn>
          </v-col>
        </v-row>
        <v-row>
          <v-col align-self="center">
            Vous souhaitez accéder à votre <b>espace de suivi</b> ? Recevez à
            nouveau votre lien de connexion par email.
          </v-col>
          <v-col align-self="center" cols="3">
            <v-text-field
              label="Email"
              hide-details
              density="compact"
              variant="outlined"
              v-model="mail"
              color="primary"
            ></v-text-field>
          </v-col>
          <v-col align-self="center" cols="auto">
            <v-btn
              color="accent"
              :disabled="mail.length < 3"
              @click="searchMail"
              >Retrouver mon lien</v-btn
            >
          </v-col>
        </v-row>
      </v-card-text>
    </v-card>

    <v-card class="border-primary border-lg" title="Directeurs" width="100%">
      <v-card-text>
        <v-row>
          <v-col align-self="center">
            Vous êtes un <b>directeur de séjour</b> ? C'est par ici !
          </v-col>
          <v-col align-self="center" cols="auto">
            <v-btn color="accent" href="/directeurs"
              >Portail des directeurs</v-btn
            >
          </v-col>
        </v-row>
      </v-card-text>
    </v-card>

    <v-dialog
      :model-value="searchOut != null"
      @update:model-value="searchOut = null"
      max-width="600px"
    >
      <v-card title="Recherche de votre espace de suivi" v-if="searchOut">
        <v-card-text>
          <div v-if="searchOut.Found == 0">
            Navré, mais votre adresse mail ne fait pas partie des inscriptions
            enregistrées dans notre base de données.
          </div>
          <div v-else>
            Bonne nouvelle, votre adresse <b>{{ mail }}</b> est bien présente
            dans notre base de données ! <br /><br />
            Un mail contenant le lien de connexion vous y a été envoyé.
          </div>
        </v-card-text>
      </v-card>
    </v-dialog>
  </v-container>
</template>

<script lang="ts" setup>
import { ref } from "vue";
import NavBar from "../components/NavBar.vue";
import { controller } from "../logic/logic";
import type { SearchMailOut } from "../logic/api";

const asso = import.meta.env.VITE_ASSO_TITLE;
const vitrineURL = import.meta.env.VITE_ASSO_URL;

const mail = ref("");
const searchOut = ref<SearchMailOut | null>(null);
async function searchMail() {
  const res = await controller.SearchMail({ mail: mail.value });
  if (res === undefined) return;
  searchOut.value = res;
}
</script>
