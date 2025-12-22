<template>
  <v-row no-gutters>
    <v-col cols="10">
      {{ props.label }}
    </v-col>
    <v-col cols="2" align-self="center">
      <v-badge
        inline
        :content="formatted"
        :color="color"
        v-if="props.value != 0"
      ></v-badge>
    </v-col>
  </v-row>
</template>

<script lang="ts" setup>
import { computed } from "vue";

const props = defineProps<{
  label: string;
  value: number;
}>();

const formatted = computed(() =>
  Math.floor(props.value) == props.value
    ? props.value.toString()
    : props.value.toFixed(1)
);

const color = computed(() => {
  const ratio = (props.value - 1) / 3; // in [0;1]
  const red = 200 * (1 - ratio);
  const green = 255 * ratio;
  const blue = 0;
  return `rgb(${red},${green},${blue})`;
});
</script>
