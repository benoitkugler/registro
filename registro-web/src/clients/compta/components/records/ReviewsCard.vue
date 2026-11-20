<template>
  <div class="my-2">
    <div v-if="!props.reviews?.length" class="text-center mb-4">
      <i>Aucun message.</i>
    </div>
    <v-row
      no-gutters
      v-for="(review, i) in props.reviews"
      :key="i"
      :class="{ 'rounded-lg': true, [color(review)]: true, 'my-2': true }"
    >
      <v-col align-self="center" cols="9" class="pa-2">
        <div v-for="(line, i) in review.Message.split('\n')" :key="i">
          {{ line }}
        </div>
      </v-col>
      <v-col align-self="center" cols="3" class="pa-2 font-italic text-right">{{
        Formatters.time(review.Time)
      }}</v-col>
    </v-row>
  </div>
</template>

<script lang="ts" setup>
import { Formatters } from "@/utils";
import type { Review, Reviews } from "../../logic/api";

const props = defineProps<{
  reviews: Reviews;
}>();

const color = (review: Review) =>
  review.IsAdmin ? "bg-teal-lighten-4" : "bg-grey-lighten-4";
</script>
