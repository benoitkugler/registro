<template>
  <v-card
    title="Documents du séjour"
    subtitle="Documents à lire ou remplir par les familles"
    class="ma-2"
  >
    <v-skeleton-loader v-if="data == null"></v-skeleton-loader>
    <v-card-text v-else>
      <v-list>
        <!-- generated documents -->
        <v-list-item
          title="Lettre aux familles"
          subtitle="Document officiel pour le premier contact"
        ></v-list-item>
        <v-list-item title="Liste de vêtements"></v-list-item>
        <v-list-item
          title="Liste des participants"
          subtitle="Permet d'organiser le co-voiturage"
        ></v-list-item>
      </v-list>

      {{ data }}
    </v-card-text>
  </v-card>
</template>

<script lang="ts" setup>
import { onMounted, ref } from "vue";
import { controller } from "../../logic/logic";
import type { DocumentsOut } from "../../logic/api";

const props = defineProps<{}>();

// const emit = defineEmits<{
//   (e: "save", options: LettreOptions): void;
// }>();

onMounted(fetchData);

const data = ref<DocumentsOut | null>(null);
async function fetchData() {
  const res = await controller.DocumentsGet();
  if (res === undefined) return;
  data.value = res;
}
</script>
