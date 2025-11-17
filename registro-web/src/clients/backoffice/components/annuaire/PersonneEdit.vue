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
          <v-col cols="3">
            <SexeField v-model="inner.Sexe"></SexeField>
          </v-col>
        </v-row>
        <v-row>
          <v-col>
            <DateNaissanceField
              v-model="inner.DateNaissance"
            ></DateNaissanceField>
          </v-col>
          <v-col>
            <v-text-field
              hide-details
              label="Ville de naissance"
              density="compact"
              variant="outlined"
              v-model="inner.VilleNaissance"
            ></v-text-field>
          </v-col>
          <v-col>
            <DepartementField
              label="Département de naissance"
              v-model="inner.DepartementNaissance"
            ></DepartementField>
          </v-col>
          <v-col>
            <NationaliteField v-model="inner.Nationnalite"></NationaliteField>
          </v-col>
        </v-row>
        <v-row>
          <v-col>
            <v-text-field
              hide-details
              label="Adresse"
              density="compact"
              variant="outlined"
              v-model="inner.Adresse"
            ></v-text-field>
          </v-col>
          <v-col>
            <v-text-field
              hide-details
              label="Code postal"
              density="compact"
              variant="outlined"
              v-model="inner.CodePostal"
            ></v-text-field>
          </v-col>
          <v-col>
            <v-text-field
              hide-details
              label="Ville"
              density="compact"
              variant="outlined"
              v-model="inner.Ville"
            ></v-text-field>
          </v-col>
          <v-col>
            <PaysField v-model="inner.Pays"></PaysField>
          </v-col>
        </v-row>
        <v-row>
          <v-col>
            <v-text-field
              hide-details
              label="Mail"
              density="compact"
              variant="outlined"
              v-model="inner.Mail"
            ></v-text-field>
          </v-col>
          <v-col>
            <StringList
              label="Téléphone"
              v-model="inner.Tels"
              :formatter="Formatters.tel"
            ></StringList>
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
import { copy, Formatters } from "@/utils";
import StringList from "@/components/StringList.vue";
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
