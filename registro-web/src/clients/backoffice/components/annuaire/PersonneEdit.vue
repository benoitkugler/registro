<template>
  <v-card title="Editer le profil">
    <v-card-text>
      <v-form>
        <v-row>
          <v-col>
            <v-text-field
              autofocus
              hide-details
              label="Nom"
              density="compact"
              variant="outlined"
              v-model="inner.Nom"
            ></v-text-field>
          </v-col>
          <v-col>
            <v-text-field
              hide-details
              label="Prénom"
              density="compact"
              variant="outlined"
              v-model="inner.Prenom"
            ></v-text-field>
          </v-col>
          <v-col>
            <SexeField v-model="inner.Sexe"></SexeField>
          </v-col>
          <v-col>
            <DateNaissanceField
              v-model="inner.DateNaissance"
            ></DateNaissanceField>
          </v-col>
        </v-row>
      </v-form>
    </v-card-text>
    <v-card-actions>
      <v-spacer></v-spacer>
      <v-btn :disabled="!areFieldsValid" @click="emit('save', inner)"
        >Enregistrer</v-btn
      >
    </v-card-actions>
  </v-card>
</template>

<script setup lang="ts">
import { ref, computed } from "vue";
import { type Personne } from "@/clients/backoffice/logic/api";
import { copy } from "@/utils";
const props = defineProps<{
  personne: Personne;
}>();
const emit = defineEmits<{
  (e: "save", personne: Personne): void;
}>();
const inner = ref(copy(props.personne));

const areFieldsValid = computed(
  () => !!(inner.value.Nom && inner.value.Prenom)
);
</script>
