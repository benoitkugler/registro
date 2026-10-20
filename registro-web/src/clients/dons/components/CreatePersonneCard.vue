<template>
  <v-card
    title="Ajouter une personne"
    subtitle="Le profil sera ajouté à l'annuaire de l'association."
  >
    <v-card-text>
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
      </v-row>
    </v-card-text>
    <v-card-actions>
      <v-spacer></v-spacer>
      <v-btn @click="emit('save', inner)" :disabled="!areFieldsValid"
        >Enregistrer</v-btn
      >
    </v-card-actions>
  </v-card>
</template>

<script lang="ts" setup>
import { computed, ref } from "vue";
import { type PersonneDonateur } from "../logic/api";

const props = defineProps<{}>();

const emit = defineEmits<{
  (e: "save", personne: PersonneDonateur): void;
}>();

const inner = ref(newPersonneDonateur());
function newPersonneDonateur(): PersonneDonateur {
  return {
    Nom: "",
    Prenom: "",
    Mail: "",
    Adresse: "",
    CodePostal: "",
    Ville: "",
    Pays: "FR",
  };
}

const areFieldsValid = computed(() => {
  const p = inner.value;
  return (
    p.Nom != "" &&
    p.Prenom != "" &&
    p.Mail != "" &&
    p.Adresse != "" &&
    p.CodePostal != "" &&
    p.Ville != "" &&
    p.Pays != ""
  );
});
</script>
