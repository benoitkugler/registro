<template>
  <v-row no-gutters>
    <v-col cols="12">
      {{ props.label }}
    </v-col>
    <v-col cols="12" class="text-center mt-1">
      <v-rating
        :hover="!props.readonly"
        color="secondary"
        :length="4"
        v-model="model"
        size="32"
        :readonly="props.readonly"
        :half-increments="props.readonly"
      ></v-rating>
    </v-col>
  </v-row>
</template>

<script lang="ts" setup generic="T extends Satisfaction | number">
import { Satisfaction } from "../../clients/espaceperso/logic/api";

const props = defineProps<{
  label: string;
  readonly?: boolean;
}>();

const model = defineModel<T>({ required: true });
</script>
