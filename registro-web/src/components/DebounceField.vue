<template>
  <v-text-field v-bind="props" v-model="inner" @update:model-value="update">
    <template v-slot:append>
      <slot name="append"></slot>
    </template>
  </v-text-field>
</template>

<script setup lang="ts">
import { ref, watch } from "vue";

const props = defineProps<{}>();

const inner = ref<string>("");

const model = defineModel<string>({ required: true });

watch(
  () => model.value,
  () => {
    inner.value = model.value;
  }
);

// debounce feature for text field
let timerId: ReturnType<typeof setTimeout>;
const debounceDelay = 300;
function update(s: string | null) {
  inner.value = s || "";

  // cancel pending call
  clearTimeout(timerId);

  // if the string is empty, immediatly trigger the update
  if (!s) {
    model.value = inner.value;
    return;
  }

  // delay new call
  timerId = setTimeout(() => {
    model.value = inner.value;
  }, debounceDelay);
}
</script>
