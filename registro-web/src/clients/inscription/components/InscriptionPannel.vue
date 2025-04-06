<template>
  <v-stepper v-model="tab" editable :mobile="!smAndUp">
    <template #default="{ prev, next }">
      <v-stepper-header>
        <v-stepper-item
          title="Responsable"
          :value="1"
          :complete="state1 === true"
          :error="state1 === false"
          :subtitle="state1 === false ? 'Champs manquants' : ''"
          :color="state1 === true ? 'primary' : state1 === false ? 'red' : ''"
        ></v-stepper-item>
        <v-divider></v-divider>
        <v-stepper-item
          title="Participants"
          :value="2"
          :complete="state2 === true"
          :error="state2 === false"
          :subtitle="state2 === false ? 'Information manquante' : ''"
          :color="state2 === true ? 'primary' : state2 === false ? 'red' : ''"
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
          :complete="state4 === true"
          :error="state4 === false"
          :subtitle="state4 === false ? 'Charte à accepter' : ''"
          :color="state4 === true ? 'primary' : state4 === false ? 'red' : ''"
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
            <template v-if="smAndUp"> Valider mon inscription </template>
            <template v-else> Valider </template>
          </v-btn>
        </template>
      </v-stepper-actions>
    </template>
  </v-stepper>

  <!-- confirmation après API call -->
  <v-dialog v-model="showInscriptionSaved" max-width="600px">
    <v-card title="Confirmation">
      <v-card-text>
        Merci pour votre demande d'inscription ! <br />
        <br />
        Par mesure de sécurité, nous devons vérifier votre adresse mail. Un mail
        de confirmation a été envoyé à <b>{{ inner.Responsable.Mail }}</b> :
        veuillez valider votre demande en suivant le lien que vous y trouverez.
        <br />
        <br />
        <div class="text-grey font-italic">
          Vous pouvez désormais quitter cette page.
        </div>
      </v-card-text>
    </v-card>
  </v-dialog>
</template>

<script lang="ts" setup>
import { computed, reactive, ref, watch } from "vue";
import Step1 from "./Step1.vue";
import Step2 from "./Step2.vue";
import Step3 from "./Step3.vue";
import Step4 from "./Step4.vue";
import { Sexe, type Data } from "../logic/api";
import { ageFrom, isDateZero } from "@/components/date";
import { copy } from "@/utils";
import { controller } from "../logic/logic";
import { useDisplay } from "vuetify";

const props = defineProps<{
  data: Data;
}>();

const { smAndUp } = useDisplay();

const inner = reactive(copy(props.data.InitialInscription));

const tab = ref(1);

const hasStartedStep1 = ref(false);
const hasStartedStep2 = ref(false);
const hasStartedStep4 = ref(false);

const state1 = computed(() =>
  !hasStartedStep1.value ? undefined : isStep1Valid.value
);
const state2 = computed(() =>
  !hasStartedStep2.value ? undefined : isStep2Valid.value
);
const state4 = computed(() =>
  !hasStartedStep4.value ? undefined : isStep4Valid.value
);

watch(
  () => tab.value,
  (_, old) => {
    if (old == 1) hasStartedStep1.value = true;
    if (old == 2) hasStartedStep2.value = true;
    if (old == 4) hasStartedStep4.value = true;
  }
);

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
  return isStep1Valid.value && isStep2Valid.value && isStep4Valid.value;
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
