<template>
  <v-card
    v-if="data != null"
    title="Fiches sanitaires et vaccins"
    max-width="1200px"
    class="mx-auto"
  >
    <template #append>
      <v-menu>
        <template #activator="{ props: menuProps }">
          <v-btn icon="mdi-download" v-bind="menuProps"></v-btn>
        </template>
        <v-list>
          <v-list-item
            title="Fiches sanitaires"
            subtitle="Document unique au format .pdf (sans les vaccins)"
          >
          </v-list-item>
          <!-- <v-list-item
            title="Fiches et vaccins"
            subtitle="Archive au format .zip"
            :href="
              endpoints.ParticipantsStreamFichesAndVaccins(controller.authToken)
            "
          >
          </v-list-item> -->
        </v-list>
      </v-menu>
    </template>
    <v-card-text>
      <v-table>
        <tr v-for="participant in data" :title="participant.Personne">
          <td class="py-2">{{ participant.Personne }}</td>
          <td class="px-2 text-center">
            <v-chip :color="color(participant)"
              >{{ FichesanitaireStateLabels[participant.State] }}
            </v-chip>
          </td>
          <td class="px-2">
            <v-chip
              v-if="participant.Fiche.TraitementMedical"
              prepend-icon="mdi-alert"
            >
              Traitement
            </v-chip>
          </td>
          <td class="px-2">
            <!-- <v-menu v-if="allergies(participant.Fiche) != null">
              <template #activator="{ props: menuProps }">
                <v-chip v-bind="menuProps" prepend-icon="mdi-alert">
                  {{ allergies(participant.Fiche) }}
                </v-chip>
              </template>
              <v-card subtitle="Allergies et conduite à tenir">
                <v-card-text>
                  <div v-if="participant.Fiche.Allergies.Autres" class="mb-2">
                    {{ participant.Fiche.Allergies.Autres }}
                  </div>
                  <div>
                    {{ participant.Fiche.Allergies.ConduiteATenir }}
                  </div>
                </v-card-text>
              </v-card>
            </v-menu> -->
          </td>
          <td class="px-2">
            <v-menu>
              <template #activator="{ props: menuProps }">
                <v-btn
                  v-bind="menuProps"
                  icon="mdi-dots-vertical"
                  size="x-small"
                ></v-btn>
              </template>
              <v-list>
                <v-list-item
                  title="Fiche sanitaire"
                  subtitle="Télécharger au format .pdf"
                  prepend-icon="mdi-download"
                  @click="downloadOneFiche(participant.IdParticipant)"
                  :disabled="participant.State == FichesanitaireState.NoFiche"
                ></v-list-item>
                <v-divider thickness="1"></v-divider>
              </v-list>
            </v-menu>
          </td>
        </tr>
      </v-table>
    </v-card-text>
  </v-card>
  <v-skeleton-loader v-else type="card"></v-skeleton-loader>
</template>

<script lang="ts" setup>
import { ref, onMounted } from "vue";
import { controller } from "../../logic/logic";
import {
  FichesanitaireState,
  FichesanitaireStateLabels,
  type Fichesanitaire,
  type FicheSanitaireExt,
  type IdParticipant,
  type PublicFile,
} from "../../logic/api";
import { saveBlobAsFile } from "@/utils";

const props = defineProps<{}>();

onMounted(loadFiches);

const data = ref<FicheSanitaireExt[] | null>(null);
async function loadFiches() {
  const res = await controller.ParticipantsGetFichesSanitaires();
  if (res === undefined) return;
  data.value = res || [];
}

function color(f: FicheSanitaireExt) {
  switch (f.State) {
    case FichesanitaireState.NoFiche:
      return "red";
    case FichesanitaireState.Outdated:
      return "orange";
    case FichesanitaireState.UpToDate:
      return "green";
  }
}

function vaccinsLabel(l: PublicFile[] | null) {
  if (!l?.length) {
    return "Pas de vaccins";
  } else if (l.length == 1) {
    return "1 vaccin";
  } else {
    return `${l.length} vaccins`;
  }
}

// function allergies(fiche: Fichesanitaire): string | null {
//   const chunks: string[] = [];
//   if (fiche.Allergies.Alimentaires) chunks.push("alimentaires");
//   if (fiche.Allergies.Asthme) chunks.push("asthme");
//   if (fiche.Allergies.Medicamenteuses) chunks.push("médicamenteuses");
//   if (fiche.Allergies.Autres) chunks.push("autres");
//   return chunks.length ? `Allergies : ${chunks.join(", ")}` : null;
// }

async function downloadOneFiche(idParticipant: IdParticipant) {
  const res = await controller.ParticipantsDownloadFicheSanitaire({
    idParticipant,
  });
  if (res === undefined) return;
  controller.showMessage("Fiche sanitaire téléchargée avec succès.");
  saveBlobAsFile(res.blob, res.filename);
}
</script>
