<template>
  <!-- bienvenue -->
  <v-alert type="info" class="my-2">
    Bienvenu{{ props.equipier.Personne.Sexe == Sexe.Woman ? "e" : "" }}
    <b>{{ props.equipier.Personne.Prenom }}</b
    >, et merci pour ton engagement !<br />
    Ce formulaire te permet de remplir directement les informations et documents
    nécessaires.
    <br />
    Merci de vérifier que tout est à jour...
  </v-alert>

  <!-- joomeo -->
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

  <!-- formulaire -->
  <v-card title="Informations personnelles">
    <v-card-text>
      <v-form class="my-6">
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
            <DepartementField
              label="Département de naissance"
              v-model="inner.DepartementNaissance"
              :rules="[
                FormRules.required(
                  'Ton département (ou pays) est requis par Jeunesse et Sport.'
                ),
              ]"
            ></DepartementField>
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
              :formatter="
                inner.Pays == 'CH' ? Formatters.telCh : Formatters.telFr
              "
              label="Téléphones"
            ></StringList>
          </v-col>
        </v-row>
        <v-row>
          <v-col md="3" sm="6">
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
          <v-col md="3" sm="6">
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
          <v-col md="3" sm="6">
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
          <v-col md="3">
            <PaysField v-model="inner.Pays"></PaysField>
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
          <!-- <v-col>
            <v-text-field
              variant="outlined"
              density="compact"
              v-model="inner.SecuriteSociale"
              label="Sécurité sociale"
              :rules="[
                FormRules.required(
                  `Ton numéro de Sécurité Sociale est requis par l'Ursaf.`
                ),
              ]"
            ></v-text-field>
          </v-col> -->
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
    <CharteEquipier :sexe="inner.Sexe" @update="updateCharte"></CharteEquipier>
  </v-dialog>
</template>

<script lang="ts" setup>
import { computed, reactive, ref } from "vue";
import {
  Categorie,
  CategorieLabels,
  Sexe,
  type EquipierExt,
  type IdDemande,
  type Photos,
  type PublicFile,
} from "../logic/api";
import { Camps, copy, Formatters, FormRules } from "@/utils";
import { controller } from "../logic/logic";
import { isDateZero } from "@/components/date";
import CharteEquipier from "./CharteEquipier.vue";

const props = defineProps<{
  token: string;
  equipier: EquipierExt;
  album: Photos | null;
}>();

const inner = ref(copy(props.equipier.Personne));
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
      (
        inner.value.Nom &&
        inner.value.Prenom &&
        inner.value.Sexe != Sexe.NoSexe &&
        !isDateZero(inner.value.DateNaissance) &&
        inner.value.VilleNaissance &&
        inner.value.DepartementNaissance &&
        inner.value.Mail &&
        inner.value.Adresse &&
        inner.value.CodePostal &&
        inner.value.Profession
      )
      //   inner.value.Pays &&
      //   inner.value.SecuriteSociale
    )
);
async function save() {
  const res = await controller.Update({
    Token: props.token,
    Personne: inner.value,
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
