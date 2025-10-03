<template>
  <v-card title="Régler le séjour" class="mt-2">
    <v-card-text>
      Le montant à régler est de
      <b> {{ props.dossier.Bilan.Restant }} </b>.
      <v-row class="mt-4">
        <v-col align-self="start" cols="4">
          <v-list class="py-0" selectable v-model:selected="mode">
            <v-list-item
              rounded
              class="my-2"
              title="Payer en ligne"
              subtitle="par carte bancaire"
              :value="ModePaiement.EnLigne"
              v-if="props.settings.SupportPaiementEnLigne"
            >
              <template #append>
                <v-icon>mdi-credit-card</v-icon>
              </template>
            </v-list-item>
            <v-list-item
              rounded
              class="my-2"
              title="Payer par virement"
              :value="ModePaiement.Virement"
            >
              <template #append>
                <v-icon>mdi-bank</v-icon>
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
                <v-icon>mdi-checkbook</v-icon>
              </template>
            </v-list-item>
          </v-list>
        </v-col>
        <v-col align-self="start" cols="8">
          <v-card :title="ModePaiementLabels[mode]">
            <v-card-text v-if="mode == ModePaiement.EnLigne">
              TODO: see https://github.com/benoitkugler/registro/issues/76
            </v-card-text>
            <v-card-text v-else-if="mode == ModePaiement.Virement">
              Vous pouvez effectuer votre virement vers
              {{
                props.settings.BankAccounts?.length || 0 > 1
                  ? `l'un des comptes suivants`
                  : `le compte suivant`
              }}
              :
              <v-list class="my-2">
                <v-list-item
                  :title="account[0]"
                  :subtitle="account[1]"
                  v-for="account in props.settings.BankAccounts"
                >
                  <template #append>
                    <v-btn
                      size="x-small"
                      icon="mdi-content-copy"
                      @click="copyIBAN(account[1])"
                    ></v-btn>
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
            <v-card-text v-else-if="mode == ModePaiement.Cheque">
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
import { ref } from "vue";
import {
  ModePaiement,
  ModePaiementLabels,
  type DossierExt,
  type PaiementSettings,
} from "../logic/api";
import { copyToClipboard } from "@/utils";
import { controller } from "../logic/logic";
const props = defineProps<{
  token: string;
  dossier: DossierExt;
  settings: PaiementSettings;
}>();

const emit = defineEmits<{}>();

const mode = ref<ModePaiement>(
  props.settings.SupportPaiementEnLigne
    ? ModePaiement.EnLigne
    : ModePaiement.Virement
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
