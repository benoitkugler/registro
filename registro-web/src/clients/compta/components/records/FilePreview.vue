<template>
  <v-card width="400">
    <v-card-text>
      <v-row>
        <v-col align-self="center">
          <v-list>
            <v-list-item :href="props.downloadURL" prepend-icon="mdi-download"
              >Télécharger</v-list-item
            >
            <v-list-item
              @click="emit('upload')"
              prepend-icon="mdi-upload"
              :disabled="props.readonly"
              >Remplacer</v-list-item
            >
            <v-list-item
              @click="emit('delete')"
              prepend-icon="mdi-delete"
              :disabled="props.readonly"
              >Supprimer</v-list-item
            >
          </v-list>
        </v-col>
        <v-col cols="auto" align-self="center">
          <v-img :src="props.miniatureURL" width="160" aspect-ratio="1.2">
            <template v-slot:placeholder>
              <div class="d-flex align-center justify-center fill-height">
                <v-progress-circular
                  color="grey-lighten-4"
                  indeterminate
                ></v-progress-circular>
              </div>
            </template>
          </v-img>
        </v-col>
      </v-row>
    </v-card-text>
  </v-card>
</template>

<script lang="ts" setup>
const props = defineProps<{
  downloadURL: string;
  miniatureURL: string;
  readonly: boolean;
}>();

const emit = defineEmits<{
  (e: "upload"): void;
  (e: "delete"): void;
}>();
</script>
