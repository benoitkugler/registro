<template>
  <v-card :title="props.title">
    <v-card-text>
      <v-progress-linear
        :max="props.total"
        :model-value="props.current"
        color="primary"
        height="24"
      >
        {{ props.current }} / {{ props.total }}
      </v-progress-linear>
    </v-card-text>
  </v-card>
</template>

<script setup lang="ts">
import { type Int } from "@/clients/backoffice/logic/api";

const props = defineProps<{
  title: string;
  current: Int;
  total: Int;
}>();
</script>
