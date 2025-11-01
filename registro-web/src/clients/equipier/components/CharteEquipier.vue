<template>
  <v-card title="Charte">
    <v-card-text>
      Ce document est destiné à présenter les
      <b>conditions d’engagement</b> avec notre association, organisateur de
      séjours chrétiens pour enfants, adolescents et adultes.

      <CharteEquipierACVE v-if="asso == 'acve'"></CharteEquipierACVE>
      <CharteEquipierRepere v-else-if="asso == 'repere'"></CharteEquipierRepere>
    </v-card-text>

    <v-card-actions>
      <v-btn color="warning" @click="emit('update', false)">
        J'émets des réserves
      </v-btn>
      <v-spacer></v-spacer>
      <v-btn color="green" @click="emit('update', true)">
        J'approuve cette charte
      </v-btn>
    </v-card-actions>
  </v-card>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { Sexe } from "../logic/api";
import { Formatters } from "@/utils";
import CharteEquipierACVE from "./CharteEquipierACVE.vue";
import CharteEquipierRepere from "./CharteEquipierRepere.vue";

const props = defineProps<{
  sexe: Sexe;
}>();

const emit = defineEmits<{
  (e: "update", accept: boolean): void;
}>();

const accord = computed(() => Formatters.accord(props.sexe));

const asso = import.meta.env.VITE_ASSO;
</script>

<style></style>
