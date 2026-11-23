<template>
  <v-card title="Régler le séjour" class="mt-2">
    <v-card-text>
      Le montant à régler est de <b> {{ amountToPay }} </b>.
      <template v-if="props.state == RequiredPaiementState.PaiementAcompte">
        <br />
        Il s'agit de l'acompte demandé pour votre inscription. Le solde sera
        demandé ultérieurement.
      </template>
      <v-row class="mt-4">
        <v-col align-self="start" cols="12" md="4">
          <v-list class="py-0" selectable v-model:selected="selectedMode">
            <v-list-item
              rounded
              class="my-2"
              title="Payer en ligne"
              subtitle="par carte bancaire"
              :value="ModePaiement.EnLigne"
              v-if="props.settings.SupportPaiementEnLigne"
            >
              <template #append>
                <v-icon
                  :icon="Formatters.paiementIcon(ModePaiement.EnLigne)"
                ></v-icon>
              </template>
            </v-list-item>
            <v-list-item
              rounded
              class="my-2"
              title="Payer par virement"
              :value="ModePaiement.Virement"
            >
              <template #append>
                <v-icon
                  :icon="Formatters.paiementIcon(ModePaiement.Virement)"
                ></v-icon>
              </template>
            </v-list-item>
            <v-list-item
              rounded
              class="my-2"
              title="Payer par chèque"
              :value="ModePaiement.Cheque"
              v-if="props.settings.Cheques.IsValid"
            >
              <template #append>
                <v-icon
                  :icon="Formatters.paiementIcon(ModePaiement.Cheque)"
                ></v-icon>
              </template>
            </v-list-item>
          </v-list>
        </v-col>
        <v-col align-self="start" cols="12" md="8">
          <v-card :title="ModePaiementLabels[selectedMode]">
            <v-card-text v-if="selectedMode == ModePaiement.EnLigne">
              TODO: see https://github.com/benoitkugler/registro/issues/76
            </v-card-text>
            <v-card-text v-else-if="selectedMode == ModePaiement.Virement">
              Vous pouvez effectuer votre virement vers
              {{
                props.settings.BankAccounts?.length || 0 > 1
                  ? `l'un des comptes suivants`
                  : `le compte suivant`
              }}
              :
              <v-list class="my-2">
                <v-list-item
                  :title="account.Owner"
                  :subtitle="account.IBAN"
                  v-for="account in props.settings.BankAccounts"
                >
                  <template #append>
                    <v-row no-gutters>
                      <v-col align-self="center">
                        {{ account.Name }}
                      </v-col>
                      <v-col align-self="center" cols="auto" class="ml-2">
                        <v-btn
                          size="x-small"
                          icon="mdi-content-copy"
                          @click="copyIBAN(account.IBAN)"
                        ></v-btn>
                      </v-col>
                    </v-row>
                  </template>
                </v-list-item>
              </v-list>
              <v-alert type="info">
                Merci d'indiquer impérativement le <b>label</b> suivant sur
                votre virement, qui nous permettra de l'identifier :
                <v-row no-gutters class="mt-2">
                  <v-col align-self="center" class="text-center">
                    <v-chip variant="elevated">
                      {{ props.settings.VirementCode }}
                    </v-chip>
                  </v-col>
                  <v-col align-self="center" cols="auto">
                    <v-btn
                      class="my-2"
                      size="x-small"
                      icon="mdi-content-copy"
                      @click="copyCode(props.settings.VirementCode)"
                    ></v-btn>
                  </v-col>
                </v-row>
              </v-alert>
            </v-card-text>
            <v-card-text v-else-if="selectedMode == ModePaiement.Cheque">
              Merci d'envoyer votre chèque à l'ordre
              <b>{{ props.settings.Cheques.Ordre }}</b> à l'adresse suivante :
              <v-row justify="center">
                <v-col cols="auto">
                  <v-card class="my-2">
                    <v-card-text>
                      {{ props.settings.Cheques.Adresse[0] }} <br />
                      {{ props.settings.Cheques.Adresse[1] }}
                    </v-card-text>
                  </v-card>
                </v-col>
              </v-row>

              Il sera encaissé environ une dizaine de jours avant le début du
              séjour.
            </v-card-text>
          </v-card>
        </v-col>
      </v-row>
    </v-card-text>
  </v-card>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import {
  ModePaiement,
  ModePaiementLabels,
  RequiredPaiementState,
  type DossierExt,
  type PaiementSettings,
} from "../logic/api";
import { copyToClipboard, Formatters } from "@/utils";
import { controller } from "../logic/logic";
const props = defineProps<{
  token: string;
  dossier: DossierExt;
  settings: PaiementSettings;
  state: RequiredPaiementState;
}>();

const emit = defineEmits<{}>();

const amountToPay = computed(() =>
  props.state === RequiredPaiementState.PaiementAcompte
    ? props.dossier.Bilan.AcompteRestant
    : props.dossier.Bilan.Restant,
);

const selectedMode = ref<ModePaiement>(
  props.settings.SupportPaiementEnLigne
    ? ModePaiement.EnLigne
    : ModePaiement.Virement,
);

async function copyIBAN(iban: string) {
  await copyToClipboard(iban);
  controller.showMessage("IBAN copié avec succès.");
}

async function copyCode(code: string) {
  await copyToClipboard(code);
  controller.showMessage("Identifiant de virement copié avec succès.");
}
</script>
