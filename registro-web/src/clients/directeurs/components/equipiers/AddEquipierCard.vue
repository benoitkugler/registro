<template>
  <v-card
    title="Ajouter un équipier"
    subtitle="Sélectionner un profil existant ou créer un nouveau."
  >
    <v-card-text class="mt-4">
      <v-row>
        <v-col>
          <SelectPersonne
            v-if="!args.CreatePersonne"
            label="Personne"
            v-model="args.IdPersonne"
            initial-personne=""
            :api="{
              SelectPersonne: controller.SelectPersonne.bind(controller),
            }"
          ></SelectPersonne>
          <v-row v-else>
            <v-col>
              <v-text-field
                autofocus
                hide-details
                label="Nom"
                density="compact"
                variant="outlined"
                v-model="args.Nom"
              ></v-text-field>
            </v-col>
            <v-col>
              <v-text-field
                hide-details
                label="Prénom"
                density="compact"
                variant="outlined"
                v-model="args.Prenom"
              ></v-text-field>
            </v-col>
            <v-col cols="12">
              <v-text-field
                variant="outlined"
                density="compact"
                v-model="args.Mail"
                label="Adresse mail"
                type="email"
                :rules="[
                  FormRules.required(
                    `Une adresse mail est nécessaire pour inviter l'équipier.`
                  ),
                ]"
              ></v-text-field>
            </v-col>
          </v-row>
        </v-col>
        <v-col cols="auto" align-self="center">
          <v-btn @click="args.CreatePersonne = !args.CreatePersonne">
            {{ args.CreatePersonne ? "Sélectionner" : "Créer" }}
          </v-btn>
        </v-col>
      </v-row>
      <v-row>
        <v-col>
          <RolesField v-model="args.Roles"></RolesField>
        </v-col>
      </v-row>
    </v-card-text>
    <v-card-actions>
      <v-spacer></v-spacer>
      <v-btn :disabled="!isValid" @click="emit('create', args)">Ajouter</v-btn>
    </v-card-actions>
  </v-card>
</template>

<script lang="ts" setup>
import { computed, ref } from "vue";
import { type EquipiersCreateIn, type IdPersonne } from "../../logic/api";
import { controller } from "../../logic/logic";
import RolesField from "./RolesField.vue";
import { FormRules } from "@/utils";

// const props = defineProps<{}>();

const emit = defineEmits<{
  (e: "create", args: EquipiersCreateIn): void;
}>();

const args = ref<EquipiersCreateIn>({
  CreatePersonne: false,
  IdPersonne: 0 as IdPersonne,
  Nom: "",
  Prenom: "",
  Mail: "",
  Roles: [],
});

const isValid = computed(
  () =>
    !!(
      (args.value.CreatePersonne
        ? args.value.Nom && args.value.Prenom && args.value.Mail
        : args.value.IdPersonne) && args.value.Roles?.length
    )
);
</script>
