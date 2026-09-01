<template>
  <v-card :subtitle="`Champ - ${kind}`">
    <template #append>
      <v-btn
        color="red"
        icon="mdi-delete"
        size="small"
        @click="emit('delete')"
      ></v-btn>
    </template>
    <v-card-text>
      <v-row>
        <v-col>
          <v-text-field
            label="Titre du champ"
            density="compact"
            variant="outlined"
            v-model="champ.Titre"
            hide-details
          ></v-text-field>
        </v-col>
      </v-row>
      <v-row>
        <v-col>
          <v-textarea
            label="Description du champ"
            density="compact"
            variant="outlined"
            v-model="champ.Description"
            hide-details
          ></v-textarea> </v-col
      ></v-row>
      <template v-if="champ.Question.Kind == 'ChampTexte'">
        <v-row>
          <v-col>
            <v-checkbox
              density="compact"
              v-model="champ.Question.Data.MultiLignes"
              label="Réponse sur plusieurs lignes"
              hint="Cocher si la réponse attendue est longue."
            ></v-checkbox>
          </v-col>
        </v-row>
      </template>
      <template v-else-if="champ.Question.Kind == 'ChampQCM'">
        <v-row>
          <v-col>
            <v-checkbox
              v-model="champ.Question.Data.Multiple"
              label="Permettre plusieurs réponses"
              hide-details
            ></v-checkbox>
          </v-col>
        </v-row>
        <v-row>
          <v-col>
            <v-list>
              <v-row justify="space-between" no-gutters>
                <v-col align-self="center">
                  <v-list-subheader> Choix possibles </v-list-subheader>
                </v-col>
                <v-col align-self="center" cols="auto">
                  <v-btn
                    color="success"
                    prepend-icon="mdi-plus"
                    size="small"
                    @click="addChoice"
                  >
                    Ajouter un choix
                  </v-btn>
                </v-col>
              </v-row>
              <v-list-item
                v-for="(proposition, i) in champ.Question.Data.Propositions"
                :title="proposition"
              >
                <template
                  #title
                  v-if="propositionToEdit && propositionToEdit[1] == i"
                >
                  <v-text-field
                    autofocus
                    density="compact"
                    variant="outlined"
                    v-model="propositionToEdit[0]"
                    hide-details
                  ></v-text-field>
                </template>
                <template #append>
                  <v-row no-gutters>
                    <v-col class="ml-1">
                      <v-btn
                        v-if="propositionToEdit && propositionToEdit[1] == i"
                        icon="mdi-check"
                        size="x-small"
                        color="success"
                        @click="commitProposition"
                      >
                      </v-btn>
                      <v-btn
                        v-else
                        icon="mdi-pencil"
                        size="x-small"
                        @click="propositionToEdit = [proposition, i]"
                      >
                      </v-btn>
                    </v-col>
                    <v-col class="ml-1">
                      <v-btn
                        color="red"
                        icon="mdi-delete"
                        size="x-small"
                        @click="removeChoice(i)"
                      ></v-btn>
                    </v-col>
                  </v-row>
                </template>
              </v-list-item>
            </v-list>
          </v-col>
        </v-row>
      </template>
    </v-card-text>
  </v-card>
</template>

<script lang="ts" setup>
import { computed, ref } from "vue";
import type { Champ, Int } from "../../logic/api";

const props = defineProps<{}>();

const emit = defineEmits<{
  (e: "delete"): void;
}>();

const champ = defineModel<Champ>({ required: true });

const kind = computed(() => {
  switch (champ.value.Question.Kind) {
    case "ChampQCM":
      return "QCM";
    case "ChampTexte":
      return "Texte libre";
  }
});

function addChoice() {
  if (champ.value.Question.Kind != "ChampQCM") return;
  const d = champ.value.Question.Data;
  d.Propositions = (d.Propositions || []).concat("Autre choix");
  const i = d.Propositions.length - 1;
  propositionToEdit.value = [d.Propositions[i], i];
}

function removeChoice(i: number) {
  if (champ.value.Question.Kind != "ChampQCM") return;
  champ.value.Question.Data.Propositions!.splice(i, 1);
}

const propositionToEdit = ref<[string, number] | null>(null);
function commitProposition() {
  if (champ.value.Question.Kind != "ChampQCM" || !propositionToEdit.value)
    return;
  const [value, i] = propositionToEdit.value;
  champ.value.Question.Data.Propositions![i] = value;
  propositionToEdit.value = null;
}
</script>
