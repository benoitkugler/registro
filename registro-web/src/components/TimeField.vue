<template>
  <v-row no-gutters>
    <v-col>
      <DateField
        class="mr-2"
        label="Date"
        :model-value="newDate_(new Date(modelValue))"
        @update:model-value="onUpdateDate"
        :hide-details="props.hideDetails"
        :readonly="props.readonly"
      ></DateField>
    </v-col>
    <v-col cols="auto">
      <div class="ma-2"></div>
    </v-col>
    <v-col>
      <v-menu :close-on-content-click="false" v-model="showMenu">
        <template #activator="{ props: innerProps }">
          <v-text-field
            class="center-text"
            v-bind="innerProps"
            variant="outlined"
            density="compact"
            :hide-details="props.hideDetails"
            label="Moment"
            :model-value="formatedTime"
          >
          </v-text-field>
        </template>

        <v-time-picker
          :model-value="new Date(modelValue)"
          @update:model-value="onUpdateTime"
          format="24hr"
        ></v-time-picker>
      </v-menu>
    </v-col>
  </v-row>
</template>

<script setup lang="ts">
import type { Date_, Time } from "@/clients/backoffice/logic/api";
import { computed, ref } from "vue";
import { newDate_ } from "./date";
const props = defineProps<{
  hideDetails?: boolean;
  readonly?: boolean;
}>();

const modelValue = defineModel<Time>({ required: true });

const formatedTime = computed(() =>
  new Date(modelValue.value).toLocaleTimeString(undefined, {
    hour: "numeric",
    minute: "numeric",
  })
);

function onUpdateTime(val: string) {
  // format is HH:MM
  const [hour, min] = val.split(":");
  const current = new Date(modelValue.value);
  current.setHours(Number(hour), Number(min));
  modelValue.value = current.toISOString() as Time;
}

function onUpdateDate(d: Date_) {
  const newDate = new Date(d);
  const current = new Date(modelValue.value);
  newDate.setHours(
    current.getHours(),
    current.getMinutes(),
    current.getSeconds()
  );
  modelValue.value = newDate.toISOString() as Time;
}

const showMenu = ref(false);
</script>

<style scoped>
.center-text :deep(input) {
  text-align: center;
}
</style>
