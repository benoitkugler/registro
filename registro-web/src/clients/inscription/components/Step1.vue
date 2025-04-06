<template>
  <v-card
    title="Profil du responsable légal"
    :subtitle="
      smAndUp
        ? `Personne majeure qui sera le contact pour le suivi de l'inscription.`
        : `Contact pour le suivi de l'inscription.`
    "
  >
    <v-card-text>
      <v-form class="mt-2">
        <v-row>
          <v-col md="3" sm="6" cols="12">
            <v-text-field
              variant="outlined"
              density="compact"
              v-model="respo.Nom"
              label="Nom"
              :rules="[FormRules.required('Merci de remplir votre nom.')]"
            ></v-text-field>
          </v-col>
          <v-col md="3" sm="6" cols="12">
            <v-text-field
              variant="outlined"
              density="compact"
              v-model="respo.Prenom"
              label="Prénom"
              :rules="[FormRules.required('Merci de remplir votre prénom.')]"
            ></v-text-field>
          </v-col>
          <v-col md="3" sm="4" cols="12">
            <SexeField
              v-model="respo.Sexe"
              :rules="[
                FormRules.required(
                  'Nous avons besoin de votre sexe pour personnaliser nos courriels.'
                ),
              ]"
            ></SexeField>
          </v-col>
          <v-col md="3" sm="8" cols="12">
            <DateNaissanceField
              v-model="respo.DateNaissance"
              :rule="checkDateNaissance"
            ></DateNaissanceField>
          </v-col>
        </v-row>

        <v-row>
          <v-col md="6" cols="12">
            <v-text-field
              variant="outlined"
              density="compact"
              v-model="respo.Mail"
              label="Adresse mail"
              type="email"
              :rules="[
                FormRules.required(
                  'Une adresse mail est nécessaire pour recevoir les informations sur le suivi de votre inscription.'
                ),
              ]"
            ></v-text-field>
          </v-col>
          <v-col md="6" cols="12">
            <StringList
              v-model="respo.Tels"
              :formatter="Formatters.tel"
              label="Téléphones"
              :rule="
                FormRules.noEmptyList(
                  `Merci de fournir un numéro en cas d'urgence.`
                )
              "
            ></StringList>
          </v-col>
        </v-row>

        <v-row>
          <v-col md="4" sm="8" cols="12">
            <v-textarea
              variant="outlined"
              density="compact"
              v-model="respo.Adresse"
              label="Adresse"
              rows="2"
              :rules="[
                FormRules.required(
                  `L'adresse est requise pour l'émission d'une facture.`
                ),
              ]"
            >
            </v-textarea>
          </v-col>
          <v-col md="2" sm="4" cols="12">
            <v-text-field
              variant="outlined"
              density="compact"
              v-model="respo.CodePostal"
              label="Code postal"
              :rules="[
                FormRules.required(
                  `Le code postal est requis pour l'émission d'une facture.`
                ),
              ]"
            ></v-text-field>
          </v-col>
          <v-col md="3" sm="6" cols="12">
            <v-text-field
              variant="outlined"
              density="compact"
              v-model="respo.Ville"
              label="Ville"
              :rules="[
                FormRules.required(
                  `La ville est requise pour l'émission d'une facture.`
                ),
              ]"
            ></v-text-field>
          </v-col>
          <v-col md="3" sm="6" cols="12">
            <PaysField v-model="respo.Pays"></PaysField>
          </v-col>
        </v-row>
      </v-form>
    </v-card-text>
  </v-card>
</template>

<script lang="ts" setup>
import type { Date_, ResponsableLegal } from "../logic/api";
import { ageFrom, isDateZero } from "@/components/date";
import { FormRules, Formatters } from "@/utils";
import { useDisplay } from "vuetify";

const respo = defineModel<ResponsableLegal>({ required: true });

function checkDateNaissance(d: Date_) {
  if (isDateZero(d)) {
    return "Merci de préciser votre date de naissance.";
  }
  const age = ageFrom(d);
  if (age === null || age < 18) {
    return "Le responsable légal doit être majeur.";
  }
  return true;
}

const { smAndUp } = useDisplay();
</script>
