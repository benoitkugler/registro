<template>
  <v-card class="my-2" :title="props.champ.Titre">
    <v-card-text>
      <MultilineText
        class="mb-4"
        :text="props.champ.Description"
      ></MultilineText>
      <template v-if="props.champ.Question.Kind == 'ChampTexte'">
        <v-textarea
          variant="outlined"
          density="compact"
          v-if="props.champ.Question.Data.MultiLignes"
          label="Votre réponse"
          v-model="reponse.Data"
          hide-details
        ></v-textarea>
        <v-text-field
          variant="outlined"
          density="compact"
          label="Votre réponse"
          v-model="reponse.Data"
          hide-details
          v-else
        ></v-text-field>
      </template>
      <template v-else>
        <v-select
          v-if="props.champ.Question.Data.Multiple"
          density="compact"
          variant="outlined"
          label="Votre réponse"
          :items="
            (props.champ.Question.Data.Propositions || []).map((p, i) => ({
              title: p,
              value: i,
            }))
          "
          v-model="reponse.Data as ChampReponseQCM"
          multiple
          hide-details
        ></v-select>
        <v-select
          v-else
          density="compact"
          variant="outlined"
          label="Votre réponse"
          :items="
            (props.champ.Question.Data.Propositions || []).map((p, i) => ({
              title: p,
              value: i,
            }))
          "
          :multiple="false"
          :model-value="reponse.Data?.length ? reponse.Data[0] : null"
          @update:model-value="(v) => (reponse.Data = [v as Int])"
          :rules="[(v) => (v == null ? 'Champ requis' : true)]"
        >
        </v-select>
      </template>
    </v-card-text>
  </v-card>
</template>

<script setup lang="ts">
import type { Champ, ChampReponse, ChampReponseQCM, Int } from "../logic/api";

const props = defineProps<{
  champ: Champ;
}>();

const emit = defineEmits<{
  (e: "accept"): void;
}>();

const reponse = defineModel<ChampReponse>("reponse", { required: true });
</script>
