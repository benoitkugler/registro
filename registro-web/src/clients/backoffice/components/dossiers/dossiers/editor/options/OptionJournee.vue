<template>
  <v-card>
    <v-card-text>
      <div class="text-grey mb-2">
        Le prix est variable en fonction des jours de participation au camp.
      </div>

      <v-btn-toggle
        class="overflow-x-auto"
        color="primary"
        multiple
        :model-value="!modelValue?.length ? allDays : modelValue"
        @update:model-value="
          (v) => (modelValue = v.length == allDays.length ? [] : (v as Int[]))
        "
      >
        <v-btn v-for="(_, i) in props.option.Jours">J {{ i + 1 }}</v-btn>
      </v-btn-toggle>
    </v-card-text>
  </v-card>
</template>

<script setup lang="ts">
import {
  type CampItem,
  type Int,
  type Jours,
  type OptionPrixCamp,
} from "@/clients/backoffice/logic/api";
import { computed } from "vue";
const props = defineProps<{
  camp: CampItem;
  option: OptionPrixCamp;
}>();

const modelValue = defineModel<Jours>({ required: true });
const allDays = computed(() => (props.option.Jours || []).map((_, i) => i));
</script>
