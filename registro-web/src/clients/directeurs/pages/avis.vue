<template>
  <NavBar :title="`${controller.camp?.Label} - Avis sur le séjour`"> </NavBar>

  <div v-if="data == null" class="text-center my-6">
    <v-progress-circular indeterminate></v-progress-circular>
  </div>
  <template v-else>
    <v-card
      title="Vue d'ensembe"
      :subtitle="`Moyenne des notes atttribuées (${data.Sondages.length})`"
      class="ma-2"
      v-if="data.Sondages?.length"
    >
      <v-card-text>
        <v-row>
          <v-col cols="12" sm="6" md="4">
            <RatingField
              readonly
              label="Informations avant le séjour"
              :model-value="data.Moyennes.InfosAvantSejour"
            ></RatingField
          ></v-col>
          <v-col cols="12" sm="6" md="4">
            <RatingField
              readonly
              label="Informations pendant le séjour"
              :model-value="data.Moyennes.InfosPendantSejour"
            ></RatingField
          ></v-col>
          <v-col cols="12" sm="6" md="4">
            <RatingField
              readonly
              label="L'hébergement"
              :model-value="data.Moyennes.Hebergement"
            ></RatingField
          ></v-col>
        </v-row>

        <v-row>
          <v-col cols="12" sm="6" md="4">
            <RatingField
              readonly
              label="Les activités"
              :model-value="data.Moyennes.Activites"
            ></RatingField
          ></v-col>
          <v-col cols="12" sm="6" md="4">
            <RatingField
              readonly
              label="Le thème"
              :model-value="data.Moyennes.Theme"
            ></RatingField
          ></v-col>
          <v-col cols="12" sm="6" md="4">
            <RatingField
              readonly
              label="La nourriture"
              :model-value="data.Moyennes.Nourriture"
            ></RatingField
          ></v-col>
        </v-row>
        <v-row>
          <v-col cols="12" sm="6" md="4">
            <RatingField
              readonly
              label="L'hygiène corporelle et vestimentaire"
              :model-value="data.Moyennes.Hygiene"
            ></RatingField
          ></v-col>
          <v-col cols="12" sm="6" md="4">
            <RatingField
              readonly
              label="L'ambiance du groupe"
              :model-value="data.Moyennes.Ambiance"
            ></RatingField
          ></v-col>
          <v-col cols="12" sm="6" md="4">
            <RatingField
              readonly
              label="Le ressenti global de votre enfant"
              :model-value="data.Moyennes.Ressenti"
            ></RatingField
          ></v-col>
        </v-row>
      </v-card-text>
    </v-card>
    <v-container v-else>
      <v-alert type="info"> Aucun avis n'a encore été rempli. </v-alert>
    </v-container>

    <v-card
      v-for="sondage in data.Sondages"
      :title="`${sondage.ResponsableNom} - ${sondage.ResponsableMail}`"
      :subtitle="(sondage.Participants || []).join(', ')"
      class="ma-2"
    >
      <template #append>
        {{ Formatters.time(sondage.Sondage.Modified) }}
      </template>
      <v-card-text>
        <v-row>
          <v-col cols="7">
            <v-row>
              <v-col cols="12" sm="6" md="4">
                <RatingField
                  readonly
                  label="Informations avant le séjour"
                  :model-value="sondage.Sondage.InfosAvantSejour"
                ></RatingField
              ></v-col>
              <v-col cols="12" sm="6" md="4">
                <RatingField
                  readonly
                  label="Informations pendant le séjour"
                  :model-value="sondage.Sondage.InfosPendantSejour"
                ></RatingField
              ></v-col>
              <v-col cols="12" sm="6" md="4">
                <RatingField
                  readonly
                  label="L'hébergement"
                  :model-value="sondage.Sondage.Hebergement"
                ></RatingField
              ></v-col>
            </v-row>

            <v-row>
              <v-col cols="12" sm="6" md="4">
                <RatingField
                  readonly
                  label="Les activités"
                  :model-value="sondage.Sondage.Activites"
                ></RatingField
              ></v-col>
              <v-col cols="12" sm="6" md="4">
                <RatingField
                  readonly
                  label="Le thème"
                  :model-value="sondage.Sondage.Theme"
                ></RatingField
              ></v-col>
              <v-col cols="12" sm="6" md="4">
                <RatingField
                  readonly
                  label="La nourriture"
                  :model-value="sondage.Sondage.Nourriture"
                ></RatingField
              ></v-col>
            </v-row>
            <v-row>
              <v-col cols="12" sm="6" md="4">
                <RatingField
                  readonly
                  label="L'hygiène corporelle et vestimentaire"
                  :model-value="sondage.Sondage.Hygiene"
                ></RatingField
              ></v-col>
              <v-col cols="12" sm="6" md="4">
                <RatingField
                  readonly
                  label="L'ambiance du groupe"
                  :model-value="sondage.Sondage.Ambiance"
                ></RatingField
              ></v-col>
              <v-col cols="12" sm="6" md="4">
                <RatingField
                  readonly
                  label="Le ressenti global de votre enfant"
                  :model-value="sondage.Sondage.Ressenti"
                ></RatingField
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
                  :model-value="sondage.Sondage.MessageEnfant"
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
                  :model-value="sondage.Sondage.MessageResponsable"
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
</template>

<script lang="ts" setup>
import { onMounted, ref } from "vue";
import NavBar from "../components/NavBar.vue";
import { controller } from "../logic/logic";
import type { SondagesOut } from "../logic/api";
import { Formatters } from "@/utils";

onMounted(loadData);

const data = ref<SondagesOut | null>(null);
async function loadData() {
  const res = await controller.SondagesGet();
  if (res === undefined) return;

  data.value = res;
}
</script>
