<template>
  <v-card title="Editer les paramètres du séjour">
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
            <v-select
              label="Ouverture aux inscriptions"
              variant="outlined"
              density="compact"
              :items="campStatutItems"
              hide-details
              v-model="inner.Statut"
            ></v-select>
          </v-col>
        </v-row>
        <v-row>
          <v-col align-self="center" cols="2">
            <DateField
              v-model="inner.DateDebut"
              label="Date de début"
              hide-details
            ></DateField>
          </v-col>
          <v-col align-self="center" cols="2">
            <IntField
              label="Durée (en jours)"
              v-model="inner.Duree"
              hide-details
            ></IntField>
          </v-col>
          <v-col align-self="center" cols="2">
            <DateField
              :model-value="Camps.dateFin(inner)"
              label="Date de fin"
              readonly
              hide-details
            ></DateField>
          </v-col>
          <v-col align-self="center" cols="3">
            <MontantField
              label="Prix"
              v-model="inner.Prix"
              hide-details
            ></MontantField>
          </v-col>
          <v-col align-self="center" cols="3" class="text-center">
            <v-menu :close-on-content-click="false">
              <template #activator="{ props: menuProps }">
                <v-chip v-bind="menuProps" elevation="1" label>
                  <div
                    class="mx-1"
                    v-if="
                      !Camps.isQuotientFamilialActive(
                        inner.OptionQuotientFamilial
                      ) && inner.OptionPrix.Active == OptionPrixKind.NoOption
                    "
                  >
                    Aucune option
                  </div>
                  <div
                    class="mx-1"
                    v-if="
                      Camps.isQuotientFamilialActive(
                        inner.OptionQuotientFamilial
                      )
                    "
                  >
                    Quotient familial
                  </div>
                  <div
                    class="mx-1"
                    v-if="inner.OptionPrix.Active != OptionPrixKind.NoOption"
                  >
                    {{ OptionPrixKindLabels[inner.OptionPrix.Active] }}
                  </div>
                </v-chip>
              </template>

              <CampOptionsPrix
                :camp="inner"
                v-model:option-prix="inner.OptionPrix"
                v-model:option-qf="inner.OptionQuotientFamilial"
              ></CampOptionsPrix>
            </v-menu>
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
            <BoolField
              v-model="inner.NeedEquilibreGF"
              label="Equilibre G/F demandé"
              density="compact"
              persistent-hint
              hint="Place un inscrit en liste d'attente en cas de déséquilibre Garçons/Filles"
            ></BoolField>
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
            <BoolField
              label="Navette"
              v-model="inner.Navette.Actif"
            ></BoolField>
          </v-col>
          <v-col>
            <v-textarea
              variant="outlined"
              density="compact"
              label="Navette"
              hint="Ce texte est affiché sur l'espace personnel de suivi."
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
              label="Image"
              density="compact"
              variant="outlined"
              hint="Lien vers une image affichée sur le formulaire d'inscription"
              persistent-hint
              v-model="inner.ImageURL"
            ></v-text-field>
          </v-col>
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
import {
  OptionPrixKind,
  OptionPrixKindLabels,
  StatutCampLabels,
  type Camp,
} from "@/clients/backoffice/logic/api";
import { Camps, copy, selectItems } from "@/utils";
import CampOptionsPrix from "./CampOptionsPrix.vue";
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
    inner.value.Places >= 1 &&
    inner.value.Nom != "" &&
    inner.value.AgeMax >= inner.value.AgeMin &&
    inner.value.OptionQuotientFamilial.every((p) => 0 <= p && p <= 100) &&
    !(
      inner.value.OptionPrix.Active == OptionPrixKind.PrixStatut &&
      !inner.value.OptionPrix.Statuts?.length
    ) &&
    !(
      inner.value.OptionPrix.Active == OptionPrixKind.PrixJour &&
      inner.value.OptionPrix.Jours?.length != inner.value.Duree
    )
);

const campStatutItems = selectItems(StatutCampLabels);
</script>
