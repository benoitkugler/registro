<template>
  <v-tooltip :text="props.text">
    <template v-slot:activator="{ props: innerProps }">
      <v-icon v-bind="innerProps" :class="props.class"
        >mdi-help-circle-outline</v-icon
      >
    </template>
  </v-tooltip>
</template>

<script lang="ts" setup>
const props = defineProps<{
  text: string;
  class?: string;
}>();
</script>
