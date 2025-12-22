<template>
  <NavBar :title="`${controller.camp?.Label} - Avis sur le séjour`"> </NavBar>

  <div v-if="data == null" class="text-center my-6">
    <v-progress-circular indeterminate></v-progress-circular>
  </div>
  <div class="ma-2" v-else>
    <CampSondagesV :sondages="data"></CampSondagesV>
  </div>
</template>

<script lang="ts" setup>
import { onMounted, ref } from "vue";
import NavBar from "../components/NavBar.vue";
import { controller } from "../logic/logic";
import type { CampSondages } from "../logic/api";
import CampSondagesV from "@/components/sondages/CampSondagesV.vue";

onMounted(loadData);

const data = ref<CampSondages | null>(null);
async function loadData() {
  const res = await controller.SondagesGet();
  if (res === undefined) return;

  data.value = res;
}
</script>
