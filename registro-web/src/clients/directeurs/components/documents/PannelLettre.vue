<template>
  <v-skeleton-loader type="card" v-if="data == null"></v-skeleton-loader>

  <v-card
    class="mt-2 mx-auto"
    title="Editer la lettre"
    :subtitle="subtitle"
    max-width="800px"
  >
    <template #append>
      <v-btn class="mx-2" @click="showOptions = true">
        <template #prepend>
          <v-icon>mdi-cog</v-icon>
        </template>
        Options</v-btn
      >
      <v-divider thickness="2" vertical></v-divider>
      <v-progress-circular
        v-if="isSaving"
        color="primary"
        indeterminate
        class="mx-2"
        size="42"
      ></v-progress-circular>
      <v-tooltip v-else text="Enregistrer et visualiser">
        <template #activator="{ props: tooltipProps }">
          <v-btn
            v-bind="tooltipProps"
            class="mx-2"
            icon="mdi-content-save"
            @click="saveLettre"
            :disabled="!data"
          ></v-btn>
        </template>
      </v-tooltip>
    </template>
    <v-card-text v-if="data">
      <Editor
        licenseKey="gpl"
        v-model="data.Lettre.Html"
        :init="tinyMceOptions"
      ></Editor>
    </v-card-text>
  </v-card>

  <v-dialog v-model="showOptions" max-width="800px">
    <LettreOptionsCard
      v-if="data"
      :options="{
        UseCoordCentre: data.Lettre.UseCoordCentre,
        ShowAdressePostale: data.Lettre.ShowAdressePostale,
        ColorCoord: data.Lettre.ColorCoord,
      }"
      @save="saveOptions"
    ></LettreOptionsCard>
  </v-dialog>

  <!-- preview PDF -->
  <v-dialog v-model="showPreview">
    <object
      v-if="showPreview"
      type="application/pdf"
      :data="urlPreviewPDF"
      style="height: 95vh"
    ></object>
  </v-dialog>
</template>

<script lang="ts" setup>
import { computed, onMounted, ref } from "vue";
import {
  controller,
  type LettreOptions,
} from "@/clients/directeurs/logic/logic";
import type { LettreOut } from "@/clients/directeurs/logic/api";
import { endpoints, Formatters } from "@/utils";

/* Import TinyMCE */
import "tinymce";

/* Default icons are required. After that, import custom icons if applicable */
import "tinymce/icons/default/icons.min.js";

/* Required TinyMCE components */
import "tinymce/themes/silver/theme.min.js";
import "tinymce/models/dom/model.min.js";

/* Import a skin (can be a custom skin instead of the default) */
import "tinymce/skins/ui/oxide/skin.js";

/* Import plugins */
// "lists", "advlist", "image", "code", "link"
import "tinymce/plugins/lists";
import "tinymce/plugins/advlist";
import "tinymce/plugins/image";
import "tinymce/plugins/code";
import "tinymce/plugins/link";

/* content UI CSS is required */
import "tinymce/skins/ui/oxide/content.js";

/* The default content CSS can be changed or replaced with appropriate CSS for the editor content. */
import "tinymce/skins/content/default/content.js";

import "@/clients/directeurs/plugins/tinymce_fr_FR";

import Editor from "@tinymce/tinymce-vue";
import type { EditorOptions } from "tinymce";
import LettreOptionsCard from "./LettreOptionsCard.vue";

onMounted(fetchLettre);

const data = ref<LettreOut | null>(null);

async function fetchLettre() {
  const res = await controller.LettreGet();
  if (res === undefined) return;
  // TODO: https://github.com/benoitkugler/registro/issues/61
  // show info as popup on first use
  if (res.File.Id == 0) {
    // ajout pavé Espace personnel pour liste vide
    if (res.Lettre.Html == "") {
      res.Lettre.Html = paveEspacePerso;
    }
  }
  data.value = res;
}

const subtitle = computed(() => {
  const file = data.value?.File;
  if (!file) return "";
  if (file.Id == 0) return "En attente de création";
  return `Enregistrée le ${Formatters.time(
    file.Uploaded
  )} (taille : ${Formatters.size(file.Taille)})`;
});

const tinyMceOptions = computed<Partial<EditorOptions>>(() => {
  return {
    language: "fr_FR",
    height: "75vh",
    plugins: ["lists", "advlist", "image", "code", "link"],
    menubar: false,
    statusbar: false,
    paste_data_images: true,
    image_advtab: false, // style_format permet d'ajouter une marge à droite
    browser_spellcheck: true,
    contextmenu: [],
    font_formats: "Arial=arial;",
    toolbar: [
      "undo redo | cut copy paste | bold italic underline strikethrough | fontsize |  forecolor backcolor current_backcolor  |  alignleft aligncenter alignright | bullist numlist outdent indent | image styles link",
    ],
    target_list: [{ title: "Nouvelle page", value: "_blank" }],
    link_title: false,
    link_assume_external_targets: true,
    relative_urls: false,
    remove_script_host: false,
    image_description: false,
    images_upload_url: endpoints.LettreImageUpload(controller.authToken),
    automatic_uploads: true,
    formats: {
      imageMargin: { selector: "img", styles: { marginRight: "20px" } },
    },
    style_formats: [
      { title: "Espacer l'image à droite", format: "imageMargin" },
    ],
  };
});

const isSaving = ref(false);
async function saveLettre() {
  if (data.value == null) return;
  isSaving.value = true;
  const res = await controller.LettreUpdate(data.value.Lettre);
  isSaving.value = false;
  if (res === undefined) return;
  controller.showMessage("Lettre enregistrée avec succès.");
  data.value = res;
  showPreview.value = true;
}

const showOptions = ref(false);
async function saveOptions(options: LettreOptions) {
  if (!data.value) return;
  showOptions.value = false;
  data.value.Lettre.UseCoordCentre = options.UseCoordCentre;
  data.value.Lettre.ShowAdressePostale = options.ShowAdressePostale;
  data.value.Lettre.ColorCoord = options.ColorCoord;
  // build the letter again
  saveLettre();
}

const showPreview = ref(false);
const urlPreviewPDF = computed(() =>
  data.value ? endpoints.LoadDocument(data.value.File.Key) : ""
);

// Bloc à ajouter à la lettre.
// Attention, en cas de modification, vérifier la mise à page via tinmyce + htmltopdf
const paveEspacePerso = `<p><span style="font-size: 12pt; color: #3598db;" data-mce-style="font-size: 12pt; color: #3598db;">ESPACE PERSONNEL</span></p><p>Lors de votre inscription, un espace personnel de suivi vous a été attribué et un lien vers celui-ci envoyé dans le mail de confirmation (<span style="background-color: #ffffff;" data-mce-style="background-color: #ffffff;">Mon</span> <span style="background-color: #ffffff;" data-mce-style="background-color: #ffffff;">Dossier</span>). Dans cet espace vous trouverez :</p><ul><li>le <strong>suivi financier</strong> : vous pourrez alors joindre en ligne les aides auxquelles vous avez droit</li><li>les <strong>documents liés au séjour</strong> : liste de vêtement, lettre aux parents, plan d’accès au site ...</li><li>les <strong>documents à compléter</strong> ou joindre en ligne : test d'aisance aquatique si besoin ...</li><li>l'<strong>album photo</strong> du séjour</li><li>la <strong>fiche sanitaire</strong> à compléter en ligne</li></ul><p><em>TOUTES LES INFOS ET DOCUMENTS DU SÉJOUR SE TROUVENT DANS VOTRE ESPACE DÉDIÉ.</em></p>`;
</script>
