<template>
  <v-card
    class="mx-auto text-start"
    max-width="900px"
    title="Enregistrer un justificatif depuis votre téléphone"
    subtitle="Installez l'application sur Android ou utilisez un raccourci sur iPhone"
    prepend-icon="mdi-lightbulb-on-outline"
  >
    <v-card-text>
      <v-row>
        <v-col cols="12" md="6">
          <div class="d-flex align-center mb-3">
            <v-icon size="large" class="mr-2">mdi-android</v-icon>
            <div class="text-h6">Android (application web)</div>
          </div>
          <ol class="pl-5">
            <li>Ouvrez le site dans Chrome.</li>
            <li>
              Choisissez « Ajouter à l'écran d'accueil » ou « Installer
              l'application ».
            </li>
            <li>Connectez-vous depuis l'icône installée.</li>
            <li>
              Depuis Photos ou Fichiers, sélectionnez un justificatif puis
              <b>Partager</b> → <b>Comptabilité des séjours</b>.
            </li>
            <li>
              Le fichier s'ouvre sur l'écran des enregistrements : vous pouvez
              créer un nouvel enregistrement ou l'associer à un enregistrement
              existant.
            </li>
          </ol>
        </v-col>

        <v-col cols="12" md="6">
          <div class="d-flex align-center mb-3">
            <v-icon size="large" class="mr-2">mdi-apple</v-icon>
            <div class="text-h6">iPhone (Raccourcis Apple)</div>
          </div>
          <ol class="pl-5">
            <li>
              Copiez l'adresse de téléversement ci-dessous : elle sera demandée
              lors de l'installation du raccourci.
              <div class="mt-2 mb-2 text-center">
                <v-btn
                  variant="tonal"
                  prepend-icon="mdi-content-copy"
                  :disabled="!directUploadLink"
                  @click="copyUploadURL"
                >
                  Copier l'adresse de téléversement
                </v-btn>
              </div>
            </li>
            <li>
              Installez le raccourci :
              <div class="mt-2 mb-2 text-center">
                <v-btn
                  :href="isAppleDevice ? APPLE_SHORTCUT_URL : undefined"
                  target="_blank"
                  color="primary"
                  prepend-icon="mdi-download"
                  :disabled="!isAppleDevice"
                >
                  Installer le raccourci
                </v-btn>
              </div>
            </li>

            <li>
              Depuis Photos ou Fichiers, partagez un justificatif avec le
              raccourci <b>Comptabilité des séjours</b>: il est enregistré
              directement comme nouvel enregistrement.
            </li>
          </ol>
        </v-col>
      </v-row>
    </v-card-text>
  </v-card>
</template>

<script lang="ts" setup>
import { computed } from "vue";
import { controller } from "../../logic/logic";

const props = defineProps<{
  directUploadLink: string;
}>();

// Shortcut: "Comptabilité des séjours" (Share Sheet / action extension)
// In: image or PDF (errors with "Aucun fichier à téléverser" if none).
// 1. Convert input to image.
// 2. UploadURL = text entered at import ("URL de téléversement").
// 3. POST multipart/form-data to UploadURL, field "justificatifs" = image.
// 4. Read "RedirectURL" from JSON response, then open it.
// Server contract: accept multipart field "justificatifs"; reply {"RedirectURL": "..."}.
const APPLE_SHORTCUT_URL =
  "https://www.icloud.com/shortcuts/5928299390a44a929675a96ed8c04123";

// iOS/iPadOS/macOS only: the Shortcuts link can't be installed elsewhere.
// iPadOS 13+ reports itself as "MacIntel", so also check for touch support.
const isAppleDevice = computed(() => {
  if (typeof navigator === "undefined") return false;
  const ua = navigator.userAgent;
  return (
    /iPhone|iPad|iPod|Macintosh/.test(ua) ||
    (navigator.platform === "MacIntel" && navigator.maxTouchPoints > 1)
  );
});

async function copyUploadURL() {
  if (!props.directUploadLink) return;
  try {
    await navigator.clipboard.writeText(props.directUploadLink);
  } catch {
    console.log(props.directUploadLink); // required on localhost in dev mode
    controller.showMessage("Impossible de copier l'adresse de dépôt.", "error");
    return;
  }
  controller.showMessage("Adresse de dépôt copiée dans le presse-papier.");
}
</script>
