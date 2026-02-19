<template>
  <!-- bienvenue -->
  <v-alert type="info" class="my-2">
    Bienvenu{{ props.equipier.PersonneBase.Sexe == Sexe.Woman ? "e" : "" }}
    <b>{{ props.equipier.PersonneBase.Prenom }}</b
    >, et merci pour ton engagement !<br />
    Ce formulaire te permet de remplir directement les informations et documents
    nécessaires.
    <br />
    Merci de vérifier que tout est à jour...
  </v-alert>

  <!-- album photo -->
  <v-alert
    v-if="props.album?.HasAlbum"
    class="my-2"
    icon="mdi-panorama-variant"
  >
    <v-row no-gutters>
      <v-col align-self="center"> Lien vers les photos du séjour </v-col>
      <v-col align-self="center" class="text-right">
        <v-btn
          :href="props.album.URL"
          target="_blank"
          prepend-icon="mdi-open-in-new"
          >Album photos</v-btn
        >
      </v-col>
    </v-row>
  </v-alert>

  <!-- formulaire commun ACVE/Repere -->
  <!-- Nom Prénom Sexe DateNaissance Adresse CodePostal Ville Pays Mail -->
  <v-card title="Informations personnelles">
    <v-card-text>
      <v-form class="my-6">
        <v-row>
          <v-col md="3" sm="6">
            <v-text-field
              variant="outlined"
              density="compact"
              v-model="innerBase.Nom"
              label="Nom"
              :rules="[FormRules.required('Merci de remplir ton nom.')]"
            ></v-text-field>
          </v-col>
          <v-col md="3" sm="6">
            <v-text-field
              variant="outlined"
              density="compact"
              v-model="innerBase.Prenom"
              label="Prénom"
              :rules="[FormRules.required('Merci de remplir ton prénom.')]"
            ></v-text-field>
          </v-col>
          <v-col md="2" sm="4">
            <SexeField
              v-model="innerBase.Sexe"
              :rules="[FormRules.required('Merci de préciser ton sexe.')]"
            ></SexeField>
          </v-col>
          <v-col md="4" sm="12">
            <DateNaissanceField
              v-model="innerBase.DateNaissance"
              :rule="
                FormRules.requiredDate(
                  'Merci de préciser ta date de naissance.'
                )
              "
            ></DateNaissanceField>
          </v-col>
        </v-row>
        <v-row>
          <v-col md="3" sm="6">
            <v-textarea
              variant="outlined"
              density="compact"
              v-model="innerBase.Adresse"
              label="Adresse"
              rows="2"
              :rules="[FormRules.required(`Merci de préciser ton adresse.`)]"
            >
            </v-textarea>
          </v-col>
          <v-col md="3" sm="6">
            <v-text-field
              variant="outlined"
              density="compact"
              v-model="innerBase.CodePostal"
              label="Code postal"
              :rules="[
                FormRules.required(`Merci de préciser ton code postal.`),
              ]"
            ></v-text-field>
          </v-col>
          <v-col md="3" sm="6">
            <v-text-field
              variant="outlined"
              density="compact"
              v-model="innerBase.Ville"
              label="Ville"
              :rules="[FormRules.required(`Merci de préciser ton adresse.`)]"
            ></v-text-field>
          </v-col>
          <v-col md="3">
            <PaysField v-model="innerBase.Pays"></PaysField>
          </v-col>
        </v-row>
        <v-row>
          <v-col md="5" sm="6">
            <v-text-field
              variant="outlined"
              density="compact"
              v-model="innerBase.Mail"
              label="Adresse mail"
              type="email"
              :rules="[
                FormRules.required(
                  'Ton adresse mail est utile pour partager les photos du séjour.'
                ),
              ]"
            ></v-text-field>
          </v-col>
          <v-col md="7" sm="6">
            <StringList
              v-model="innerBase.Tels"
              :formatter="
                innerBase.Pays == 'CH' ? Formatters.telCh : Formatters.telFr
              "
              label="Téléphones"
            ></StringList>
          </v-col>
        </v-row>

        <!-- champs Repère -->
        <v-row v-if="asso == 'repere'">
          <v-col>
            <NationaliteField
              v-model="innerBase.Nationnalite"
            ></NationaliteField>
          </v-col>
          <v-col>
            <EtatcivilField
              v-model="innerDetails.EtatCivil"
              :rules="[FormRules.required('Merci de préciser ton état civil.')]"
            ></EtatcivilField>
          </v-col>
          <v-col>
            <IntField
              label="Nombre d'enfants"
              v-model="innerDetails.NombreEnfants"
              :min="(0 as Int)"
            ></IntField>
          </v-col>
        </v-row>
      </v-form>
    </v-card-text>
  </v-card>

  <!-- Champs ACVE -->
  <!-- Diplome Approfondissement SecuriteSociale Fonctionnaire-->
  <!-- TODO -->

  <v-card v-if="asso == 'acve'">
    <v-row>
      <v-col>
        <DiplomeField v-model="innerDetails.Diplome"></DiplomeField>
      </v-col>
      <v-col>
        <ApprofondissementField
          v-model="innerDetails.Approfondissement"
        ></ApprofondissementField>
      </v-col>
    </v-row>

    <!-- <v-row>
          <v-col md="6">
            <v-text-field
              variant="outlined"
              density="compact"
              v-model="innerDetails.SecuriteSociale"
              label="Sécurité sociale"
              :rules="[
                FormRules.required(
                  `Merci de préciser ton numéro de Sécurité Sociale.`
                ),
              ]"
            ></v-text-field>
          </v-col>
        </v-row> -->

    <!-- <v-row>
          <v-col>
            <v-text-field
              variant="outlined"
              density="compact"
              v-model="innerDetails.Profession"
              label="Profession"
              :rules="[FormRules.required(`Merci de préciser ta profession.`)]"
            ></v-text-field>
          </v-col>
         
        </v-row>
        <v-row>
          <v-col>
            <v-checkbox
              density="compact"
              hide-details
              v-model="innerDetails.Fonctionnaire"
              label="Je suis fonctionnaire"
            ></v-checkbox>
          </v-col>
        </v-row> -->
  </v-card>

  <!-- Champs Repère -->
  <template v-if="asso == 'repere'">
    <v-card title="Expérience" class="my-2">
      <v-card-text>
        <v-form>
          <v-row>
            <v-col>
              <v-textarea
                label="Formation"
                variant="outlined"
                density="compact"
                hint="Jeunesse et Sport, BAFA, cuisine, autres..."
                v-model="innerDetails.Formation"
                rows="2"
                :rules="[FormRules.required('Merci de décrire ta formation.')]"
              ></v-textarea>
            </v-col>
          </v-row>
          <v-row>
            <v-col>
              <v-text-field
                label="Profession"
                variant="outlined"
                density="compact"
                v-model="innerDetails.Profession"
                :rules="[FormRules.required(`Merci d'indiquer ta profession.`)]"
              ></v-text-field>
            </v-col>
          </v-row>
          <v-row>
            <v-col>
              <v-textarea
                label="Expérience avec les jeunes"
                variant="outlined"
                density="compact"
                hint="Expérience(s) de travail parmi les enfants et les jeunes"
                v-model="innerDetails.ExperienceTravailJeunes"
                rows="2"
                :rules="[
                  FormRules.required('Merci de décrire tes expériences.'),
                ]"
              ></v-textarea>
            </v-col>
          </v-row>
        </v-form>
      </v-card-text>
    </v-card>

    <v-card title="Expérience spirituelle">
      <v-card-text>
        <v-row>
          <v-col>
            <v-textarea
              label="Parcours spirituel"
              variant="outlined"
              density="compact"
              hint="Décrit tes étapes de découverte de la foi, ta relation avec Dieu, ta conversion, tes étapes d'engagement, tes doutes..."
              v-model="innerDetails.ParcoursSpirituel"
              rows="5"
              :rules="[
                FormRules.required('Merci de décrire ton parcours spirituel.'),
              ]"
            ></v-textarea>
          </v-col>
        </v-row>
        <v-row>
          <v-col>
            <v-text-field
              label="Église ou communauté fréquentée "
              variant="outlined"
              density="compact"
              v-model="innerDetails.Eglise"
              :rules="[FormRules.required(`Merci d'indiquer ton église.`)]"
            ></v-text-field>
          </v-col>
        </v-row>
      </v-card-text>
    </v-card>

    <v-card
      title="Contact pour recommandation"
      subtitle="Indique un responsable de ton église locale qui peut te recommander"
      class="my-2"
    >
      <v-card-text>
        <v-row>
          <v-col>
            <v-text-field
              label="Nom"
              variant="outlined"
              density="compact"
              v-model="innerDetails.Recommandation.Nom"
              :rules="[FormRules.required(`Merci d'indiquer un contact.`)]"
            ></v-text-field>
          </v-col>
          <v-col>
            <v-text-field
              label="Prénom"
              variant="outlined"
              density="compact"
              v-model="innerDetails.Recommandation.Prenom"
              :rules="[FormRules.required(`Merci d'indiquer un contact.`)]"
            ></v-text-field>
          </v-col>
          <v-col>
            <v-text-field
              label="Mail"
              variant="outlined"
              density="compact"
              v-model="innerDetails.Recommandation.Mail"
              :rules="[FormRules.validMail()]"
            ></v-text-field>
          </v-col>
          <v-col>
            <v-text-field
              label="Téléphone"
              variant="outlined"
              density="compact"
              v-model="innerDetails.Recommandation.Tel"
              :rules="[FormRules.required(`Merci d'indiquer un contact.`)]"
            ></v-text-field>
          </v-col>
        </v-row>
      </v-card-text>
    </v-card>

    <v-card title="Santé">
      <v-card-text>
        <v-row>
          <v-col>
            <v-textarea
              label="Informations sur ma santé"
              variant="outlined"
              density="compact"
              hint="Éventuelles informations sur ta santé qui pourraient avoir un impact sur la vie du camp."
              v-model="innerDetails.Sante"
              rows="2"
            ></v-textarea>
          </v-col>
        </v-row>
        <v-row>
          <v-col>
            <v-text-field
              label="Assurance Maladie"
              variant="outlined"
              density="compact"
              v-model="innerDetails.AssuranceMaladie"
              :rules="[
                FormRules.required(`Merci d'indiquer ton assurance Maladie.`),
              ]"
            ></v-text-field>
          </v-col>
          <v-col>
            <v-text-field
              label="Assurance Accident ou Responsabilité civile"
              variant="outlined"
              density="compact"
              v-model="innerDetails.AssuranceAccident"
              :rules="[
                FormRules.required(`Merci d'indiquer ton assurance Accident.`),
              ]"
            ></v-text-field>
          </v-col>
          <v-col>
            <v-text-field
              label="Numéro AVS"
              variant="outlined"
              density="compact"
              v-model="innerDetails.SecuriteSociale"
              hint="Si tu vis en Suisse ou as la nationalité Suisse."
            ></v-text-field>
          </v-col>
        </v-row>
      </v-card-text>
    </v-card>

    <v-card title="Membre de l'association" class="my-2">
      <v-card-text>
        <v-row>
          <v-col>
            <v-checkbox
              label="En m'engageant comme équipier je deviens membre de l'association l'année de mon engagement. Je souhaite faire partie de l'association Repère en tant que membre au-delà du camp et être informé de la tenue des assemblées générales."
            ></v-checkbox>
          </v-col>
        </v-row>
      </v-card-text>
    </v-card>
  </template>

  <v-card
    v-if="asso == 'acve'"
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
    <template #append>
      <v-chip
        :color="missingFiles.length ? 'orange' : 'green'"
        :prepend-icon="missingFiles.length ? 'mdi-alert' : 'mdi-check'"
      >
        {{
          missingFiles.length ? `${missingFiles.length}  à fournir` : "A jour"
        }}
      </v-chip>
    </template>
    <v-card-text>
      <FilesDemande
        v-for="demande in demandes"
        :demande="demande.Demande"
        :files="demande.Files || []"
        :in-upload="isUploading.has(demande.Demande.Id)"
        :optionnelle="demande.Optionnelle"
        @upload="(f) => uploadFile(f, demande.Demande.Id)"
        @delete="(f) => deleteFile(f, demande.Demande.Id)"
      ></FilesDemande>
    </v-card-text>
  </v-card>

  <v-row>
    <v-col cols="3">
      <v-btn
        :color="innerCharte.Bool ? 'green' : 'info'"
        @click="showCharteDialog = true"
        :variant="innerCharte.Bool ? 'outlined' : undefined"
      >
        <template #prepend>
          <v-icon v-if="innerCharte.Bool">mdi-check</v-icon>
        </template>
        Charte d'engagement
      </v-btn>
    </v-col>
    <v-spacer></v-spacer>
    <v-col cols="6">
      <v-btn block color="green" @click="save" :disabled="!isFormValid">
        Enregistrer</v-btn
      >
    </v-col>
  </v-row>

  <v-dialog v-model="showMissingFilesDialog" max-width="600px">
    <v-card title="Documents manquants">
      <v-card-text>
        Attention, il te reste à fournir les documents suivants :

        <div class="my-2 text-center">
          <v-chip v-for="piece in missingFiles" color="warning">
            {{ CategorieLabels[piece.Demande.Categorie] }}
          </v-chip>
        </div>

        Merci de revenir les ajouter !
      </v-card-text>
    </v-card>
  </v-dialog>

  <v-dialog v-model="showCharteDialog">
    <CharteEquipierACVE
      v-if="asso == 'acve'"
      :sexe="innerBase.Sexe"
      @update="updateCharte"
    ></CharteEquipierACVE>
    <CharteEquipierRepere
      v-if="asso == 'repere'"
      :sexe="innerBase.Sexe"
      @update="updateCharte"
    ></CharteEquipierRepere>
  </v-dialog>
</template>

<script lang="ts" setup>
import { computed, reactive, ref } from "vue";
import {
  Categorie,
  CategorieLabels,
  EtatCivil,
  Sexe,
  type EquipierExt,
  type IdDemande,
  type Photos,
  type PublicFile,
  type Int,
} from "../logic/api";
import { Camps, copy, Formatters, FormRules } from "@/utils";
import { controller } from "../logic/logic";
import { isDateZero } from "@/components/date";
import CharteEquipierACVE from "./CharteEquipierACVE.vue";
import CharteEquipierRepere from "./CharteEquipierRepere.vue";

const props = defineProps<{
  token: string;
  equipier: EquipierExt;
  album: Photos | null;
}>();

const asso = import.meta.env.VITE_ASSO;

const innerBase = ref(copy(props.equipier.PersonneBase));
const innerDetails = ref(copy(props.equipier.PersonneDetails));
const innerPresences = ref(copy(props.equipier.Equipier.Presence));
const innerCharte = ref(copy(props.equipier.Equipier.AccepteCharte));
const demandesL = ref(copy(props.equipier.Demandes || []));

const demandes = computed(() => {
  const out = demandesL.value.map((p) => p);
  out.sort((a, b) =>
    a.Optionnelle == b.Optionnelle
      ? a.Demande.Id - b.Demande.Id
      : a.Optionnelle
      ? 1
      : -1
  );
  return out;
});

const isUploading = reactive(new Set<IdDemande>());
async function uploadFile(file: File, idDemande: IdDemande) {
  isUploading.add(idDemande);
  const res = await controller.UploadDocument(file, {
    token: props.token,
    idDemande,
  });
  isUploading.delete(idDemande);
  if (res === undefined) return;
  controller.showMessage("Document téléversé avec succès.");
  const item = demandesL.value.find((d) => d.Demande.Id == idDemande)!;
  item.Files = (item.Files || []).concat(res);
}

async function deleteFile(file: PublicFile, idDemande: IdDemande) {
  const res = await controller.DeleteDocument({ key: file.Key });
  if (res === undefined) return;
  controller.showMessage("Document supprimé avec succès.");
  const item = demandesL.value.find((d) => d.Demande.Id == idDemande)!;
  item.Files = (item.Files || []).filter((f) => f.Id != file.Id);
}

const missingFiles = computed(() => {
  const bafa = demandesL.value.find(
    (p) => p.Demande.Categorie == Categorie.Bafa
  );
  const bafd = demandesL.value.find(
    (p) => p.Demande.Categorie == Categorie.Bafd
  );

  return demandesL.value
    .filter((dem) => !dem.Optionnelle) // contraintes bloquantes uniquement
    .filter((dem) => {
      // doc nons remplis
      const notOK = (dem.Files || []).length == 0;
      // cas spécial pour les catégories équivalent bafa et équivalent bafd,
      let okByEquiv = false;
      if (dem.Demande.Categorie == Categorie.BafaEquiv) {
        okByEquiv = bafa != undefined && (bafa.Files || []).length != 0;
      }
      if (dem.Demande.Categorie == Categorie.BafdEquiv) {
        okByEquiv = bafd != undefined && (bafd.Files || []).length != 0;
      }
      return notOK && !okByEquiv;
    });
});

const showMissingFilesDialog = ref(false);
const showCharteDialog = ref(false);

const isFormValid = computed(
  () =>
    !!(
      // champs communs ACVE/Repere
      (
        innerBase.value.Nom &&
        innerBase.value.Prenom &&
        innerBase.value.Sexe != Sexe.NoSexe &&
        !isDateZero(innerBase.value.DateNaissance) &&
        innerBase.value.Adresse &&
        innerBase.value.CodePostal &&
        innerBase.value.Ville &&
        innerBase.value.Pays &&
        innerBase.value.Mail &&
        innerBase.value.Tels?.length &&
        innerDetails.value.SecuriteSociale &&
        ((asso == "repere" && isFormRepereValid()) ||
          (asso == "acve" && isFormAcveValid()))
      )
    )
);

function isFormRepereValid() {
  const d = innerDetails.value;
  return !!(
    d.EtatCivil != EtatCivil.NoEtatCivil &&
    d.Formation &&
    d.Profession &&
    d.ExperienceTravailJeunes &&
    d.ParcoursSpirituel &&
    d.Eglise &&
    d.Recommandation.Nom &&
    d.Recommandation.Prenom &&
    d.Recommandation.Mail &&
    d.Recommandation.Tel &&
    d.AssuranceMaladie &&
    d.AssuranceAccident
  );
}

function isFormAcveValid() {
  // TODO
  return true;
}

async function save() {
  const res = await controller.Update({
    Token: props.token,
    PersonneBase: innerBase.value,
    PersonneDetails: innerDetails.value,
    Presence: innerPresences.value,
  });
  if (res === undefined) return;
  controller.showMessage("Profil mis à jour avec succès.");
  // s'il y a des documents manquants, affiche une alerte
  if (missingFiles.value.length) {
    showMissingFilesDialog.value = true;
  }
  // si la charte n'a pas encore été remplie, affiche la
  if (!innerCharte.value.Valid) {
    showCharteDialog.value = true;
  }
}

async function updateCharte(accept: boolean) {
  showCharteDialog.value = false;
  const res = await controller.UpdateCharte({ token: props.token, accept });
  if (res === undefined) return;
  innerCharte.value = { Valid: true, Bool: accept };
  controller.showMessage("Ton avis a bien été pris en compte. Merci !");
}
</script>
