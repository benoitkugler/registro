<template>
  <div style="white-space: pre-wrap">
    <template v-for="(line, i) in lines">
      <br v-if="i != 0" />
      {{ line }}
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";

const props = defineProps<{
  text: string;
}>();

const lines = computed(() => props.text.split("\n"));
</script>
