<template>
  <v-card title="Paramètres du séjour" :subtitle="Camps.label(props.camp)">
    <v-card-text>
      <v-card class="my-2" subtitle="Général">
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
          </v-row>

          <v-row>
            <v-col align-self="center">
              <DateField
                v-model="inner.DateDebut"
                label="Date de début"
                hide-details
              ></DateField>
            </v-col>
            <v-col align-self="center">
              <IntField
                label="Durée (en jours)"
                v-model="inner.Duree"
                hide-details
              ></IntField>
            </v-col>
            <v-col align-self="center">
              <DateField
                :model-value="Camps.dateFin(inner)"
                label="Date de fin"
                readonly
                hide-details
              ></DateField>
            </v-col>
          </v-row>
        </v-card-text>
      </v-card>

      <v-card class="my-2" subtitle="Prix">
        <v-card-text>
          <v-row>
            <v-col align-self="center" cols="6">
              <MontantField
                label="Prix"
                v-model="inner.Prix"
                hide-details
              ></MontantField>
            </v-col>
            <v-col align-self="center" cols="6" class="text-center">
              <v-menu :close-on-content-click="false">
                <template #activator="{ props: menuProps }">
                  <v-text-field
                    v-bind="menuProps"
                    readonly
                    label="Option sur le prix"
                    variant="outlined"
                    density="compact"
                    hide-details
                    :model-value="formatOption(inner)"
                  >
                  </v-text-field>
                </template>

                <CampOptionsPrix
                  :camp="inner"
                  v-model:option-prix="inner.OptionPrix"
                  v-model:option-qf="inner.OptionQuotientFamilial"
                ></CampOptionsPrix>
              </v-menu>
            </v-col>
          </v-row>
        </v-card-text>
      </v-card>

      <v-card class="my-2" subtitle="Visibilité et inscriptions">
        <v-card-text>
          <v-row>
            <v-col align-self="center">
              <v-select
                label="Ouverture aux inscriptions"
                variant="outlined"
                density="compact"
                :items="campStatutItems"
                hide-details
                v-model="inner.Statut"
              ></v-select>
            </v-col>
            <v-col align-self="center">
              <v-switch
                label="Camp sans inscription"
                v-model="inner.WithoutInscription"
                color="primary"
                density="compact"
                persistent-hint
                :hint="
                  inner.WithoutInscription
                    ? `Camp externe, uniquement visible sur l'API publique.`
                    : ``
                "
              ></v-switch>
            </v-col>
          </v-row>
        </v-card-text>
      </v-card>

      <v-card class="my-2" subtitle="Filtres sur les inscrits">
        <v-card-text>
          <v-row>
            <v-col align-self="center">
              <IntField
                v-model="inner.Places"
                label="Nombre de places"
              ></IntField>
            </v-col>
            <v-col align-self="center">
              <IntField v-model="inner.AgeMin" label="Âge minimum"></IntField>
            </v-col>
            <v-col align-self="center">
              <IntField v-model="inner.AgeMax" label="Âge maximum"></IntField>
            </v-col>
          </v-row>
          <v-row>
            <v-col>
              <BoolField
                v-model="inner.NeedEquilibreGF"
                label="Equilibre G/F demandé"
                density="compact"
                persistent-hint
                hint="Place un inscrit en liste d'attente en cas de déséquilibre Garçons/Filles"
              ></BoolField>
            </v-col>
          </v-row>
        </v-card-text>
      </v-card>

      <v-card class="my-2" subtitle="Formulaire d'inscription">
        <v-card-text>
          <v-row>
            <v-col>
              <v-text-field
                label="Image"
                density="compact"
                variant="outlined"
                hint="URL d'une image affichée sur le formulaire d'inscription"
                persistent-hint
                v-model="inner.ImageURL"
              ></v-text-field>
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
                density="compact"
                variant="outlined"
                label="URL de préselection"
                hint="Cette URL permet de sélectionner automatiquement le séjour sur le formulaire d'inscription."
                persistent-hint
                readonly
                :model-value="props.preselectionUrl"
              >
                <template #append>
                  <v-btn
                    icon="mdi-content-copy"
                    size="small"
                    @click="copyPreselectionURL"
                  ></v-btn>
                </template>
              </v-text-field>
            </v-col>
          </v-row>
        </v-card-text>
      </v-card>

      <v-card class="my-2" subtitle="Autre">
        <v-card-text>
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
          </v-row>
        </v-card-text>
      </v-card>

      <CampMetaEdit
        :meta-entries-hints="metaEntriesHints"
        v-model="inner.Meta"
      ></CampMetaEdit>
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
import { Camps, copy, copyToClipboard, selectItems } from "@/utils";
import CampOptionsPrix from "./CampOptionsPrix.vue";
import CampMetaEdit from "./CampMetaEdit.vue";
import { controller } from "../../logic/logic";
const props = defineProps<{
  camp: Camp;
  preselectionUrl: string;
  metaEntriesHints: { keys: string[]; values: string[] };
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

function formatOption(inner: Camp) {
  const hasQF = Camps.isQuotientFamilialActive(inner.OptionQuotientFamilial);
  const hasOption = inner.OptionPrix.Active != OptionPrixKind.NoOption;
  if (!hasQF && !hasOption) {
    return "Aucune";
  }

  const chunks: string[] = [];
  if (Camps.isQuotientFamilialActive(inner.OptionQuotientFamilial)) {
    chunks.push("Quotient familial");
  }
  if (inner.OptionPrix.Active != OptionPrixKind.NoOption) {
    chunks.push(OptionPrixKindLabels[inner.OptionPrix.Active]);
  }
  return chunks.join(", ");
}

async function copyPreselectionURL() {
  await copyToClipboard(props.preselectionUrl);
  controller.showMessage("URL copiée dans le presse-papier.");
}
</script>
