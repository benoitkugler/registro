<template>
  <v-card
    title="Formulaire à remplir"
    :subtitle="`${props.form.Form.Nom} - ${props.form.Personne}`"
  >
    <v-card-text>
      <div class="mb-2">
        <MultilineText :text="props.form.Form.Introduction"></MultilineText>
      </div>

      <FormChamp
        v-for="(champ, i) in props.form.Form.Champs"
        :champ="champ"
        v-model:reponse="reponses[i]"
      ></FormChamp>
    </v-card-text>
    <v-card-actions>
      <v-btn
        color="green"
        block
        variant="outlined"
        prepend-icon="mdi-content-save"
        :disabled="!areReponsesValid"
        @click="emit('save', reponses)"
        >Enregistrer mes réponses</v-btn
      >
    </v-card-actions>
  </v-card>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import type {
  ChampReponse,
  ChampReponseQCM,
  ChampReponseTexte,
  FormParticipant,
  FormReponses,
} from "../logic/api";
import { copy } from "@/utils";
import FormChamp from "./FormChamp.vue";

const props = defineProps<{
  form: FormParticipant;
}>();

const emit = defineEmits<{
  (e: "save", reponses: FormReponses): void;
}>();

const reponses = ref<NonNullable<FormReponses>>(ensureAnswers());

function ensureAnswers() {
  if (props.form.Reponses?.length == props.form.Form.Champs?.length) {
    return copy(props.form.Reponses || []);
  }
  // otherwise, this is a new reponse, build the correct types
  return (props.form.Form.Champs || []).map((c): ChampReponse => {
    switch (c.Question.Kind) {
      case "ChampQCM":
        return { Kind: "ChampReponseQCM", Data: [] satisfies ChampReponseQCM };
      case "ChampTexte":
        return {
          Kind: "ChampReponseTexte",
          Data: "" satisfies ChampReponseTexte,
        };
    }
  });
}

const areReponsesValid = computed(() => {
  return reponses.value.every((r, i) => {
    const champ = props.form.Form.Champs![i];
    switch (champ.Question.Kind) {
      case "ChampQCM":
        return (
          champ.Question.Data.Multiple || !!(r.Data as ChampReponseQCM)?.length
        );
      case "ChampTexte":
        return true;
    }
  });
});
</script>
