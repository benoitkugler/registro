<template>
  <v-chip
    size="small"
    :color="Formatters.reviewStatus(props.status).color"
    :prepend-icon="Formatters.reviewStatus(props.status).icon"
  >
    {{ FileReviewStatusLabels[props.status] }}
  </v-chip>
</template>

<script lang="ts" setup>
import {
  FileReviewStatus,
  FileReviewStatusLabels,
} from "@/clients/backoffice/logic/api";
import { Formatters } from "@/utils";

const props = defineProps<{
  status: FileReviewStatus;
}>();
</script>
