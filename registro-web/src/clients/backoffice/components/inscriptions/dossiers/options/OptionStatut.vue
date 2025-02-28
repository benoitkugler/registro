<template>
  <v-card>
    <v-card-text>
      <div class="text-grey mb-4">Le prix dépend du statut du participant.</div>

      <v-select
        clearable
        :items="items"
        label="Statut"
        :model-value="isValid ? modelValue : null"
        @update:model-value="(v) => (modelValue = v == null ? 0 as Int : v)"
        variant="outlined"
        density="comfortable"
        hide-details
      >
        <template v-slot:item="{ item, props: menuProps }">
          <v-list-item
            v-bind="menuProps"
            :title="item.title"
            :subtitle="item.raw.subtitle"
          ></v-list-item>
        </template>
      </v-select>
    </v-card-text>
  </v-card>
</template>

<script setup lang="ts">
import { type Camp, type Int } from "@/clients/backoffice/logic/api";
import { computed, watch } from "vue";
const props = defineProps<{
  camp: Camp;
}>();

const modelValue = defineModel<Int>({ required: true });

const items = computed(() =>
  (props.camp.OptionPrix.Statuts || []).map((s) => ({
    value: s.Id,
    title: s.Label,
    subtitle: s.Description,
  }))
);

const isValid = computed(() =>
  items.value.map((item) => item.value).includes(modelValue.value)
);
</script>
