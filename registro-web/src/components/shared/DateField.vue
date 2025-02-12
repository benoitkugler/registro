<template>
  <v-text-field
    v-if="props.readonly"
    class="center-text"
    variant="outlined"
    density="compact"
    :hide-details="props.hideDetails"
    :label="props.label"
    :model-value="(adapter.parseISO(modelValue) as Date).toLocaleDateString()"
    readonly
  >
  </v-text-field>
  <v-menu v-else :close-on-content-click="false" v-model="showMenu">
    <template v-slot:activator="{ props: innerProps }">
      <v-text-field
        class="center-text"
        v-bind="innerProps"
        variant="outlined"
        density="compact"
        :hide-details="props.hideDetails"
        :label="props.label"
        :model-value="(adapter.parseISO(modelValue) as Date).toLocaleDateString()"
      >
      </v-text-field>
    </template>

    <v-date-picker
      :model-value="adapter.parseISO(modelValue)"
      @update:model-value="d => onUpdate(d as Date)"
    ></v-date-picker>
  </v-menu>
</template>

<script setup lang="ts">
import type { Date_, Int } from "@/logic/backoffice/api";
import { newDate_ } from "@/logic/backoffice/logic";
import { ref } from "vue";
import { useDate } from "vuetify";
const props = defineProps<{
  label: string;
  hideDetails?: boolean;
  readonly?: boolean;
}>();

const modelValue = defineModel<Date_>({ required: true });

const adapter = useDate();

function onUpdate(d: Date) {
  showMenu.value = false;
  modelValue.value = newDate_(d);
}

const showMenu = ref(false);
</script>

<style scoped>
.center-text :deep(input) {
  text-align: center;
}
</style>
