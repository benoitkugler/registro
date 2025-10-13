<template>
  <v-card subtitle="Métadonnées exposées sur l'API publique">
    <template #append>
      <v-btn
        size="small"
        class="ml-2"
        @click="
          inner.push(['', '']);
          emitSave();
        "
      >
        <template #prepend>
          <v-icon color="green">mdi-plus</v-icon>
        </template>
        Ajouter une valeur</v-btn
      >
    </template>
    <v-card-text class="pt-2">
      <v-row v-for="(entry, index) in inner">
        <v-col cols="4">
          <v-combobox
            variant="outlined"
            density="compact"
            hide-details
            label="Clé"
            v-model="entry[0]"
            @update:model-value="emitSave"
            autofocus
            :items="props.metaEntriesHints.keys"
          ></v-combobox>
        </v-col>
        <v-col>
          <v-combobox
            variant="outlined"
            density="compact"
            hide-details
            label="Valeur"
            v-model="entry[1]"
            @update:model-value="emitSave"
            :items="props.metaEntriesHints.values"
          ></v-combobox>
        </v-col>
        <v-col cols="auto">
          <v-btn
            icon
            size="small"
            flat
            @click="
              inner.splice(index, 1);
              emitSave();
            "
          >
            <v-icon color="red">mdi-close</v-icon>
          </v-btn>
        </v-col>
      </v-row>
    </v-card-text>
  </v-card>
</template>

<script setup lang="ts">
import type { Meta } from "@/clients/backoffice/logic/api";
import { computed, ref } from "vue";
const props = defineProps<{
  metaEntriesHints: { keys: string[]; values: string[] };
}>();

const meta = defineModel<Meta>({ required: true });

// const emit = defineEmits<{}>();

const inner = ref(Object.entries(meta.value || {}));

function emitSave() {
  const m = Object.fromEntries(inner.value);
  meta.value = m;
}

const isDisabled = computed(() => {
  const keysL = inner.value.map((entry) => entry[0]);
  if (keysL.includes("")) return true;
  const keysS = new Set(keysL);
  return keysS.size != inner.value.length;
});
</script>
