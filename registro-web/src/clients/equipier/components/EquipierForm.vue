<template>
  <!-- bienvenue -->
  <v-alert type="info" class="my-2">
    Bienvenu{{ props.equipier.Personne.Sexe == Sexe.Woman ? "e" : "" }}
    <b>{{ props.equipier.Personne.Prenom }}</b
    >, et merci pour ton engagement avec l'ACVE !<br />
    Ce formulaire te permet de remplir directement les informations et documents
    nécessaires à Jeunesse et Sport.
    <br />
    Merci de vérifier que tout est à jour...
  </v-alert>

  <!-- joomeo -->
  <v-alert v-if="props.joomeo != null" class="my-2" icon="mdi-panorama-variant">
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
  <v-card title="Informations personnelles">
    <v-card-text>
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
              placeholder="Optionnel"
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
            <v-combobox
              variant="outlined"
              density="compact"
              v-model="inner.DepartementNaissance"
              label="Département de naissance"
              hint="Pays de naissance si hors de France"
              :rules="[
                FormRules.required(
                  'Ton département (ou pays) est requis par Jeunesse et Sport.'
                ),
              ]"
              :items="Departements"
            ></v-combobox>
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
        <v-row>
          <v-col>
            <DiplomeField v-model="inner.Diplome"></DiplomeField>
          </v-col>
          <v-col>
            <ApprofondissementField
              v-model="inner.Approfondissement"
            ></ApprofondissementField>
          </v-col>
          <v-col>
            <v-text-field
              variant="outlined"
              density="compact"
              v-model="inner.Profession"
              label="Profession"
              :rules="[
                FormRules.required(
                  `Ta profession est requise par Jeunesse et Sport.`
                ),
              ]"
            ></v-text-field>
          </v-col>
        </v-row>
        <v-row>
          <v-col>
            <v-text-field
              variant="outlined"
              density="compact"
              v-model="inner.SecuriteSociale"
              label="Sécurité sociale"
              :rules="[
                FormRules.required(
                  `Ton numéro de Sécurité Sociale est requis par Jeunesse et Sport.`
                ),
              ]"
            ></v-text-field>
          </v-col>
          <v-col>
            <v-checkbox
              density="compact"
              hide-details
              v-model="inner.Etudiant"
              label="Etudiant ?"
            ></v-checkbox>
          </v-col>
          <v-col>
            <v-checkbox
              density="compact"
              hide-details
              v-model="inner.Fonctionnaire"
              label="Fonctionnaire ?"
            ></v-checkbox>
          </v-col>
        </v-row>
      </v-form>
    </v-card-text>
  </v-card>

  <v-card
    class="my-2"
    title="Dates de présence"
    subtitle="Tu peux décaler ton jour d'arrivée ou de départ."
  >
    <v-card-text>
      <v-row no-gutters>
        <v-col>
          <DayOffsetField
            label="Arrivée"
            v-model="innerPresences.Debut"
            :min="-6"
            :max="6"
            :ref-date="props.equipier.Camp.DateDebut"
          ></DayOffsetField>
        </v-col>
        <v-divider thickness="4" vertical color="black"></v-divider>
        <v-col>
          <DayOffsetField
            label="Départ"
            v-model="innerPresences.Fin"
            :ref-date="Camps.dateFin(props.equipier.Camp)"
            :min="-6"
            :max="6"
          ></DayOffsetField>
        </v-col>
      </v-row>
    </v-card-text>
  </v-card>

  <v-card
    title="Documents"
    subtitle="Les documents ci-dessous sont à ajouter (ou vérifier)."
    class="my-2"
  >
    <v-card-text>
      <FilesDemande
        :demande="demande.Demande"
        :files="demande.Files || []"
        :optionnelle="demande.Optionnelle"
        v-for="demande in demandes"
      ></FilesDemande>
    </v-card-text>
  </v-card>
</template>

<script lang="ts" setup>
import { computed, ref } from "vue";
import { Sexe, type EquipierExt, type Joomeo } from "../logic/api";
import { Camps, copy, Departements, Formatters, FormRules } from "@/utils";

const props = defineProps<{
  equipier: EquipierExt;
  joomeo: Joomeo | null;
}>();

const inner = ref(copy(props.equipier.Personne));
const innerPresences = ref(copy(props.equipier.Equipier.Presence));

const demandes = computed(() => {
  const out = (props.equipier.Demandes || []).map((p) => p);
  out.sort((a, b) =>
    a.Optionnelle == b.Optionnelle
      ? a.Demande.Id - b.Demande.Id
      : a.Optionnelle
      ? 1
      : -1
  );
  return out;
});
</script>
