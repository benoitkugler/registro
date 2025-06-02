<template>
  <v-card title="Liste de vêtements" class="ma-2 mx-auto" max-width="1000px">
    <template #append>
      <v-btn icon @click="addVetement" size="small" class="mx-1">
        <v-icon color="green">mdi-plus</v-icon>
      </v-btn>
      <v-menu>
        <template #activator="{ props: menuProps }">
          <v-btn v-bind="menuProps" class="mx-1" size="small"
            >Importer...</v-btn
          >
        </template>
        <v-list>
          <v-list-item
            title="Liste d'été"
            subtitle="Remplace la liste actuelle"
            @click="importDefault('Ete')"
          ></v-list-item>
          <v-list-item
            title="Liste d'hiver"
            subtitle="Remplace la liste actuelle"
            @click="importDefault('Hiver')"
          ></v-list-item>
        </v-list>
      </v-menu>
      <v-divider vertical thickness="2" class="mx-1"> </v-divider>
      <v-btn size="small" class="mx-1" @click="showComplement = true">
        <template #prepend>
          <v-icon>mdi-text-box-edit</v-icon>
        </template>
        Texte
      </v-btn>
      <v-divider vertical thickness="2" class="mx-1"> </v-divider>
      <v-btn icon @click="saveAndPreview" size="small" class="mx-1">
        <v-badge :color="isDirty ? 'pink' : 'transparent'" dot>
          <v-icon>mdi-content-save</v-icon>
        </v-badge>
      </v-btn>
    </template>
    <v-skeleton-loader v-if="data == null"></v-skeleton-loader>
    <v-card-text v-else>
      <v-list
        density="compact"
        style="max-height: 75vh"
        class="overflow-y-auto"
      >
        <v-list-item class="text-center" v-if="!data.Vetements?.length">
          <i>Aucun vêtement.</i>
        </v-list-item>
        <template v-for="(item, index) in data.Vetements">
          <!-- drop zone -->
          <DropZone
            :isDragging="isDragging"
            @on-drop="(ev) => onDrop(ev, index)"
          ></DropZone>

          <v-row v-if="indexToEdit == index" class="mx-0">
            <v-col align-self="center" cols="2">
              <IntField
                label="Quantité"
                v-model="data.Vetements![index].Quantite"
              ></IntField>
            </v-col>
            <v-col align-self="center">
              <v-text-field
                density="compact"
                variant="outlined"
                label="Description"
                hide-details
                v-model="data.Vetements![index].Description"
              ></v-text-field>
            </v-col>
            <v-col align-self="center" cols="auto">
              <v-checkbox
                density="compact"
                variant="outlined"
                label="Important"
                hide-details
                v-model="data.Vetements![index].Important"
              ></v-checkbox>
            </v-col>
            <v-col align-self="center" cols="auto">
              <v-btn size="small" @click="indexToEdit = null">
                <template #prepend>
                  <v-icon color="green">mdi-check</v-icon>
                </template>
                Terminer</v-btn
              >
            </v-col>
          </v-row>
          <v-list-item
            v-else
            :title="item.Description"
            :subtitle="item.Important ? 'Important' : ''"
          >
            <template #prepend>
              <v-badge
                inline
                :content="item.Quantite"
                color="secondary"
              ></v-badge>
            </template>
            <template #append>
              <v-icon
                class="mx-1"
                style="cursor: grab"
                :draggable="true"
                @dragstart="(ev: DragEvent) => onDragStart(ev, index)"
                @dragend="isDragging = false"
                >mdi-drag-vertical</v-icon
              >
              <v-btn
                icon="mdi-pencil"
                size="x-small"
                class="mx-1"
                @click="indexToEdit = index"
              ></v-btn>
              <v-btn icon @click="deleteVetement(index)" size="x-small">
                <v-icon color="red">mdi-close</v-icon>
              </v-btn>
            </template>
          </v-list-item>
        </template>
        <!-- drop zone -->
        <DropZone
          :isDragging="isDragging"
          @on-drop="(ev) => onDrop(ev, data?.Vetements?.length || 0)"
        ></DropZone>

        <div ref="bottom"></div>
      </v-list>
    </v-card-text>
  </v-card>

  <!-- preview PDF -->
  <v-dialog v-model="showPreview">
    <object
      v-if="showPreview"
      type="application/pdf"
      :data="urlPreviewPDF"
      style="height: 95vh"
    ></object>
  </v-dialog>

  <!-- complement -->
  <v-dialog v-model="showComplement" max-width="600px">
    <v-card
      title="Texte complémentaire"
      subtitle="Ce texte apparaît à la fin de la liste de vêtements."
    >
      <v-card-text v-if="data">
        <Editor
          licenseKey="gpl"
          v-model="data.Complement"
          :init="tinyMceOptions"
        ></Editor>
      </v-card-text>
    </v-card>
  </v-dialog>
</template>

<script lang="ts" setup>
import { computed, nextTick, onMounted, ref, useTemplateRef } from "vue";
import { controller } from "../../logic/logic";
import type { Int, ListeVetements, Vetement } from "../../logic/api";
import { DefaultListe } from "./default_liste_vetements";
import { copy, swapItems } from "@/utils";

const props = defineProps<{}>();

// const emit = defineEmits<{
//   (e: "save", options: LettreOptions): void;
// }>();

onMounted(fetchData);

const data = ref<ListeVetements | null>(null);
const lastSaved = ref<ListeVetements | null>(null);
const isDirty = computed(
  () => JSON.stringify(lastSaved.value) != JSON.stringify(data.value)
);
async function fetchData() {
  const res = await controller.VetementsGet();
  if (res === undefined) return;
  data.value = res;
  lastSaved.value = copy(res);
}

function addVetement() {
  if (!data.value) return;
  data.value.Vetements = (data.value.Vetements || []).concat({
    Description: "",
    Important: false,
    Quantite: 1 as Int,
  });
  indexToEdit.value = data.value.Vetements.length - 1;
  nextTick(() => {
    if (listBottom.value) {
      listBottom.value.scrollIntoView();
    }
  });
}

const indexToEdit = ref<number | null>(null);

function importDefault(key: keyof typeof DefaultListe) {
  if (!data.value) return;
  data.value.Vetements = copy(DefaultListe[key]);
}

function deleteVetement(index: number) {
  if (!data.value) return;
  (data.value.Vetements || []).splice(index, 1);
}

const isDragging = ref(false);

const showPreview = ref(false);
const urlPreviewPDF = computed(() => controller.listeVetementsURL());
async function saveAndPreview() {
  if (!data.value) return;
  const res = await controller.VetementsUpdate(data.value);
  if (res === undefined) return;
  lastSaved.value = copy(data.value);
  showPreview.value = true;
  controller.showMessage("Liste de vêtements enregistrée avec succès.");
}

function onDragStart(event: DragEvent, index: number) {
  isDragging.value = true;
  event.dataTransfer!.effectAllowed = "move";
  event.dataTransfer!.setData("text/json", JSON.stringify({ index }));
}

function onDrop(event: DragEvent, target: number) {
  if (!data.value) return;
  const origin = JSON.parse(event.dataTransfer!.getData("text/json")) as {
    index: number;
  };
  data.value.Vetements = swapItems(origin.index, target, data.value.Vetements!);
}

const listBottom = useTemplateRef("bottom");

/* Import TinyMCE */
import "tinymce";

/* Default icons are required. After that, import custom icons if applicable */
import "tinymce/icons/default/icons.min.js";

/* Required TinyMCE components */
import "tinymce/themes/silver/theme.min.js";
import "tinymce/models/dom/model.min.js";

/* Import a skin (can be a custom skin instead of the default) */
import "tinymce/skins/ui/oxide/skin.js";

/* content UI CSS is required */
import "tinymce/skins/ui/oxide/content.js";

/* The default content CSS can be changed or replaced with appropriate CSS for the editor content. */
import "tinymce/skins/content/default/content.js";

import "@/clients/directeurs/plugins/tinymce_fr_FR";

import Editor from "@tinymce/tinymce-vue";
import type { EditorOptions } from "tinymce";

const showComplement = ref(false);

// minimal setup
const tinyMceOptions: Partial<EditorOptions> = {
  language: "fr_FR",
  height: 150,
  menubar: false,
  statusbar: false,
  formats: {
    underline: { inline: "u" },
  },
  entity_encoding: "raw",
  font_formats: "Arial=arial,helvetica,sans-serif;",
  toolbar: ["undo redo | cut copy paste | bold italic underline "],
};
</script>
