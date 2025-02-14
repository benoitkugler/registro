<template>
  <v-card>
    <v-card-text>
      <v-row>
        <v-col cols="8">
          du {{ formatDate(new Date(props.camp.DateDebut)) }} au
          {{ formatDate(new Date(Camps.dateFin(props.camp))) }}</v-col
        >
        <v-col cols="4" class="text-right">
          {{ props.camp.Prix }}
        </v-col>
      </v-row>
    </v-card-text>
  </v-card>
</template>

<script lang="ts" setup>
import { formatDate } from "@/components/format";
import type { CampExt } from "../logic/api";
import { Camps } from "@/utils";

const props = defineProps<{
  camp: CampExt;
}>();

const emit = defineEmits<{
  (e: "delete"): void;
}>();
</script>
