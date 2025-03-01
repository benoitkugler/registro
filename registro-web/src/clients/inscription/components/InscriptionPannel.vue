<template>
  <!-- confirmation après API call -->
  <v-dialog v-model="showInscriptionSaved" max-width="600px">
    <v-card title="Confirmation">
      <v-card-text>
        Merci pour votre demande d'inscription ! <br />
        <br />
        Par mesure de sécurité, nous devons vérifier votre adresse mail. Un mail
        de confirmation a été envoyé à <b>{{ inner.Responsable.Mail }}</b> :
        veuillez valider définitivement votre inscription en suivant le lien que
        vous y trouverez. <br />
        <br />
        <div class="text-grey font-italic">
          Vous pouvez désormais quitter cette page.
        </div>
      </v-card-text>
    </v-card>
  </v-dialog>

  <v-stepper v-model="tab" editable>
    <template #default="{ prev, next }">
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
          <Step1 v-model="inner.Responsable"></Step1>
        </v-stepper-window-item>
        <v-stepper-window-item :value="2">
          <Step2
            :model-value="inner.Participants || []"
            @update:model-value="(l) => (inner.Participants = l)"
            :camps="props.data.Camps || []"
            :responsable="inner.Responsable"
            :preselected="props.data.Settings.PreselectedCamp"
          ></Step2>
        </v-stepper-window-item>
        <v-stepper-window-item :value="3">
          <Step3 :settings="props.data.Settings"></Step3>
        </v-stepper-window-item>
        <v-stepper-window-item :value="4">
          <Step4
            v-model:partage-adresse="inner.PartageAdressesOK"
            v-model:mails="inner.CopiesMails"
            v-model:message="inner.Message"
            v-model:charte="isCharteOK"
            v-model:fond-soutien="inner.DemandeFondSoutien"
            :settings="props.data.Settings"
          ></Step4>
        </v-stepper-window-item>
      </v-stepper-window>

      <v-stepper-actions @click:prev="prev" @click:next="next">
        <template #next>
          <v-btn v-if="tab != 4" @click="next" variant="text">Suivant</v-btn>
          <v-btn
            v-else
            color="green"
            variant="outlined"
            @click="validInscription"
            :disabled="!isInscValid || isLoading"
          >
            <template #prepend>
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
import { computed, reactive, ref } from "vue";
import Step1 from "./Step1.vue";
import Step2 from "./Step2.vue";
import Step3 from "./Step3.vue";
import Step4 from "./Step4.vue";
import { Sexe, type Data } from "../logic/api";
import { ageFrom, isDateZero } from "@/components/date";
import { copy } from "@/utils";
import { controller } from "../logic/logic";

const props = defineProps<{
  data: Data;
}>();

const inner = reactive(copy(props.data.InitialInscription));

const tab = ref(1);

const isStep1Valid = computed(() => {
  const resp = inner.Responsable;
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
  const parts = inner.Participants;
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
  () => isCharteOK.value || !props.data.Settings.ShowCharteConduite
);
const isCharteOK = ref(false);

const isInscValid = computed(() => {
  return isStep1Valid.value && isStep2Valid.value && isStep4Valid;
});

const isLoading = ref(false);
const showInscriptionSaved = ref(false);
async function validInscription() {
  isLoading.value = true;
  const out = await controller.SaveInscription(inner);
  isLoading.value = false;
  if (out === undefined) return;

  showInscriptionSaved.value = true;
}
</script>
