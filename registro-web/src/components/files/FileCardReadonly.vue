<template>
  <v-card
    class="my-2"
    :href="
      props.isGeneratedDoc
        ? endpoints.RenderDocumentCamp(props.fileKey, false)
        : contentURL(props.fileKey)
    "
  >
    <v-img
      cover
      :src="
        props.isGeneratedDoc
          ? endpoints.RenderDocumentCamp(props.fileKey, true)
          : miniatureURL(props.fileKey)
      "
      :width="props.large ? '142px' : '48px'"
    >
      <template v-slot:placeholder>
        <div class="d-flex align-center justify-center fill-height">
          <v-progress-circular
            color="grey-lighten-4"
            indeterminate
          ></v-progress-circular>
        </div>
      </template>
    </v-img>
  </v-card>
</template>

<script setup lang="ts">
import { endpoints } from "@/clients/directeurs/logic/logic";
import { contentURL, miniatureURL } from "@/utils";

const props = defineProps<{
  fileKey: string;
  isGeneratedDoc: boolean;
  large?: boolean;
}>();
</script>

<style scoped></style>
