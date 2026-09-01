<template>
  <v-card title="Configurer le formulaire" class="ma-2">
    <v-card-text>
      <v-form>
        <v-row>
          <v-col>
            <v-text-field
              autofocus
              label="Nom"
              density="compact"
              variant="outlined"
              v-model="inner.Nom"
              hide-details
            ></v-text-field>
          </v-col>
        </v-row>
        <v-row>
          <v-col>
            <v-textarea
              label="Introduction"
              density="compact"
              variant="outlined"
              v-model="inner.Introduction"
              hint="Texte affiché au début du formulaire."
              persistent-hint
            ></v-textarea> </v-col
        ></v-row>
        <v-row v-for="(champ, i) in inner.Champs">
          <v-col>
            <FormChamp
              v-model="inner.Champs![i]"
              @delete="deleteChamp(i)"
            ></FormChamp>
          </v-col>
        </v-row>
        <v-row
          ><v-col>
            <v-btn-group
              class="d-flex flex-row"
              size="small"
              color="success"
              variant="outlined"
            >
              <v-btn
                class="flex-grow-1"
                prepend-icon="mdi-form-textarea"
                @click="addTexte"
                >Ajouter un champ libre</v-btn
              >
              <v-btn
                class="flex-grow-1"
                prepend-icon="mdi-form-select"
                @click="addQCM"
                >Ajouter un QCM</v-btn
              >
            </v-btn-group>
          </v-col></v-row
        >
      </v-form>
    </v-card-text>
    <v-card-actions>
      <v-spacer></v-spacer>
      <v-btn prepend-icon="mdi-content-save" @click="emit('save', inner)"
        >Enregistrer</v-btn
      >
    </v-card-actions>
  </v-card>
</template>

<script lang="ts" setup>
import { copy } from "@/utils";
import type { Demande, DemandeDirecteur, Form } from "../../logic/api";
import { ref } from "vue";
import FormChamp from "./FormChamp.vue";

const props = defineProps<{
  form: Form;
}>();

const emit = defineEmits<{
  (e: "save", form: Form): void;
}>();

const inner = ref(copy(props.form));

function addTexte() {
  inner.value.Champs = (inner.value.Champs || []).concat({
    Titre: "",
    Description: "",
    Question: { Kind: "ChampTexte", Data: { MultiLignes: false } },
  });
}

function addQCM() {
  inner.value.Champs = (inner.value.Champs || []).concat({
    Titre: "",
    Description: "",
    Question: {
      Kind: "ChampQCM",
      Data: { Multiple: false, Propositions: ["Choix 1", "Choix 2"] },
    },
  });
}

function deleteChamp(i: number) {
  inner.value.Champs!.splice(i, 1);
}
</script>
