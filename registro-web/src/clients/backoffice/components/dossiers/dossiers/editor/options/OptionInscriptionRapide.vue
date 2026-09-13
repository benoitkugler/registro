<template>
  <v-card>
    <v-card-text>
      <div class="text-grey mb-4">Le prix dépend de la date d'inscription.</div>
    </v-card-text>
  </v-card>
</template>

<script setup lang="ts">
import {
  type CampItem,
  type OptionPrixCamp,
  type Time,
} from "@/clients/backoffice/logic/api";
const props = defineProps<{
  camp: CampItem;
  option: OptionPrixCamp;
}>();
</script>
