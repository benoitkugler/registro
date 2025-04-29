<template>
  <v-row no-gutters>
    <v-col cols="12">
      {{ props.label }}
    </v-col>
    <v-col cols="12">
      <v-rating
        hover
        color="primary"
        :length="4"
        v-model="model"
        size="32"
      ></v-rating>
    </v-col>
  </v-row>
</template>

<script lang="ts" setup>
import { Satisfaction } from "../logic/api";

const props = defineProps<{
  label: string;
}>();

const model = defineModel<Satisfaction>({ required: true });
</script>
