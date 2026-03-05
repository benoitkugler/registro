<template>
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
            color="primary"
            v-model="data.Description"
            :readonly="props.readonly"
            rows="2"
            hide-details
            :rules="[FormRules.required('Champ requis')]"
          ></v-textarea> </v-col
      ></v-row>
    </v-card-text>
  </v-card>

  <v-card class="my-2">
    <v-card-text>
      <v-row
        ><v-col>
          Quel est le programme général de l'animation et des activités ? En
          quoi ce programme oriente-t-il vers les objectifs spirituels du camp ?
        </v-col></v-row
      >
      <v-row
        ><v-col>
          <v-textarea
            variant="outlined"
            density="comfortable"
            label="Programme"
            color="primary"
            v-model="data.Programme"
            :readonly="props.readonly"
            rows="2"
            hide-details
            :rules="[FormRules.required('Champ requis')]"
          ></v-textarea> </v-col
      ></v-row>
    </v-card-text>
  </v-card>

  <v-card class="my-2">
    <v-card-text>
      <v-row
        ><v-col>
          Dans une journée standard, quels sont les moments spirituels (lecture,
          prière, partage...) que vous pensez inclure dans le programme ?
        </v-col></v-row
      >
      <v-row
        ><v-col>
          <v-textarea
            variant="outlined"
            density="comfortable"
            label="Journée type"
            color="primary"
            v-model="data.JourneeType"
            :readonly="props.readonly"
            rows="2"
            hide-details
            :rules="[FormRules.required('Champ requis')]"
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
            color="primary"
            v-model="data.DynamiqueCampeur"
            :readonly="props.readonly"
            rows="2"
            hide-details
            :rules="[FormRules.required('Champ requis')]"
          ></v-textarea> </v-col
      ></v-row>
    </v-card-text>
  </v-card>

  <v-card class="my-2">
    <v-card-text>
      <v-row
        ><v-col>
          Quand et comment prévoyez-vous de présenter le message de l'Évangile ?
        </v-col></v-row
      >
      <v-row
        ><v-col>
          <v-textarea
            variant="outlined"
            density="comfortable"
            label="Message de l'Évangile"
            color="primary"
            v-model="data.Evangile"
            :readonly="props.readonly"
            rows="2"
            hide-details
            :rules="[FormRules.required('Champ requis')]"
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
          projet (talents, dons, maturité..) ? Y a-t-il encore des besoins en
          termes de ressources ?
        </v-col></v-row
      >
      <v-row
        ><v-col>
          <v-textarea
            variant="outlined"
            density="comfortable"
            label="Équipe"
            color="primary"
            v-model="data.Equipe"
            :readonly="props.readonly"
            rows="2"
            hide-details
            :rules="[FormRules.required('Champ requis')]"
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
            color="primary"
            v-model="data.Cuisine"
            :readonly="props.readonly"
            rows="2"
            hide-details
            :rules="[FormRules.required('Champ requis')]"
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
            color="primary"
            v-model="data.Suite"
            :readonly="props.readonly"
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
          La librairie "Au jardin du livre" se met à disposition des camps pour
          les visiter ; souhaitez-vous une visite de leur part ?
        </v-col></v-row
      >
      <v-row
        ><v-col>
          <v-radio-group
            variant="outlined"
            density="comfortable"
            color="primary"
            v-model="data.VisiteLibrairie"
            :readonly="props.readonly"
            hide-details
            :rules="[FormRules.required('Champ requis')]"
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
          Avez-vous prévu des Bibles adaptées à l'âge des participants à donner
          aux campeurs qui n'en auraient pas ?
        </v-col></v-row
      >
      <v-row
        ><v-col>
          <v-radio-group
            variant="outlined"
            density="comfortable"
            color="primary"
            v-model="data.Bibles"
            :readonly="props.readonly"
            hide-details
            :rules="[FormRules.required('Champ requis')]"
          >
            <v-radio :value="true" label="Oui"></v-radio>
            <v-radio :value="false" label="Non"></v-radio>
          </v-radio-group> </v-col
      ></v-row>
    </v-card-text>
  </v-card>

  <v-card class="my-2">
    <v-card-text>
      <v-row
        ><v-col>
          Auriez-vous des questions ou des besoins spécifiques ? Sur quels
          points pourrions-nous vous être utile ?
        </v-col></v-row
      >
      <v-row
        ><v-col>
          <v-textarea
            variant="outlined"
            density="comfortable"
            label="Question"
            color="primary"
            v-model="data.Question"
            :readonly="props.readonly"
            rows="2"
            hide-details
          ></v-textarea> </v-col
      ></v-row>
    </v-card-text>
  </v-card>
</template>

<script lang="ts" setup>
import {
  VisiteLibrairieLabels,
  type ProjetSpi,
} from "@/clients/backoffice/logic/api";
import { FormRules, selectItems } from "@/utils";

const props = defineProps<{ readonly: boolean }>();

const data = defineModel<ProjetSpi>({ required: true });
</script>
