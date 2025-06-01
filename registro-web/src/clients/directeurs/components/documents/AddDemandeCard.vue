<template>
  <v-card
    title="Demander un document"
    subtitle="Les familles devront fournir ce document sur leur espace personnel."
    class="ma-2"
  >
    <template #append>
      <v-btn @click="createDemande">
        <template #prepend>
          <v-icon color="green">mdi-plus</v-icon>
        </template>
        Nouvelle demande</v-btn
      >
    </template>
    <v-card-text>
      <v-list>
        <v-list-item v-if="!props.availableDemandes.length" class="text-center">
          <i>Aucune demande.</i>
        </v-list-item>
        <v-list-item
          v-for="demande in props.availableDemandes"
          :title="demande.Demande.Description"
          :subtitle="
            demande.Demande.IdDirecteur.Valid
              ? 'Demande personnelle'
              : 'Demande globale'
          "
          rounded
          class="my-1"
          @[!props.selectedDemandes.includes(demande.Demande.Id)&&`click`]="
            emit('selected', demande.Demande)
          "
        >
          <template
            #prepend
            v-if="props.selectedDemandes.includes(demande.Demande.Id)"
          >
            <v-icon color="green">mdi-check</v-icon>
          </template>
          <template #append>
            <div class="mr-4" v-if="demande.File.Id != 0">
              <FileCard
                :file="demande.File"
                @delete="fileToDeleteDemande = demande"
              ></FileCard>
            </div>
            <!-- upload button -->
            <v-btn
              v-else
              size="small"
              class="mr-2"
              @click="showFileUploadDialog(demande)"
            >
              <template #prepend>
                <v-icon color="green" icon="mdi-upload"> </v-icon>
              </template>
              Document à remplir
            </v-btn>

            <v-btn
              icon
              size="x-small"
              class="mx-1"
              @click.stop="toEdit = demande"
              :disabled="!demande.Demande.IdDirecteur.Valid"
              ><v-icon>mdi-pencil</v-icon>
            </v-btn>
            <v-btn
              icon
              size="x-small"
              class="mx-1"
              @click.stop="toDelete = demande.Demande"
              :disabled="!demande.Demande.IdDirecteur.Valid"
            >
              <v-icon color="red">mdi-delete</v-icon></v-btn
            >
          </template>
        </v-list-item>
      </v-list>
    </v-card-text>

    <!-- delete demande -->
    <v-dialog
      :model-value="toDelete != null"
      @update:model-value="toDelete = null"
      max-width="400px"
    >
      <v-card title="Confirmer la suppression" v-if="toDelete">
        <v-card-text>
          Confirmez-vous la suppression de cette demande ? Les éventuels
          documents déjà téléversés seront supprimés.
          <br /><br />

          Attention, cette opération est irréversible.
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="red" @click="deleteDemande">Supprimer</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <v-dialog
      :model-value="toEdit != null"
      @update:model-value="toEdit = null"
      max-width="600px"
    >
      <v-card v-if="toEdit" title="Document à fournir">
        <v-card-text>
          <v-row
            ><v-col>
              <v-text-field
                density="compact"
                variant="outlined"
                label="Description"
                v-model="toEdit.Demande.Description"
                hint="Texte affiché sur l'espace de suivi."
              ></v-text-field> </v-col
          ></v-row>
          <v-row>
            <v-col>
              <IntField
                density="compact"
                variant="outlined"
                label="Document temporaire"
                :min="(0 as Int)"
                suffix="jours"
                v-model="toEdit.Demande.JoursValide"
                hint="Nombre de jours de validité (0 pour un document permanent)"
              ></IntField>
            </v-col>
            <v-col>
              <IntField
                density="compact"
                variant="outlined"
                label="Nombre maximum de fichiers"
                v-model="toEdit.Demande.MaxDocs"
                hint="Limite le nombre de fichiers pouvant être déposés."
              ></IntField>
            </v-col>
          </v-row>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn @click="updateDemande">Enregistrer</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <v-dialog
      :model-value="fileToUploadDemande != null"
      @update:model-value="fileToUploadDemande = null"
      max-width="600px"
    >
      <v-card
        title="Ajouter un document"
        subtitle="Document à remplir par les familles"
      >
        <v-card-text>
          <FileInput
            ref="fileInput"
            @update="(f) => (fileToUpload = f)"
          ></FileInput>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn :disabled="!fileToUpload" @click="uploadFile"
            >Téléverser</v-btn
          >
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- delete file -->
    <v-dialog
      :model-value="fileToDeleteDemande != null"
      @update:model-value="fileToDeleteDemande = null"
      max-width="400px"
    >
      <v-card title="Confirmer la suppression" v-if="fileToDeleteDemande">
        <v-card-text>
          Confirmez-vous la suppression du document
          <i> {{ fileToDeleteDemande.File.NomClient }} </i> ? <br /><br />

          Attention, cette opération est irréversible.
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="red" @click="deleteFile">Supprimer</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-card>
</template>

<script lang="ts" setup>
import { computed, nextTick, ref, useTemplateRef } from "vue";
import { controller } from "../../logic/logic";
import type {
  Demande,
  DemandeDirecteur,
  IdDemande,
  IdFile,
  Int,
  PublicFile,
  Time,
} from "../../logic/api";
import { copy } from "@/utils";

const props = defineProps<{
  availableDemandes: DemandeDirecteur[];
  selectedDemandes: IdDemande[];
}>();

const emit = defineEmits<{
  (e: "selected", demande: Demande): void;
  (e: "created", demande: DemandeDirecteur): void;
  (e: "updated", demande: DemandeDirecteur): void;
  (e: "deleted", demande: Demande): void;
}>();

async function createDemande() {
  const res = await controller.DocumentsCreateDemande();
  if (res === undefined) return;
  emit("created", res);
  // start edit
  toEdit.value = copy(res);
}

const toEdit = ref<DemandeDirecteur | null>(null);
async function updateDemande() {
  const demande = toEdit.value;
  if (demande == null) return;
  toEdit.value = null;
  const res = await controller.DocumentsUpdateDemande(demande.Demande);
  if (res === undefined) return;
  controller.showMessage("Modifications enregistrées avec succès.");
  emit("updated", demande);
}

const toDelete = ref<Demande | null>(null);
async function deleteDemande() {
  const demande = toDelete.value;
  if (demande == null) return;
  toDelete.value = null;
  const res = await controller.DocumentsDeleteDemande({
    idDemande: demande.Id,
  });
  if (res === undefined) return;
  controller.showMessage("Demande supprimée avec succès.");
  emit("deleted", demande);
}

const fileToUploadDemande = ref<DemandeDirecteur | null>(null);
const fileInput = useTemplateRef("fileInput");
const fileToUpload = ref<File | null>(null);
function showFileUploadDialog(demande: DemandeDirecteur) {
  fileToUploadDemande.value = demande;
  nextTick(() => fileInput.value?.openDialog());
}
async function uploadFile() {
  const demande = fileToUploadDemande.value;
  const file = fileToUpload.value;
  if (file == null || demande == null) return;
  fileToUploadDemande.value = null;
  fileToUpload.value = null;
  const res = await controller.DocumentsUploadDemandeFile(file, {
    idDemande: demande.Demande.Id,
  });
  if (res === undefined) return;
  controller.showMessage("Document téléversé avec succès.");
  emit("updated", { Demande: demande.Demande, File: res });
}

const fileToDeleteDemande = ref<DemandeDirecteur | null>(null);
async function deleteFile() {
  const demande = fileToDeleteDemande.value;
  if (demande == null) return;
  fileToDeleteDemande.value = null;
  const res = await controller.DocumentsDeleteDemandeFile({
    key: demande.File.Key,
  });
  if (res === undefined) return;
  controller.showMessage("Document supprimé avec succès.");
  emit("updated", {
    Demande: demande.Demande,
    File: {
      Id: 0 as IdFile,
      Key: "",
      NomClient: "",
      Taille: 0 as Int,
      Uploaded: "" as Time,
    } satisfies PublicFile,
  });
}
</script>
