<template>
  <v-card
    v-if="data != null"
    title="Documents à récupérer"
    subtitle="Documents déposés  sur les espaces de suivi"
  >
    <v-card-text>
      <v-list>
        <v-list-item
          v-for="demande in data.DemandesDocuments"
          :title="demande.Demande.Description"
        >
          <template #prepend>
            <v-menu>
              <template #activator="{ props: menuProps }">
                <v-chip
                  elevation="1"
                  v-bind="menuProps"
                  :color="colorFor(demande.UploadedBy)"
                  class="mr-4"
                >
                  {{ demande.UploadedBy?.length || 0 }} /
                  {{ Object.values(data.Personnes || {}).length }}
                </v-chip>
              </template>
              <DocumentsUploadedBy
                :personnes="data.Personnes"
                :uploaded-by="demande.UploadedBy || []"
              ></DocumentsUploadedBy>
            </v-menu>
          </template>
          <template #append>
            <v-btn
              icon="mdi-download"
              size="small"
              :href="controller.documentsStreamUploadedURL(demande.Demande.Id)"
            ></v-btn>
          </template>
        </v-list-item>
      </v-list>
    </v-card-text>
  </v-card>
  <v-skeleton-loader v-else type="card"></v-skeleton-loader>
</template>

<script lang="ts" setup>
import { ref, onMounted } from "vue";
import { controller } from "../../logic/logic";
import { type DocumentsUploadedOut, type IdPersonne } from "../../logic/api";
import DocumentsUploadedBy from "./DocumentsUploadedBy.vue";

const props = defineProps<{}>();

onMounted(loadDocuments);

const data = ref<DocumentsUploadedOut | null>(null);
async function loadDocuments() {
  const res = await controller.DocumentsGetUploaded();
  if (res === undefined) return;
  data.value = res || [];
}

function colorFor(uploadedBy: IdPersonne[] | null) {
  const got = uploadedBy?.length || 0;
  const exp = Object.values(data.value?.Personnes || {}).length;
  if (got == exp) {
    return "green";
  } else if (got == 0) {
    return "red";
  }
  return "orange";
}
</script>
