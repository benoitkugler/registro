<template>
  <div class="pa-2">
    <v-card :title="props.fiche.Personne" :subtitle="subtitle">
      <v-card-text>
        <v-alert v-if="props.fiche.IsLocked">
          Votre adresse mail n'est pas reconnu dans notre système comme celle du
          responsable de cette fiche sanitaire. Elle est actuellement liée
          {{ ownerMails.length >= 2 ? "aux adresses" : "à l'adresse" }}
          <b>{{ ownerMails.join(" ; ") }}</b
          >.
          <br />

          Afin de protéger vos données, vous devez demander un transfert des
          droits de lecture et d'écriture. Un mail contenant un lien de
          transfert va être envoyé
          {{ ownerMails.length >= 2 ? "aux adresses" : "à l'adresse" }}
          <i>({{ ownerMails.join(" ; ") }})</i>.

          <v-btn
            variant="outlined"
            @click="emit('transfert')"
            block
            class="mt-4"
          >
            <template #prepend>
              <v-icon>mdi-lock-open</v-icon>
            </template>
            Demander le transfert de la fiche sanitaire
          </v-btn>
        </v-alert>
        <v-alert
          v-else-if="props.fiche.State == FichesanitaireState.UpToDate"
          type="success"
        >
          La fiche sanitaire est à jour. Merci !
        </v-alert>
        <v-alert
          v-else-if="props.fiche.State == FichesanitaireState.Outdated"
          type="warning"
        >
          Merci de bien vouloir vérifier que les données ci-dessous sont à jour
          et les valider.
          <br />Nous avons besoin de votre consentement, décrit dans l'encart
          <b>Validation</b> (en bas de page).
        </v-alert>
        <v-alert v-else type="info">
          Merci de compléter la fiche sanitaire ci-dessous.
        </v-alert>
      </v-card-text>
    </v-card>

    <template v-if="!props.fiche.IsLocked">
      <!-- vaccins -->
      <!-- <v-card subtitle="Vaccinations" class="my-2">
        <v-card-text>
          Merci de joindre le scan des pages « vaccinations » du carnet de santé
          du participant. Seul le DTPolio est obligatoire pour être accueilli en
          séjour de vacances. <br />
          <i>
            Si le participant n’a pas les vaccins obligatoires, joindre un
            certificat médical de contre-indication.
          </i>

          <FilesDemande
            class="mt-2"
            :demande="props.fiche.VaccinsDemande"
            :files="props.fiche.VaccinsFiles || []"
            :in-upload="false"
            :optionnelle="null"
            show-upload-text
            @upload="(f) => emit('uploadVaccin', f)"
            @delete="(f) => emit('deleteVaccin', f)"
          ></FilesDemande>
        </v-card-text>
      </v-card> -->

      <v-card class="my-2" title="Santé">
        <v-card-text>
          <v-row
            ><v-col>
              <v-textarea
                label="Difficultés de santé"
                rows="3"
                variant="outlined"
                v-model="inner.DifficultesSante"
                hint="Préciser les problèmes de santé, allergies (non alimentaires), ou tout autre renseignement utile ..."
                persistent-hint
              ></v-textarea> </v-col
          ></v-row>

          <v-row
            ><v-col>
              <v-textarea
                label="Allergies alimentaires"
                rows="3"
                variant="outlined"
                v-model="inner.AllergiesAlimentaires"
              ></v-textarea> </v-col
          ></v-row>

          <v-row
            ><v-col>
              <v-textarea
                label="Traitement en cours et mesures proposées"
                rows="3"
                variant="outlined"
                v-model="inner.TraitementMedical"
              ></v-textarea>
              <v-fade-transition>
                <v-alert
                  :model-value="inner.TraitementMedical != ''"
                  color="blue-lighten-4"
                  type="warning"
                >
                  Merci d'apporter les médicaments prescrits et leur posologie,
                  ainsi que d'informer l'équipe à votre arrivée.
                </v-alert>
              </v-fade-transition>
            </v-col></v-row
          >
        </v-card-text>
      </v-card>

      <!-- Médecin -->
      <v-card title="Médecin traitant" class="my-2">
        <v-card-text>
          <v-row>
            <v-col>
              <v-text-field
                variant="outlined"
                density="compact"
                label="Nom"
                v-model="inner.Medecin.Nom"
                hide-details
              ></v-text-field>
            </v-col>
            <v-col>
              <v-text-field
                variant="outlined"
                density="compact"
                label="Téléphone"
                :model-value="Formatters.tel(inner.Medecin.Tel)"
                @update:model-value="(s) => (inner.Medecin.Tel = s)"
                hide-details
              ></v-text-field>
            </v-col>
          </v-row>
        </v-card-text>
      </v-card>

      <!-- Contacts -->
      <v-card title="Contacts" class="my-2">
        <v-card-text>
          <v-row>
            <v-col>
              <v-text-field
                variant="outlined"
                density="compact"
                readonly
                hide-details
                :model-value="props.fiche.ResponsableNom"
                label="Responsable légal"
              ></v-text-field>
            </v-col>
            <v-col>
              <v-text-field
                variant="outlined"
                density="compact"
                readonly
                hide-details
                :model-value="Formatters.tels(props.fiche.ResponsableTels)"
                label="Téléphone"
              ></v-text-field>
            </v-col>
          </v-row>
          <v-divider thickness="2" class="my-4"></v-divider>
          <v-row>
            <v-col>
              <v-text-field
                variant="outlined"
                density="compact"
                readonly
                v-model="inner.AutreContact.Nom"
                label="Autre contact"
                hint="Nom et prénom d'une autre personne à contacter en cas d'urgence."
                persistent-hint
              ></v-text-field>
            </v-col>
            <v-col>
              <v-text-field
                variant="outlined"
                density="compact"
                label="Téléphone du contact"
                :model-value="Formatters.tel(inner.AutreContact.Tel)"
                @update:model-value="(s) => (inner.AutreContact.Tel = s)"
              ></v-text-field>
            </v-col>
          </v-row>
        </v-card-text>
      </v-card>

      <v-card
        class="mt-4"
        variant="outlined"
        color="success"
        title="Validation"
      >
        <v-card-text class="text-black">
          Le responsable légal déclare <b>exacts</b> les renseignements portés
          sur ce formulaire. <br />
          Il autorise le responsable du séjour à prendre des mesures urgentes de
          soin (traitements médicaux d'urgence, hospitalisation, interventions
          chirurgicales) rendues nécessaires par l’état du participant.
          <br />
          Il autorise également, si nécessaire, le responsable du séjour à faire
          sortir le participant de l’hôpital après une hospitalisation.
        </v-card-text>
        <v-card-actions>
          <v-btn block color="success" @click="emit('save', inner)">
            <template #prepend>
              <v-icon>mdi-content-save</v-icon>
            </template>
            Enregistrer la fiche sanitaire de {{ props.fiche.Personne }}
          </v-btn>
        </v-card-actions>
      </v-card>
    </template>
  </div>
</template>

<script setup lang="ts">
import { copy, Formatters } from "@/utils";
import {
  FichesanitaireState,
  type Fichesanitaire,
  type FichesanitaireExt,
} from "../logic/api";
import { computed, reactive } from "vue";
const props = defineProps<{
  fiche: FichesanitaireExt;
}>();

const emit = defineEmits<{
  (e: "save", fs: Fichesanitaire): void;
  //   (e: "uploadVaccin", file: File): void;
  //   (e: "deleteVaccin", file: PublicFile): void;
  (e: "transfert"): void;
}>();

const inner = reactive(copy(props.fiche.Fichesanitaire));

const ownerMails = computed(() => props.fiche.Fichesanitaire.Owners || []);

const subtitle = computed(() =>
  props.fiche.State == FichesanitaireState.NoFiche
    ? undefined
    : `Dernière modification : ${Formatters.time(
        props.fiche.Fichesanitaire.Modified
      )} / Propriétaire(s) :
         ${ownerMails.value.join(" ; ")}`
);
</script>
