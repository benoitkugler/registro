<template>
  <v-stepper v-model="tab" editable>
    <template v-slot:default="{ prev, next }">
      <v-stepper-header>
        <v-stepper-item
          title="Responsable"
          :value="1"
          :complete="isStep1Valid"
          :error="!isStep1Valid"
          :subtitle="isStep1Valid ? '' : 'Champs manquants'"
          :color="isStep1Valid ? 'primary' : 'red'"
        ></v-stepper-item>
        <v-divider></v-divider>
        <v-stepper-item
          title="Participants"
          :value="2"
          :complete="isStep2Valid"
          :error="!isStep2Valid"
          :subtitle="isStep2Valid ? '' : 'Information manquante'"
          :color="isStep2Valid ? 'primary' : 'red'"
        ></v-stepper-item>
        <v-divider></v-divider>
        <v-stepper-item
          title="Information sur le paiement"
          :value="3"
        ></v-stepper-item>
        <v-divider></v-divider>
        <v-stepper-item
          title="Conclusion"
          :value="4"
          :complete="isStep4Valid"
          :error="!isStep4Valid"
          :subtitle="isStep4Valid ? '' : 'Charte à accepter'"
          :color="isStep4Valid ? 'primary' : 'red'"
        ></v-stepper-item>
      </v-stepper-header>

      <v-stepper-window class="mb-0">
        <v-stepper-window-item :value="1">
          <Step1 v-model="data.Responsable"></Step1>
        </v-stepper-window-item>
        <v-stepper-window-item :value="2">
          <Step2
            :model-value="data.Participants || []"
            @update:model-value="(l) => (data.Participants = l)"
            :camps="meta.Camps || []"
            :responsable="data.Responsable"
            :preselected="meta.PreselectedCamp"
          ></Step2>
        </v-stepper-window-item>
        <v-stepper-window-item :value="3">
          <Step3 :data="meta"></Step3>
        </v-stepper-window-item>
        <v-stepper-window-item :value="4">
          <Step4
            v-model:partage-adresse="data.PartageAdressesOK"
            v-model:mails="data.CopiesMails"
            v-model:message="data.Message"
            v-model:charte="isCharteOK"
            :data="meta"
          ></Step4>
        </v-stepper-window-item>
      </v-stepper-window>

      <v-stepper-actions @click:prev="prev" @click:next="next">
        <template v-slot:next>
          <v-btn v-if="tab != 4" @click="next" variant="text">Suivant</v-btn>
          <v-btn
            v-else
            color="green"
            variant="outlined"
            @click="onClickValid"
            :disabled="!isInscValid"
          >
            <template v-slot:prepend>
              <v-icon>mdi-check</v-icon>
            </template>
            Valider mon inscription
          </v-btn>
        </template>
      </v-stepper-actions>
    </template>
  </v-stepper>
</template>

<script lang="ts" setup>
import { computed, ref } from "vue";
import Step1 from "./Step1.vue";
import Step2 from "./Step2.vue";
import Step3 from "./Step3.vue";
import Step4 from "./Step4.vue";
import {
  Sexe,
  type DataInscription,
  type Date_,
  type IdCamp,
  type Inscription,
  type Int,
  type Participant,
} from "../logic/api";
import { ageFrom, isDateZero } from "@/components/date";

const data = ref<Inscription>({
  Responsable: {
    Nom: "",
    Prenom: "",
    Sexe: Sexe.Empty,
    Mail: "",
    Adresse: "",
    CodePostal: "",
    Ville: "",
    Pays: "",
    DateNaissance: "0001-01-01",
    Tels: [] as string[],
  },
  Participants: [] as Participant[],
} as Inscription);

const meta = ref<DataInscription>({
  Camps: [
    {
      Id: 1 as Int,
      DateDebut: "2025-02-12" as Date_,
      Duree: 10 as Int,
      Lieu: "Chamaloc",
      Description:
        "Dans la lettre du directeur qui vous parviendra par mail juste après votre inscription, un lien vous sera communiqué afin de choisir les jours de présence de votre enfant.",
      Navette: { Actif: true, Commentaire: "Depuis Guilherand Granges" },
      Places: 50 as Int,
      AgeMin: 8 as Int,
      AgeMax: 12 as Int,
      Nom: "C2",
      Prix: "35€ ou 25CHF",
      Direction: "Vincent JONAC",
    },
    {
      Id: 2 as Int,
      DateDebut: "2025-02-12" as Date_,
      Duree: 10 as Int,
      Lieu: "Chamaloc",
      Description:
        "Dans la lettre du directeur qui vous parviendra par mail juste après votre inscription, un lien vous sera communiqué afin de choisir les jours de présence de votre enfant.",
      Navette: { Actif: false, Commentaire: "Depuis Guilherand Granges" },
      Places: 50 as Int,
      AgeMin: 8 as Int,
      AgeMax: 90 as Int,
      Nom: "C4",
      Prix: "35€",
      Direction: "Jon",
    },
    {
      Id: 3 as Int,
      DateDebut: "2025-02-12" as Date_,
      Duree: 10 as Int,
      Lieu: "Chamaloc",
      Description: "",
      Navette: { Actif: false, Commentaire: "Depuis Guilherand Granges" },
      Places: 50 as Int,
      AgeMin: 8 as Int,
      AgeMax: 90 as Int,
      Nom: "No desc",
      Prix: "35€",
      Direction: "Jon",
    },
  ],
  InitialInscription: data.value,
  PreselectedCamp: 0 as IdCamp,
  SupportBonsCAF: false,
  SupportANCV: true,
  EmailRetraitMedia: "contact@acve.asso.fr",
  ShowCharteConduite: true,
});

const tab = ref(1);

const isStep1Valid = computed(() => {
  const resp = data.value.Responsable;
  const age = ageFrom(resp.DateNaissance) || 0;
  return !!(
    resp.Nom.length &&
    resp.Prenom.length &&
    resp.Sexe != Sexe.Empty &&
    !isDateZero(resp.DateNaissance) &&
    age >= 18 &&
    resp.Mail.length &&
    resp.Tels?.length &&
    resp.Adresse.length &&
    resp.CodePostal.length &&
    resp.Ville.length
  );
});

const isStep2Valid = computed(() => {
  const parts = data.value.Participants;
  return (
    !!parts?.length &&
    parts.every(
      (p) =>
        p.IdCamp > 0 &&
        p.Nom.length &&
        p.Prenom.length &&
        p.Sexe != Sexe.Empty &&
        !isDateZero(p.DateNaissance)
    )
  );
});

const isStep4Valid = computed(
  () => isCharteOK.value || !meta.value.ShowCharteConduite
);

const isCharteOK = ref(false);

const isInscValid = computed(() => {
  return isStep1Valid.value && isStep2Valid.value && isStep4Valid;
});

function onClickValid() {
  if (meta.value.ShowCharteConduite) {
    // display the dialog and exit early
    showCharteConduite.value = true;
    return;
  } else {
    // directly trigger the validation
    validInscription();
  }
}

async function validInscription() {}

const showCharteConduite = ref(false);
</script>
