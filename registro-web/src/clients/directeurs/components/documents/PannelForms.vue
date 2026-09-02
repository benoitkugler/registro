<template>
  <v-card
    title="Formulaires du séjour"
    subtitle="Créer des formulaires personnalisés pour votre séjour, à remplir par les inscrits."
    class="ma-2 mx-auto"
    max-width="800px"
  >
    <template #append>
      <v-btn color="success" prepend-icon="mdi-plus" @click="create"
        >Créer un formulaire</v-btn
      >
    </template>
    <v-skeleton-loader v-if="data == null"></v-skeleton-loader>
    <v-card-text v-else>
      <v-list>
        <v-list-item v-if="!data.Forms?.length">
          <i>Il n'y a encore aucun formulaire pour ce séjour.</i>
        </v-list-item>
        <v-list-item
          v-for="form in data.Forms"
          :title="form.Nom"
          @click="formToEdit = form"
        >
          <template #append>
            <v-row>
              <v-col align-self="center">
                <v-badge
                  :color="
                    (data.AnswersCount || {})[form.Id] ? 'primary' : 'grey'
                  "
                  inline
                  :content="(data.AnswersCount || {})[form.Id] || 0"
                ></v-badge>
              </v-col>
              <v-col align-self="center">
                <v-menu>
                  <template #activator="{ props: innerProps }">
                    <v-btn
                      v-bind="innerProps"
                      size="x-small"
                      class="mx-1"
                      icon="mdi-dots-vertical"
                    ></v-btn>
                  </template>
                  <v-list density="compact">
                    <v-list-item
                      title="Afficher les réponses"
                      prepend-icon="mdi-view-list"
                      @click="formToLoadReponses = form"
                    ></v-list-item>

                    <v-divider thickness="1"></v-divider>
                    <v-list-item
                      title="Supprimer"
                      prepend-icon="mdi-delete"
                      @click="formToDelete = form"
                    ></v-list-item>
                  </v-list>
                </v-menu>
              </v-col>
            </v-row>
          </template>
        </v-list-item>
      </v-list>
    </v-card-text>

    <!-- reponses -->
    <v-dialog
      :model-value="formToLoadReponses != null"
      @update:model-value="formToLoadReponses = null"
    >
      <FormReponses
        :form="formToLoadReponses"
        v-if="formToLoadReponses != null"
      ></FormReponses>
    </v-dialog>

    <!-- edit form -->
    <v-dialog
      :model-value="formToEdit != null"
      @update:model-value="formToEdit = null"
      max-width="600px"
    >
      <FormEditor
        :form="formToEdit"
        v-if="formToEdit != null"
        @save="updateForm"
      ></FormEditor>
    </v-dialog>

    <!-- delete form -->
    <v-dialog
      :model-value="formToDelete != null"
      @update:model-value="formToDelete = null"
      max-width="400px"
    >
      <v-card title="Confirmer la suppression" v-if="formToDelete">
        <v-card-text>
          Confirmez-vous la suppression du formulaire
          <i> {{ formToDelete.Nom }} </i> ? <br />
          Les éventuelles réponses des participants seront aussi supprimées.
          <br /><br />

          Attention, cette opération est irréversible.
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="red" @click="deleteForm">Supprimer</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-card>
</template>

<script lang="ts" setup>
import { onMounted, ref } from "vue";
import { controller } from "../../logic/logic";
import type { Form, FormsOut } from "../../logic/api";
import type { DocumentsTab } from "../../plugins/router";
import FormEditor from "./FormEditor.vue";
import FormReponses from "./FormReponses.vue";

const props = defineProps<{}>();

const emit = defineEmits<{
  (e: "goTo", tab: DocumentsTab): void;
}>();

onMounted(fetchData);

const data = ref<FormsOut | null>(null);
async function fetchData() {
  const res = await controller.FormsLoad();
  if (res === undefined) return;
  data.value = res;
}

async function create() {
  if (!data.value) return;
  const res = await controller.FormsCreate();
  if (res === undefined) return;
  controller.showMessage("Formulaire créé avec succès.");
  data.value.Forms = [res].concat(...(data.value?.Forms || []));

  formToEdit.value = res; // start editing the empty form
}

const formToEdit = ref<Form | null>(null);
async function updateForm(form: Form) {
  formToEdit.value = null;
  if (!data.value) return;
  const res = await controller.FormsUpdate(form);
  if (res === undefined) return;
  controller.showMessage("Formulaire modifié avec succès.");
  const index = data.value.Forms?.findIndex((f) => f.Id == form.Id)!;
  data.value.Forms![index] = form;
}

const formToDelete = ref<Form | null>(null);
async function deleteForm() {
  const toDelete = formToDelete.value;
  formToDelete.value = null;
  if (data.value == null || toDelete == null) return;
  const res = await controller.FormsDelete({
    id: toDelete.Id,
  });
  if (res === undefined) return;
  data.value.Forms = (data.value.Forms || []).filter(
    (f) => f.Id != toDelete.Id,
  );
  controller.showMessage("Formulaire supprimé avec succès.");
}

const formToLoadReponses = ref<Form | null>(null);
</script>
