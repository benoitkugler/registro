<template>
  <NavBar :title="`${controller.camp?.Label} - Projet spirituel`"> </NavBar>

  <div v-if="data == null" class="text-center my-6">
    <v-progress-circular indeterminate></v-progress-circular>
  </div>
  <div class="ma-2" v-else>
    <v-card title="Projet spirituel du camp">
      <v-card-text>
        Ce formulaire est complété par les équipes de direction et reste interne
        à l'association. <br />
        <br />
        Il a pour but d'être un soutien aux directeurs de camps dans
        l'élaboration du projet spirituel du camp. <br />
        <br />
        Nous sommes à disposition pour répondre ou réfléchir avec vous pour
        toute question / interrogation, ainsi que pour prodiguer des conseils ou
        chercher ensemble des solutions sur les différents aspects de
        l'organisation du camp (spirituel, équipe, activités, logistique...).
      </v-card-text>
    </v-card>

    <v-card class="my-2">
      <v-card-text>
        <v-row
          ><v-col>
            Décrivez en quelques lignes le projet spirituel que vous portez pour
            votre camp (objectifs, thèmes, orateurs...)
          </v-col></v-row
        >
        <v-row
          ><v-col>
            <v-textarea
              variant="outlined"
              density="comfortable"
              label="Description"
              v-model="data.Description"
              rows="2"
              hide-details
            ></v-textarea> </v-col
        ></v-row>
      </v-card-text>
    </v-card>

    <v-card class="my-2">
      <v-card-text>
        <v-row
          ><v-col>
            Quel est le programme général de l'animation et des activités ? En
            quoi ce programme oriente-t-il vers les objectifs spirituels du camp
            ?
          </v-col></v-row
        >
        <v-row
          ><v-col>
            <v-textarea
              variant="outlined"
              density="comfortable"
              label="Programme"
              v-model="data.Programme"
              rows="2"
              hide-details
            ></v-textarea> </v-col
        ></v-row>
      </v-card-text>
    </v-card>

    <v-card class="my-2">
      <v-card-text>
        <v-row
          ><v-col>
            Dans une journée standard, quels sont les moments spirituels
            (lecture, prière, partage...) que vous pensez inclure dans le
            programme ?
          </v-col></v-row
        >
        <v-row
          ><v-col>
            <v-textarea
              variant="outlined"
              density="comfortable"
              label="Journée type"
              v-model="data.JourneeType"
              rows="2"
              hide-details
            ></v-textarea> </v-col
        ></v-row>
      </v-card-text>
    </v-card>

    <v-card class="my-2">
      <v-card-text>
        <v-row
          ><v-col>
            Comment pensez-vous encourager les campeurs à se mettre dans une
            dynamique de prière, partage et lecture de la Parole ?
          </v-col></v-row
        >
        <v-row
          ><v-col>
            <v-textarea
              variant="outlined"
              density="comfortable"
              label="Dynamique des campeurs"
              v-model="data.DynamiqueCampeur"
              rows="2"
              hide-details
            ></v-textarea> </v-col
        ></v-row>
      </v-card-text>
    </v-card>

    <v-card class="my-2">
      <v-card-text>
        <v-row
          ><v-col>
            Quand et comment prévoyez-vous de présenter le message de l'Évangile
            ?
          </v-col></v-row
        >
        <v-row
          ><v-col>
            <v-textarea
              variant="outlined"
              density="comfortable"
              label="Message de l'Évangile"
              v-model="data.Evangile"
              rows="2"
              hide-details
            ></v-textarea> </v-col
        ></v-row>
      </v-card-text>
    </v-card>

    <v-card class="my-2">
      <v-card-text>
        <v-row
          ><v-col>
            Présentez l'équipe pressentie pour vous accompagner dans ce projet.
            Comment prévoyez-vous que le projet spirituel soit porté par
            l'ensemble de l'équipe ? Quelles ressources collectives et
            individuelles dans l'équipe pourront être mobilisées pour appuyer ce
            projet (talents, dons, maturité..) ? Il y a-t-il encore des besoins
            en termes de ressources ?
          </v-col></v-row
        >
        <v-row
          ><v-col>
            <v-textarea
              variant="outlined"
              density="comfortable"
              label="Équipe"
              v-model="data.Equipe"
              rows="2"
              hide-details
            ></v-textarea> </v-col
        ></v-row>
      </v-card-text>
    </v-card>

    <v-card class="my-2">
      <v-card-text>
        <v-row
          ><v-col>
            Comment prévoyez-vous d'intégrer les équipes cuisine dans le projet
            spirituel (activité ou partage à préparer, effort pour qu'elles
            puissent au moins entendre une partie des messages...?)
          </v-col></v-row
        >
        <v-row
          ><v-col>
            <v-textarea
              variant="outlined"
              density="comfortable"
              label="Intégration de l'équipe cuisine"
              v-model="data.Cuisine"
              rows="2"
              hide-details
            ></v-textarea> </v-col
        ></v-row>
      </v-card-text>
    </v-card>

    <v-card class="my-2">
      <v-card-text>
        <v-row
          ><v-col>
            Comment imaginez-vous le travail de suite avec les campeurs après le
            camp ?
          </v-col></v-row
        >
        <v-row
          ><v-col>
            <v-textarea
              variant="outlined"
              density="comfortable"
              label="Suite du camp"
              v-model="data.Suite"
              rows="2"
              hide-details
            ></v-textarea> </v-col
        ></v-row>
      </v-card-text>
    </v-card>

    <v-card class="my-2">
      <v-card-text>
        <v-row
          ><v-col>
            La librairie "Au jardin du livre" se met à disposition des camps
            pour les visiter ; souhaitez-vous une visite de leur part ?
          </v-col></v-row
        >
        <v-row
          ><v-col>
            <v-radio-group
              variant="outlined"
              density="comfortable"
              v-model="data.VisiteLibrairie"
              hide-details
            >
              <v-radio
                v-for="item in selectItems(VisiteLibrairieLabels)"
                :value="item.value"
                :label="item.title"
              ></v-radio>
            </v-radio-group> </v-col
        ></v-row>
      </v-card-text>
    </v-card>

    <v-card class="my-2">
      <v-card-text>
        <v-row
          ><v-col>
            Avez-vous prévu des Bibles adaptées à l'âge des participants à
            donner aux campeurs qui n'en auraient pas ?
          </v-col></v-row
        >
        <v-row
          ><v-col>
            <v-radio-group
              variant="outlined"
              density="comfortable"
              v-model="data.VisiteLibrairie"
              hide-details
            >
              <v-radio
                v-for="item in selectItems(VisiteLibrairieLabels)"
                :value="item.value"
                :label="item.title"
              ></v-radio>
            </v-radio-group> </v-col
        ></v-row>
      </v-card-text>
    </v-card>
  </div>
</template>

<script lang="ts" setup>
import { onMounted, ref } from "vue";
import NavBar from "../components/NavBar.vue";
import { controller } from "../logic/logic";
import { VisiteLibrairieLabels, type ProjetSpi } from "../logic/api";
import { selectItems } from "@/utils";

onMounted(loadData);

const data = ref<ProjetSpi | null>(null);
async function loadData() {
  const res = await controller.ProjetSpiLoad();
  if (res === undefined) return;

  data.value = res;
}
</script>
