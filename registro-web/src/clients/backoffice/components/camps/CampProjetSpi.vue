<template>
  <v-card
    title="Projet spirituel"
    :subtitle="
      data
        ? `Par ${Personnes.label(data.Directeur)} - ${data.Directeur.Mail}`
        : ''
    "
  >
    <v-card-text>
      <v-skeleton-loader v-if="data == null"></v-skeleton-loader>
      <ProjetSpiFields
        v-else
        :model-value="data.Projet"
        readonly
      ></ProjetSpiFields>
    </v-card-text>
  </v-card>
</template>

<script setup lang="ts">
import { onMounted, ref } from "vue";
import type { CampHeader, ProjetSpiOut } from "../../logic/api";
import { controller } from "../../logic/logic";
import { Personnes } from "@/utils";

const props = defineProps<{
  camp: CampHeader;
}>();

// const emit = defineEmits<{}>();

onMounted(fetchData);

const data = ref<ProjetSpiOut | null>(null);
async function fetchData() {
  const res = await controller.CampsLoadProjetSpi({
    idCamp: props.camp.Camp.Camp.Id,
  });
  if (res === undefined) return;
  data.value = res;
}
</script>
