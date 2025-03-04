<template>
  <!-- fix some sizeinf issue with the icon prop -->
  <v-timeline-item
    :dot-color="props.color"
    width="100%"
    :size="props.size || 'medium'"
    :elevation="2"
  >
    <template #icon>
      <div class="ma-2">
        <v-icon :icon="props.icon"></v-icon>
      </div>
    </template>
    <v-row justify="space-between" no-gutters>
      <v-col align-self="center">
        <slot></slot>
      </v-col>
      <v-col cols="2" class="text-grey ml-4">
        {{ Formatters.time(props.time) }}
      </v-col>
    </v-row>
  </v-timeline-item>
</template>

<script setup lang="ts">
import type { Time } from "@/clients/backoffice/logic/api";
import { Formatters } from "@/utils";

const props = defineProps<{
  color: string;
  icon: string;
  time: Time;
  size?: string;
}>();
</script>

<style scoped></style>
