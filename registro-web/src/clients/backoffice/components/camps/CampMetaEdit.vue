<template>
  <v-card
    title="Modifier les métadonnées"
    subtitle="Ces données sont exposées sur l'API publique des camps ouverts aux inscriptions."
  >
    <template #append>
      <v-btn size="small" class="ml-2" @click="inner.push(['', ''])">
        <template #prepend>
          <v-icon color="green">mdi-plus</v-icon>
        </template>
        Ajouter une valeur</v-btn
      >
    </template>
    <v-card-text class="pt-2">
      <v-row v-for="(entry, index) in inner">
        <v-col cols="4">
          <v-text-field
            variant="outlined"
            density="compact"
            hide-details
            label="Clé"
            v-model="entry[0]"
          ></v-text-field>
        </v-col>
        <v-col>
          <v-text-field
            variant="outlined"
            density="compact"
            hide-details
            label="Valeur"
            v-model="entry[1]"
          ></v-text-field>
        </v-col>
        <v-col cols="auto">
          <v-btn icon size="small" flat @click="inner.splice(index, 1)">
            <v-icon color="red">mdi-close</v-icon>
          </v-btn>
        </v-col>
      </v-row>
    </v-card-text>
    <v-card-actions>
      <v-spacer></v-spacer>
      <v-btn :disabled="isDisabled" @click="emitSave"> Valider</v-btn>
    </v-card-actions>
  </v-card>
</template>

<script setup lang="ts">
import type { Meta } from "@/clients/backoffice/logic/api";
import { computed, ref } from "vue";
const props = defineProps<{
  meta: Meta;
}>();

const emit = defineEmits<{
  (e: "save", meta: NonNullable<Meta>): void;
}>();

const inner = ref(Object.entries(props.meta || {}));

function emitSave() {
  const m = Object.fromEntries(inner.value);
  emit("save", m);
}

const isDisabled = computed(() => {
  const keysL = inner.value.map((entry) => entry[0]);
  if (keysL.includes("")) return true;
  const keysS = new Set(keysL);
  return keysS.size != inner.value.length;
});
</script>
