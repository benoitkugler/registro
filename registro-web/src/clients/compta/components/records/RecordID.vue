<template>
  <v-tooltip
    location="top"
    :text="props.record.IsReceipe ? 'Recette' : 'Dépense'"
  >
    <template v-slot:activator="{ props: innerProps }">
      <v-chip
        v-bind="innerProps"
        :color="props.record.IsReceipe ? 'green' : 'pink'"
        :append-icon="
          props.record.IsReceipe ? 'mdi-cash-plus' : 'mdi-cash-minus'
        "
        :size="props.small ? 'small' : undefined"
        >n° {{ props.record.Id }}
      </v-chip>
    </template>
  </v-tooltip>
</template>

<script lang="ts" setup>
import type { ComptaRecord } from "../../logic/api";

const props = defineProps<{
  record: ComptaRecord;
  small?: boolean;
}>();
</script>
