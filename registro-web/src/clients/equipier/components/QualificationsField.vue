<template>
  <v-select
    label="Qualifications"
    density="compact"
    variant="outlined"
    color="primary"
    multiple
    :model-value="toArray(modelValue)"
    @update:model-value="(v) => (modelValue = fromArray(v))"
    :items="[
      { value: 'SB', title: 'SB' },
      { value: 'Canoe', title: 'Canoë' },
      { value: 'Voile', title: 'Voile' },
    ]"
  ></v-select>
</template>

<script lang="ts" setup>
import { type Qualifications } from "../logic/api";

const modelValue = defineModel<Qualifications>({ required: true });

function toArray(q: Qualifications) {
  const out = [];
  if (q.SB) out.push("SB");
  if (q.Canoe) out.push("Canoe");
  if (q.Voile) out.push("Voile");
  return out;
}
function fromArray(l: string[]): Qualifications {
  return {
    SB: l.includes("SB"),
    Canoe: l.includes("Canoe"),
    Voile: l.includes("Voile"),
  };
}
</script>
