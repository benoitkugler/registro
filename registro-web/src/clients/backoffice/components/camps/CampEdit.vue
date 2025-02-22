<template>
  <v-card title="Editer le camp">
    <v-card-text>
      <v-form>
        <v-row>
          <v-col>
            <v-text-field
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
              label="Lieu"
              density="compact"
              variant="outlined"
              v-model="inner.Lieu"
            ></v-text-field>
          </v-col>
          <v-col>
            <v-text-field
              hide-details
              label="Numéro d'agrément"
              density="compact"
              variant="outlined"
              v-model="inner.Agrement"
            ></v-text-field>
          </v-col>
          <v-col>
            <v-checkbox
              label="Ouvert aux inscriptions"
              v-model="inner.Ouvert"
              hide-details
            ></v-checkbox>
          </v-col>
        </v-row>
        <v-row>
          <v-col cols="2">
            <DateField
              v-model="inner.DateDebut"
              label="Date de début"
            ></DateField>
          </v-col>
          <v-col cols="2">
            <IntField label="Durée (en jours)" v-model="inner.Duree"></IntField>
          </v-col>
          <v-col cols="2">
            <DateField
              :model-value="Camps.dateFin(inner)"
              label="Date de fin"
              readonly
            ></DateField>
          </v-col>
          <v-col cols="3">
            <MontantField
              label="Prix"
              v-model="inner.Prix"
              hide-details
            ></MontantField>
          </v-col>
          <v-col cols="3">
            <v-btn>TODO: Options sur le prix</v-btn>
          </v-col>
        </v-row>
        <v-row>
          <v-col align-self="center" cols="2">
            <IntField
              v-model="inner.Places"
              label="Nombre de places"
            ></IntField>
          </v-col>
          <v-col align-self="center" cols="2">
            <IntField v-model="inner.AgeMin" label="Âge minimum"></IntField>
          </v-col>
          <v-col align-self="center" cols="2">
            <IntField v-model="inner.AgeMax" label="Âge maximum"></IntField>
          </v-col>
          <v-col cols="6">
            <v-checkbox
              v-model="inner.NeedEquilibreGF"
              label="Equilibre G/F demandé"
              density="compact"
              persistent-hint
              hint="Place un inscrit en liste d'attente en cas de déséquilibre Garçons/Filles"
            ></v-checkbox>
          </v-col>
        </v-row>
        <v-row>
          <v-col>
            <v-textarea
              variant="outlined"
              density="compact"
              label="Description"
              hint="Ce texte est affiché sur le formulaire d'inscription."
              persistent-hint
              rows="3"
              v-model="inner.Description"
            ></v-textarea>
          </v-col>
          <v-col cols="2"> </v-col>
          <v-col cols="2">
            <v-checkbox
              label="Navette"
              v-model="inner.Navette.Actif"
            ></v-checkbox>
          </v-col>
          <v-col>
            <v-textarea
              variant="outlined"
              density="compact"
              label="Navette"
              hint="Ce texte est affiché sur le formulaire d'inscription."
              persistent-hint
              rows="3"
              :disabled="!inner.Navette.Actif"
              v-model="inner.Navette.Commentaire"
            ></v-textarea>
          </v-col>
        </v-row>
        <v-row>
          <v-col>
            <v-text-field
              label="Mot de passe"
              density="compact"
              variant="outlined"
              hint="Mot de passe d'accès à la page Directeur"
              persistent-hint
              v-model="inner.Password"
            ></v-text-field>
          </v-col>
          <v-col></v-col>
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
import type { Camp } from "@/clients/backoffice/logic/api";
import { Camps, copy } from "@/utils";
const props = defineProps<{
  camp: Camp;
}>();
const emit = defineEmits<{
  (e: "save", camp: Camp): void;
}>();
const inner = ref(copy(props.camp));

const areFieldsValid = computed(
  () =>
    new Date(inner.value.DateDebut).getFullYear() >= 2020 &&
    inner.value.Nom != "" &&
    inner.value.AgeMax >= inner.value.AgeMin
);
</script>
