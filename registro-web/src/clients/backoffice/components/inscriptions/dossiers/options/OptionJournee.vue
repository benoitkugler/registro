<template>
  <v-card>
    <v-card-text>
      <div class="text-grey mb-2">
        Le prix est variable en fonction des jours de participation au camp.
      </div>

      <v-btn-toggle
        class="overflow-x-auto"
        color="secondary"
        multiple
        :model-value="!modelValue?.length ? allDays : modelValue"
        @update:model-value="(v) => (modelValue = v.length == allDays.length ? [] : v as Int[])"
      >
        <v-btn v-for="(_, i) in props.camp.OptionPrix.Jours"
          >J {{ i + 1 }}</v-btn
        >
      </v-btn-toggle>
    </v-card-text>
  </v-card>
</template>

<script setup lang="ts">
import {
  type Camp,
  type Int,
  type Jours,
} from "@/clients/backoffice/logic/api";
import { computed } from "vue";
const props = defineProps<{
  camp: Camp;
}>();

const modelValue = defineModel<Jours>({ required: true });
const allDays = computed(() =>
  (props.camp.OptionPrix.Jours || []).map((_, i) => i)
);
</script>
