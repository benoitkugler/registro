<template>
  <NavBar :title="`${controller.camp?.Label} - Equipiers`"> </NavBar>

  <v-card>
    <template #append>
      <v-btn @click="showDocuments = true">
        <template #prepend>
          <v-icon>mdi-file</v-icon>
        </template>
        Documents</v-btn
      >
    </template>
    <v-card-text></v-card-text>
  </v-card>

  <v-dialog v-model="showDocuments">
    <DocumentsTable :equipiers="list"></DocumentsTable>
  </v-dialog>
</template>

<script lang="ts" setup>
import { useRouter } from "vue-router";
import NavBar from "../components/NavBar.vue";
import { computed, onMounted, ref } from "vue";
import { controller } from "../logic/logic";
import type { EquipierExt } from "../logic/api";
import DocumentsTable from "../components/equipiers/DocumentsTable.vue";

const router = useRouter();

onMounted(fetchEquipiers);

const list = ref<EquipierExt[]>([]);
async function fetchEquipiers() {
  const res = await controller.EquipiersGet();
  if (res === undefined) return;
  list.value = res || [];
}

const showDocuments = ref(false);
</script>
