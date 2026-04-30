<template>
  <v-row>
    <v-col align-self="center" cols="6" md="4">
      {{ props.label }}
    </v-col>
    <template v-if="!enabled">
      <v-col class="text-center">
        <v-chip variant="outlined" prepend-icon="mdi-check" color="green">
          {{ props.textDisabled }}
        </v-chip>
      </v-col>
      <v-col cols="auto">
        <v-btn size="small" prepend-icon="mdi-plus" @click="enabled = true"
          >Déclarer</v-btn
        >
      </v-col>
    </template>

    <v-col v-else cols="12" md="8">
      <v-textarea
        :label="props.label"
        rows="3"
        variant="outlined"
        v-model="model"
        @update:model-value="
          (v) => {
            if (!v) enabled = false;
          }
        "
        :hint="props.hint"
        persistent-hint
        :hide-details="!props.hint"
        :disabled="!enabled"
        clearable
      ></v-textarea>
      <v-fade-transition v-if="props.alertOnEnabled">
        <v-alert
          :model-value="model != ''"
          color="blue-lighten-4"
          type="warning"
        >
          {{ props.alertOnEnabled }}
        </v-alert>
      </v-fade-transition>
    </v-col>
  </v-row>
</template>

<script setup lang="ts">
import { ref } from "vue";
const props = defineProps<{
  textDisabled: string;
  label: string;
  hint: string;
  alertOnEnabled?: string;
}>();

const model = defineModel<string>({ required: true });

const enabled = ref(model.value != "");
</script>
