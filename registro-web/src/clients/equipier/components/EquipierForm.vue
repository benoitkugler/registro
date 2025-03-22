<template>
  <v-card title="Informations personnelles">
    <v-card-text>
      <!-- bienvenue -->
      <v-alert type="info">
        Bienvenu{{ props.equipier.Personne.Sexe == Sexe.Woman ? "e" : "" }}
        <b>{{ props.equipier.Personne.Prenom }}</b
        >, et merci pour ton engagement avec l'ACVE !<br />
        Ce formulaire te permet de remplir directement les informations et
        documents nécessaires à Jeunesse et Sport.
        <br />
        Merci de vérifier que tout est à jour...
      </v-alert>

      <!-- joomeo -->
      <v-alert
        v-if="props.joomeo != null"
        class="my-2"
        icon="mdi-panorama-variant"
      >
        <v-row>
          <v-col align-self="center" cols="4">
            Tu peux retrouver les
            <a :href="props.joomeo.SpaceURL">photos du séjour sur Joomeo</a>, en
            utilisant ces identifiants :
          </v-col>
          <v-col align-self="center" cols="4" class="text-center">
            Identifiant : <b>{{ props.joomeo.Login }}</b>
          </v-col>
          <v-col align-self="center" cols="4" class="text-center">
            Mot de passe : <b>{{ props.joomeo.Password }}</b>
          </v-col>
        </v-row>
      </v-alert>

      <!-- formulaire -->
      <v-form ref="form" class="my-6">
        <v-row>
          <v-col md="4" sm="6">
            <v-text-field
              variant="outlined"
              density="compact"
              v-model="inner.Nom"
              label="Nom"
              :rules="[FormRules.required('Merci de remplir ton nom.')]"
            ></v-text-field>
          </v-col>
          <v-col md="3" sm="8">
            <v-text-field
              variant="outlined"
              density="compact"
              v-model="inner.NomJeuneFille"
              placeholder="(Optionnel) Nom de jeune fille"
              label="Nom de jeune fille"
            ></v-text-field>
          </v-col>
          <v-col md="3" sm="6">
            <v-text-field
              variant="outlined"
              density="compact"
              v-model="inner.Prenom"
              label="Prénom"
              :rules="[FormRules.required('Merci de remplir ton prénom.')]"
            ></v-text-field>
          </v-col>
          <v-col md="2" sm="4">
            <SexeField
              v-model="inner.Sexe"
              :rules="[FormRules.required('Merci de préciser ton sexe.')]"
            ></SexeField>
          </v-col>
        </v-row>
        <v-row>
          <v-col md="4" sm="12">
            <DateNaissanceField
              v-model="inner.DateNaissance"
              :rule="
                FormRules.requiredDate(
                  'Ta date de naissance est requise par Jeunesse et Sport.'
                )
              "
            ></DateNaissanceField>
          </v-col>
          <v-col md="4" sm="6">
            <v-text-field
              variant="outlined"
              density="compact"
              v-model="inner.VilleNaissance"
              label="Ville de naissance"
              :rules="[
                FormRules.required(
                  'Ta ville de naissance est requise par Jeunesse et Sport.'
                ),
              ]"
            ></v-text-field>
          </v-col>
          <v-col md="4" sm="6">
            <v-autocomplete
              variant="outlined"
              density="compact"
              v-model="inner.DepartementNaissance"
              label="Département de naissance (ou pays si hors de France)"
              :rules="[
                FormRules.required(
                  'Ton département (ou pays) est requis par Jeunesse et Sport.'
                ),
              ]"
              :items="Departements"
            ></v-autocomplete>
          </v-col>
        </v-row>
        <v-row>
          <v-col md="4" sm="6">
            <v-text-field
              variant="outlined"
              density="compact"
              v-model="inner.Mail"
              label="Adresse mail"
              type="email"
              :rules="[
                FormRules.required(
                  'Ton adresse mail est utile pour partager les photos du séjour.'
                ),
              ]"
            ></v-text-field>
          </v-col>
          <v-col md="8" sm="6">
            <StringList
              v-model="inner.Tels"
              :formatter="Formatters.tel"
              label="Téléphones"
            ></StringList>
          </v-col>
        </v-row>
        <v-row>
          <v-col md="4" sm="6">
            <v-textarea
              variant="outlined"
              density="compact"
              v-model="inner.Adresse"
              label="Adresse"
              rows="2"
              :rules="[
                FormRules.required(`Ton adresse est requise par l'URSSAF.`),
              ]"
            >
            </v-textarea>
          </v-col>
          <v-col md="4" sm="6">
            <v-text-field
              variant="outlined"
              density="compact"
              v-model="inner.CodePostal"
              label="Code postal"
              :rules="[
                FormRules.required(`Ton code postal est requis par l'URSSAF.`),
              ]"
            ></v-text-field>
          </v-col>
          <v-col md="4" sm="6">
            <v-text-field
              variant="outlined"
              density="compact"
              v-model="inner.Ville"
              label="Ville"
              :rules="[
                FormRules.required(`Ton adresse est requise par l'URSSAF.`),
              ]"
            ></v-text-field>
          </v-col>
        </v-row>
        <!-- <v-row>
          <v-col>
            <v-form-group label="Diplôme">
              <v-form-select
                :options="optionsDiplome"
                v-model="equipier.diplome"
              >
              </v-form-select>
            </v-form-group>
          </v-col>
          <v-col>
            <v-form-group label="Approfondissement">
              <v-form-select :options="optionsAppro" v-model="equipier.appro">
              </v-form-select>
            </v-form-group>
          </v-col>
        </v-row>
        <v-row>
          <v-col>
            <string-field
              label="Sécurité sociale"
              invalidFeedback="Votre numéro de Sécurité Sociale est requis par Jeunesse et Sport."
              v-model="equipier.securite_sociale"
              type="securite-sociale"
              required
            ></string-field>
          </v-col>
          <v-col>
            <string-field
              label="Profession"
              invalidFeedback="Votre profession est requise par Jeunesse et Sport."
              v-model="equipier.profession"
              required
            ></string-field>
          </v-col>
        </v-row>
        <v-row>
          <v-col>
            <v-form-checkbox v-model="equipier.etudiant"
              >Etudiant ?</v-form-checkbox
            >
          </v-col>
          <v-col>
            <v-form-checkbox v-model="equipier.fonctionnaire"
              >Fonctionnaire ?</v-form-checkbox
            >
          </v-col>
        </v-row>

        <v-card title="Dates de présence" class="mt-3">
          <v-card-text>
            <v-row>
              <v-col>
                <v-form-checkbox v-model="equipier.presence.active">
                  Définir des dates de présence personnalisées
                  <small class="text-muted"
                    >(différentes de celles du séjour :
                    {{ datesSejour }})</small
                  >
                </v-form-checkbox>
              </v-col>
            </v-row>
            <v-row class="mt-2">
              <v-col>
                <date-field
                  label="Date d'arrivée"
                  invalidFeedback="Votre date d'arrivée est requise par Jeunesse et Sport"
                  v-model="equipier.presence.from"
                  :disabled="!equipier.presence.active"
                ></date-field>
              </v-col>
              <v-col>
                <date-field
                  label="Date de départ"
                  invalidFeedback="Votre date de départ est requise par Jeunesse et Sport"
                  v-model="equipier.presence.to"
                  :disabled="!equipier.presence.active"
                ></date-field>
              </v-col>
            </v-row>
          </v-card-text>
        </v-card> -->
      </v-form>
    </v-card-text>
  </v-card>
</template>

<script lang="ts" setup>
import { ref } from "vue";
import { Sexe, type EquipierExt, type Joomeo } from "../logic/api";
import { copy, Departements, Formatters, FormRules } from "@/utils";

const props = defineProps<{
  equipier: EquipierExt;
  joomeo: Joomeo | null;
}>();

const inner = ref(copy(props.equipier.Personne));

// async function fetchData() {
//   const res = await controller.Load({ key: key.value });
//   if (res === undefined) return;
//   data.value = res;
// }
</script>
