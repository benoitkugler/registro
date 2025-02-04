<template>
  <v-row>
    <v-col>
      <v-list-item
        :title="$props.camp.Camp.Nom"
        :subtitle="new Date($props.camp.Camp.DateDebut).getFullYear()"
      >
      </v-list-item>
    </v-col>
    <v-col cols="2" align-self="center">
      <v-tooltip content-class="px-1" width="600">
        <template v-slot:activator="{ isActive, props: innerProps }">
          <v-progress-linear
            v-on="{ isActive }"
            v-bind="innerProps"
            :max="$props.camp.Camp.Places"
            :model-value="$props.camp.Stats.Valides"
            height="36"
            rounded
          >
            <template v-slot:default="{ value }">
              <strong
                >{{ value }}/{{ $props.camp.Camp.Places }}
                <span v-if="allAttente > 0">
                  (+
                  {{ allAttente }})</span
                >
              </strong>
            </template></v-progress-linear
          >
        </template>
        <CampStats :stats="$props.camp.Stats"></CampStats>
      </v-tooltip>
    </v-col>
  </v-row>
</template>

<script setup lang="ts">
import { ref, computed } from "vue";
import type { CampHeader } from "@/logic/app/api";
const props = defineProps<{
  camp: CampHeader;
}>();

const allAttente = computed(
  () => props.camp.Stats.Inscriptions - props.camp.Stats.Valides
);
</script>
