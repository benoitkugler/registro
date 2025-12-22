<template>
  <v-card
    title="Vue d'ensembe"
    :subtitle="`Moyenne des (${props.sondages.Sondages.length}) notes atttribuées (de 1 à 4)`"
    class="mb-2"
    v-if="props.sondages.Sondages?.length"
  >
    <v-card-text>
      <v-row>
        <v-col cols="12" sm="6" md="4">
          <RatingBadge
            readonly
            label="Informations avant le séjour"
            :value="props.sondages.Moyennes.InfosAvantSejour"
          ></RatingBadge
        ></v-col>
        <v-col cols="12" sm="6" md="4">
          <RatingBadge
            readonly
            label="Informations pendant le séjour"
            :value="props.sondages.Moyennes.InfosPendantSejour"
          ></RatingBadge
        ></v-col>
        <v-col cols="12" sm="6" md="4">
          <RatingBadge
            readonly
            label="L'hébergement"
            :value="props.sondages.Moyennes.Hebergement"
          ></RatingBadge
        ></v-col>
      </v-row>

      <v-row>
        <v-col cols="12" sm="6" md="4">
          <RatingBadge
            readonly
            label="Les activités"
            :value="props.sondages.Moyennes.Activites"
          ></RatingBadge
        ></v-col>
        <v-col cols="12" sm="6" md="4">
          <RatingBadge
            readonly
            label="Le thème"
            :value="props.sondages.Moyennes.Theme"
          ></RatingBadge
        ></v-col>
        <v-col cols="12" sm="6" md="4">
          <RatingBadge
            readonly
            label="La nourriture"
            :value="props.sondages.Moyennes.Nourriture"
          ></RatingBadge
        ></v-col>
      </v-row>
      <v-row>
        <v-col cols="12" sm="6" md="4">
          <RatingBadge
            readonly
            label="L'hygiène corporelle et vestimentaire"
            :value="props.sondages.Moyennes.Hygiene"
          ></RatingBadge
        ></v-col>
        <v-col cols="12" sm="6" md="4">
          <RatingBadge
            readonly
            label="L'ambiance du groupe"
            :value="props.sondages.Moyennes.Ambiance"
          ></RatingBadge
        ></v-col>
        <v-col cols="12" sm="6" md="4">
          <RatingBadge
            readonly
            label="Le ressenti global de votre enfant"
            :value="props.sondages.Moyennes.Ressenti"
          ></RatingBadge
        ></v-col>
      </v-row>
    </v-card-text>
  </v-card>
  <v-container v-else>
    <v-alert type="info"> Aucun avis n'a encore été rempli. </v-alert>
  </v-container>

  <v-card
    v-for="(sondage, index) in props.sondages.Sondages"
    :title="`${sondage.ResponsableNom} - ${sondage.ResponsableMail}`"
    :subtitle="(sondage.Participants || []).join(', ')"
    class="my-2"
    :key="index"
  >
    <template #append>
      {{ Formatters.time(sondage.Sondage.Modified) }}
    </template>
    <v-card-text>
      <v-row>
        <v-col cols="7">
          <v-row>
            <v-col cols="12" sm="6" md="4">
              <RatingBadge
                label="Informations avant le séjour"
                :value="sondage.Sondage.InfosAvantSejour"
              ></RatingBadge
            ></v-col>
            <v-col cols="12" sm="6" md="4">
              <RatingBadge
                label="Informations pendant le séjour"
                :value="sondage.Sondage.InfosPendantSejour"
              ></RatingBadge
            ></v-col>
            <v-col cols="12" sm="6" md="4">
              <RatingBadge
                label="L'hébergement"
                :value="sondage.Sondage.Hebergement"
              ></RatingBadge
            ></v-col>
          </v-row>

          <v-row>
            <v-col cols="12" sm="6" md="4">
              <RatingBadge
                label="Les activités"
                :value="sondage.Sondage.Activites"
              ></RatingBadge
            ></v-col>
            <v-col cols="12" sm="6" md="4">
              <RatingBadge
                label="Le thème"
                :value="sondage.Sondage.Theme"
              ></RatingBadge
            ></v-col>
            <v-col cols="12" sm="6" md="4">
              <RatingBadge
                label="La nourriture"
                :value="sondage.Sondage.Nourriture"
              ></RatingBadge
            ></v-col>
          </v-row>
          <v-row>
            <v-col cols="12" sm="6" md="4">
              <RatingBadge
                label="L'hygiène corporelle et vestimentaire"
                :value="sondage.Sondage.Hygiene"
              ></RatingBadge
            ></v-col>
            <v-col cols="12" sm="6" md="4">
              <RatingBadge
                label="L'ambiance du groupe"
                :value="sondage.Sondage.Ambiance"
              ></RatingBadge
            ></v-col>
            <v-col cols="12" sm="6" md="4">
              <RatingBadge
                label="Le ressenti global de votre enfant"
                :value="sondage.Sondage.Ressenti"
              ></RatingBadge
            ></v-col>
          </v-row>
        </v-col>
        <v-col align-self="center">
          <v-row>
            <v-col cols="12">
              <v-textarea
                density="compact"
                label="Participant"
                variant="outlined"
                hide-details
                :value="sondage.Sondage.MessageEnfant"
                readonly
                rows="3"
              ></v-textarea>
            </v-col>
            <v-col cols="12">
              <v-textarea
                density="compact"
                label="Responsable"
                variant="outlined"
                hide-details
                :value="sondage.Sondage.MessageResponsable"
                readonly
                rows="3"
              ></v-textarea>
            </v-col>
          </v-row>
        </v-col>
      </v-row>
    </v-card-text>
  </v-card>
</template>

<script lang="ts" setup>
import type { CampSondages } from "@/clients/directeurs/logic/api";
import { Formatters } from "@/utils";
import RatingBadge from "./RatingBadge.vue";

const props = defineProps<{
  sondages: CampSondages;
}>();
</script>
