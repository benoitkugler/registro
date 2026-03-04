<template>
  <v-card :title="props.title">
    <v-card-text>
      <v-progress-linear
        :max="props.progress.Total"
        :model-value="props.progress.Current"
        color="primary"
        height="24"
      >
        {{ props.progress.Current }} / {{ props.progress.Total }}
      </v-progress-linear>
    </v-card-text>
  </v-card>
</template>

<script setup lang="ts">
import { type SendProgress } from "@/clients/backoffice/logic/api";

const props = defineProps<{
  title: string;
  progress: SendProgress;
}>();
</script>
