<template>
  <v-col align-self="center" cols="2" class="text-center">
    <ReviewStatusChip :status="props.status"></ReviewStatusChip>
  </v-col>
  <v-col cols="auto" align-self="center">
    <v-btn
      class="mx-1"
      color="green"
      :disabled="
        props.status == FileReviewStatus.FileReviewValidated || isSendingReview
      "
      @click.stop="emit('accept')"
    >
      <v-icon>mdi-check</v-icon>
      <v-tooltip activator="parent">Valider</v-tooltip>
    </v-btn>
  </v-col>
  <v-col cols="auto" align-self="center">
    <v-btn
      class="mx-1"
      color="red"
      :disabled="
        props.status == FileReviewStatus.FileReviewRejected || isSendingReview
      "
      @click.stop="emit('reject')"
    >
      <v-icon>mdi-close</v-icon>
      <v-tooltip activator="parent">Refuser</v-tooltip>
    </v-btn>
  </v-col>
</template>

<script lang="ts" setup>
import { FileReviewStatus, FileReviewStatusLabels } from "../../logic/api";
import { Formatters } from "@/utils";

const props = defineProps<{
  status: FileReviewStatus;
  isSendingReview: boolean;
}>();

const emit = defineEmits<{
  (e: "accept"): void;
  (e: "reject"): void;
}>();
</script>
