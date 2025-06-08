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
          v-else-if="
            props.fiche.State == FichesanitaireState.UpToDate &&
            props.fiche.VaccinsFiles?.length
          "
          type="success"
        >
          La fiche sanitaire est à jour. Merci !
        </v-alert>
        <v-alert
          v-else-if="props.fiche.State == FichesanitaireState.UpToDate"
          type="warning"
        >
          La fiche sanitaire est à jour, mais vous n'avez déposé aucun vaccin.
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
      <v-card subtitle="Vaccinations" class="my-2">
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
      </v-card>

      <!-- handicap -->
      <v-card subtitle="Handicap" class="my-2">
        <v-card-text>
          <v-checkbox
            v-model="inner.Handicap"
            color="primary"
            label="Le participant est porteur de handicap. "
            hide-details
          ></v-checkbox>
          <v-fade-transition>
            <v-alert v-model="inner.Handicap" color="blue-lighten-4">
              Merci de <b>contacter</b> au plus vite le directeur afin qu’un
              accueil individualisé puisse être mis en place.
            </v-alert>
          </v-fade-transition>
        </v-card-text>
      </v-card>

      <!-- traitement -->
      <v-card subtitle="Traitement" class="my-2">
        <v-card-text>
          <v-checkbox
            v-model="inner.TraitementMedical"
            color="primary"
            label="Le participant suit un traitement médical pendant le séjour."
            hide-details
          ></v-checkbox>
          <v-fade-transition>
            <v-alert v-model="inner.TraitementMedical" color="blue-lighten-4">
              Merci de joindre une <b>ordonnance</b> récente et les
              <b>médicaments</b>
              correspondants (avec la notice). Aucun médicament ne pourra être
              pris sans ordonnance.
            </v-alert>
          </v-fade-transition>
        </v-card-text>
      </v-card>

      <!-- maladies -->
      <v-card subtitle="Maladies" class="my-2">
        <v-card-text>
          Le participant a-t-il déjà eu les maladies suivantes ?
          <v-row class="mt-2">
            <v-col cols="3" class="py-0">
              <v-checkbox
                v-model="inner.Maladies.Rubeole"
                color="primary"
                label="Rubéole"
                hide-details
              ></v-checkbox>
            </v-col>
            <v-col cols="3" class="py-0">
              <v-checkbox
                v-model="inner.Maladies.Varicelle"
                color="primary"
                label="Varicelle"
                hide-details
              ></v-checkbox>
            </v-col>
            <v-col cols="3" class="py-0">
              <v-checkbox
                v-model="inner.Maladies.Angine"
                color="primary"
                label="Angine"
                hide-details
              ></v-checkbox>
            </v-col>
            <v-col cols="3" class="py-0">
              <v-checkbox
                v-model="inner.Maladies.Oreillons"
                color="primary"
                label="Oreillons"
                hide-details
              ></v-checkbox>
            </v-col>
            <v-col cols="3" class="py-0">
              <v-checkbox
                v-model="inner.Maladies.Scarlatine"
                color="primary"
                label="Scarlatine"
                hide-details
              ></v-checkbox>
            </v-col>
            <v-col cols="3" class="py-0">
              <v-checkbox
                v-model="inner.Maladies.Coqueluche"
                color="primary"
                label="Coqueluche"
                hide-details
              ></v-checkbox>
            </v-col>
            <v-col cols="3" class="py-0">
              <v-checkbox
                v-model="inner.Maladies.Otite"
                color="primary"
                label="Otite"
                hide-details
              ></v-checkbox>
            </v-col>
            <v-col cols="3" class="py-0">
              <v-checkbox
                v-model="inner.Maladies.Rougeole"
                color="primary"
                label="Rougeole"
                hide-details
              ></v-checkbox>
            </v-col>
            <v-col cols="3" class="py-0">
              <v-checkbox
                v-model="inner.Maladies.Rhumatisme"
                color="primary"
                label="Rhumatisme articulaire aigü "
                hide-details
              ></v-checkbox>
            </v-col>
          </v-row>
        </v-card-text>
      </v-card>

      <!-- allergies -->
      <v-card subtitle="Allergies" class="my-2">
        <v-card-text>
          Le participant a-t-il une des allergies suivantes ?
          <v-row class="mt-2">
            <v-col align-self="center" cols="3" class="py-0">
              <v-checkbox
                v-model="inner.Allergies.Asthme"
                color="primary"
                label="Asthme"
                hide-details
              ></v-checkbox>
            </v-col>
            <v-col align-self="center" cols="3" class="py-0">
              <v-checkbox
                v-model="inner.Allergies.Alimentaires"
                color="primary"
                label="Alimentaires"
                hide-details
              ></v-checkbox>
            </v-col>
            <v-col align-self="center" cols="3" class="py-0">
              <v-checkbox
                v-model="inner.Allergies.Medicamenteuses"
                color="primary"
                label="Médicamenteuses"
                hide-details
              ></v-checkbox>
            </v-col>
            <v-col align-self="center" cols="3" class="py-0">
              <v-text-field
                density="compact"
                variant="outlined"
                hide-details
                label="Autres allergies"
                v-model="inner.Allergies.Autres"
              ></v-text-field>
            </v-col>
          </v-row>
          <v-fade-transition>
            <v-row v-if="showConduiteATenir">
              <v-col>
                <v-textarea
                  variant="outlined"
                  label="Conduite à tenir"
                  v-model="inner.Allergies.ConduiteATenir"
                  hint="Préciser la cause de l’allergie et la conduite à tenir (si automédication le signaler). Si un PAI a été mis en place durant le temps scolaire, merci de nous le signaler."
                  persistent-hint
                  :rules="[
                    FormRules.required(
                      `Merci de nous informer de la conduite à tenir en cas d'allergie.`
                    ),
                  ]"
                ></v-textarea>
              </v-col>
            </v-row>
          </v-fade-transition>
        </v-card-text>
      </v-card>

      <!-- divers -->
      <v-card subtitle="Divers" class="my-2">
        <v-card-text>
          <v-row>
            <v-col>
              <v-textarea
                variant="outlined"
                label="Difficultés de santé"
                v-model="inner.DifficultesSante"
                hint="Indiquer les difficultés de santé (maladie, accident, crises convulsives, hospitalisation, opération, rééducation) en précisant les dates et les précautions à prendre."
                persistent-hint
              ></v-textarea>
            </v-col>
            <v-col>
              <v-textarea
                variant="outlined"
                label="Recommandations utiles"
                v-model="inner.Recommandations"
                hint="Préciser si le participant porte des lunettes, des lentilles, des prothèses auditives, des prothèses dentaires..."
                persistent-hint
              ></v-textarea>
            </v-col>
          </v-row>
        </v-card-text>
      </v-card>

      <!-- médecin -->
      <v-card subtitle="Médecin traitant" class="my-2">
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

      <!-- Responsable -->
      <v-card subtitle="Responsable légal" class="my-2">
        <v-card-text>
          <v-row>
            <v-col align-self="center">
              Téléphone :
              <v-chip label v-if="!props.fiche.RespoTels?.length"
                >Aucun numéro</v-chip
              >
              <v-chip label v-for="tel in props.fiche.RespoTels" class="mx-1">{{
                Formatters.tel(tel)
              }}</v-chip>
            </v-col>
            <v-col align-self="center">
              <v-text-field
                variant="outlined"
                density="compact"
                label="Numéro de contact"
                :model-value="Formatters.tel(inner.Tel)"
                @update:model-value="(s) => (inner.Tel = s)"
                hint="Numéro de téléphone supplémentaire (optionnel)"
                persistent-hint
              ></v-text-field>
            </v-col>
          </v-row>
          <v-row>
            <v-col>
              <v-text-field
                variant="outlined"
                density="compact"
                label="Sécurité sociale"
                v-model="respoSecuriteSociale"
                hide-details
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
          sur ce formulaire et autorise le responsable du séjour à prendre, le
          cas échéant, <b>toutes mesures</b> (traitement médical,
          hospitalisation, intervention chirurgicale) rendues nécessaires par
          l'état du participant. Il autorise également, si nécessaire, le
          directeur du séjour à faire sortir le participant de l’hôpital après
          une hospitalisation.
        </v-card-text>
        <v-card-actions>
          <v-btn
            block
            color="success"
            @click="emit('save', inner, respoSecuriteSociale)"
          >
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
import { copy, Formatters, FormRules } from "@/utils";
import {
  FichesanitaireState,
  type Fichesanitaire,
  type FichesanitaireExt,
  type PublicFile,
} from "../logic/api";
import { computed, reactive, ref } from "vue";
const props = defineProps<{
  fiche: FichesanitaireExt;
}>();

const emit = defineEmits<{
  (e: "save", fs: Fichesanitaire, respoSecuriteSociale: string): void;
  (e: "uploadVaccin", file: File): void;
  (e: "deleteVaccin", file: PublicFile): void;
  (e: "transfert"): void;
}>();

const inner = reactive(copy(props.fiche.Fichesanitaire));
const respoSecuriteSociale = ref(props.fiche.RespoSecuriteSociale);

const showConduiteATenir = computed(
  () =>
    inner.Allergies.Asthme ||
    inner.Allergies.Alimentaires ||
    inner.Allergies.Medicamenteuses ||
    inner.Allergies.Autres.length != 0
);

const ownerMails = computed(() => props.fiche.Fichesanitaire.Mails || []);

const subtitle = computed(() =>
  props.fiche.State == FichesanitaireState.NoFiche
    ? undefined
    : `Dernière modification : ${Formatters.time(
        props.fiche.Fichesanitaire.LastModif
      )} / Propriétaire(s) :
         ${ownerMails.value.join(" ; ")}`
);
</script>
