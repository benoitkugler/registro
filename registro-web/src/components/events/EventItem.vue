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
      <v-col align-self="center" cols="10">
        <slot></slot>
      </v-col>
      <v-col align-self="center" cols="2" class="text-grey pl-2">
        <small>
          {{ Formatters.time(props.time) }}
        </small>
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
