<template>
  <NavBar></NavBar>

  <v-responsive class="fill-height mx-2">
    <v-card>
      <v-dialog
        :model-value="showBusExpenses"
        @update:model-value="showBusExpenses = false"
        max-width="800"
      >
        <BusExpenses
          v-if="showBusExpenses"
          @save="updateBusExpenses"
        ></BusExpenses>
      </v-dialog>

      <v-skeleton-loader v-if="frais == null" type="card"> </v-skeleton-loader>
      <v-card-text v-else class="text-left">
        <v-card title="Utilisation des véhicules" class="my-2">
          <template v-slot:append>
            <v-btn
              color="green"
              density="comfortable"
              @click="addBus"
              :disabled="props.readonly"
            >
              <template v-slot:prepend>
                <v-icon>mdi-plus</v-icon>
              </template>
              Ajouter un véhicule</v-btn
            >
          </template>
          <v-card-text>
            <v-row>
              <v-col>Véhicule</v-col>
              <v-col class="text-center">Début (km)</v-col>
              <v-col class="text-center">Fin (km)</v-col>
              <v-col class="text-center">Total (km)</v-col>
              <v-col class="text-center">Tarif (par km)</v-col>
              <v-col class="text-right">Total</v-col>
              <v-col cols="1"></v-col>
            </v-row>
            <v-alert v-if="showMandatoryAlert" type="warning" class="my-4">
              Pour un séjour à la Maison du Rocher, ces dépenses doivent être
              fournies !
            </v-alert>
            <BusRow
              v-for="(bus, i) in frais.Bus"
              :key="i"
              :bus="bus"
              :readonly="props.readonly"
              @delete="deleteBus(i)"
              @edit="(bus) => updateBus(i, bus)"
              ref="busRows"
            ></BusRow>
            <v-row class="mt-3 bg-green-lighten-4 rounded">
              <v-col cols="7" class="text-center" align-self="center">
                <div class="mb-2">
                  Les dépenses ci-contre sont déduites de l'indemnité ci-dessus.
                </div>
                <v-chip
                  @click="showBusExpenses = true"
                  elevation="4"
                  :disabled="props.readonly"
                >
                  Sélectionner les dépenses concernant les véhicules</v-chip
                >
              </v-col>
              <v-col cols="4" align-self="center">
                <v-row v-for="(rec, i) in frais.ExpensesForBus" :key="i">
                  <v-col cols="8">
                    {{ rec.Description }}
                  </v-col>
                  <v-col class="text-right" cols="4">
                    -{{ Formatters.montant(rec.Total) }}
                  </v-col>
                </v-row>
              </v-col>
              <v-col cols="1"></v-col>
            </v-row>
            <v-row>
              <v-spacer></v-spacer>
              <v-col cols="3" class="text-right font-weight-bold">
                {{ frais.TotalBus.toFixed(2) }}€
              </v-col>
              <v-col cols="1"></v-col>
            </v-row>
          </v-card-text>
        </v-card>

        <v-card title="Utilisation du matériel" class="my-2">
          <v-card-text v-if="frais.HasMateriel">
            <v-row>
              <v-col></v-col>
              <v-col class="text-center">Nombre de personnes</v-col>
              <v-col class="text-center">Durée (jours)</v-col>
              <v-col class="text-center">Nombre total de jours</v-col>
              <v-col class="text-center">Tarif unitaire</v-col>
              <v-col class="text-center">Total</v-col>
            </v-row>
            <v-row>
              <v-col>Produits d'entretien</v-col>
              <v-col class="text-center"> {{ frais.Materiel.NbPeople }}</v-col>
              <v-col class="text-center"> {{ frais.Materiel.Days }}</v-col>
              <v-col class="text-center">
                {{ frais.Materiel.DaysPeopleEntretien }}</v-col
              >
              <v-col class="text-center">
                {{
                  Formatters.montant({
                    Cent: frais.Materiel.TarifEntretien,
                    Currency: Currency.Euros,
                  })
                }}</v-col
              >
              <v-col class="text-right font-weight-bold">{{
                Formatters.montant({
                  Cent: frais.Materiel.CostEntretien,
                  Currency: Currency.Euros,
                })
              }}</v-col>
            </v-row>
            <v-row>
              <v-col>Produits pédagogiques</v-col>
              <v-col class="text-center">
                {{ frais.Materiel.NbInscrits }}</v-col
              >
              <v-col class="text-center"> {{ frais.Materiel.Days }}</v-col>
              <v-col class="text-center">
                {{ frais.Materiel.DaysPeoplePedagogique }}</v-col
              >
              <v-col class="text-center">
                {{
                  Formatters.montant({
                    Cent: frais.Materiel.TarifMateriel,
                    Currency: Currency.Euros,
                  })
                }}</v-col
              >
              <v-col class="text-right font-weight-bold">{{
                Formatters.montant({
                  Cent: frais.Materiel.CostMateriel,
                  Currency: Currency.Euros,
                })
              }}</v-col>
            </v-row>
          </v-card-text>
          <v-card-text v-else>
            <v-alert
              >Ce séjour n'est pas concerné par les frais de matériel Maison du
              Rocher.</v-alert
            >
          </v-card-text>
        </v-card>
      </v-card-text>
    </v-card>
  </v-responsive>
</template>

<script lang="ts" setup>
import { onMounted, ref, computed, nextTick, useTemplateRef } from "vue";
import NavBar from "../components/NavBar.vue";
import {
  CarKind,
  Currency,
  type Bus,
  type IdComptaSubrecord,
  type Int,
  type ResolvedFraisMDR,
} from "../logic/api.ts";
import { controller } from "../logic/logic.ts";
import BusRow from "../components/fraisAssoAcve/BusRow.vue";
import { Formatters } from "@/utils.ts";

const props = defineProps<{
  atMdr: boolean;
  readonly: boolean;
}>();

const emit = defineEmits<{
  (e: "updateMandatoryExpenses", alert: boolean): void;
}>();

onMounted(fetch);

const showMandatoryAlert = computed(
  () => props.atMdr && !frais.value?.Bus?.length,
);

const frais = ref<ResolvedFraisMDR | null>(null);
async function fetch() {
  const res = await controller.FraisMDRLoad({
    "id-compta": controller.idCompta,
  });
  if (res === undefined) return;
  frais.value = res;

  emit("updateMandatoryExpenses", showMandatoryAlert.value);
}

async function saveBus(bus: Bus[]) {
  if (!frais.value) return;
  const res = await controller.FraisMDRUpdate({
    IdCompta: controller.idCompta,
    Frais: { Bus: bus },
  });
  if (res === undefined) return;

  controller.showMessage("Utilisation des véhicules modifiée avec succès.");
  frais.value = res;

  emit("updateMandatoryExpenses", showMandatoryAlert.value);
}

async function updateBusExpenses(selected: IdComptaSubrecord[]) {
  showBusExpenses.value = false;
  const res = await controller.FraisMDRSetBusExpense({
    IdCompta: controller.idCompta,
    Records: selected,
  });
  if (res === undefined) return;
  controller.showMessage("Dépenses liées aux véhicules modifiées avec succès.");
  frais.value = res;
}

const busRows = useTemplateRef("busRows");

async function addBus() {
  const l: Bus[] = (frais.value?.Bus || []).map((v) => v);
  l.push({
    Name: "",
    Kind: CarKind.CarMinibus,
    KmStart: 0 as Int,
    KmEnd: 0 as Int,
  });
  await saveBus(l);

  const rows = busRows.value;
  if (rows) {
    nextTick(() => rows[rows.length - 1]?.startEdit());
  }
}

function deleteBus(i: number) {
  const l = (frais.value?.Bus || []).map((v) => v);
  l.splice(i, 1);
  saveBus(l);
}

function updateBus(i: number, bus: Bus) {
  const l: Bus[] = (frais.value?.Bus || []).map((v) => v);
  l[i] = bus;
  saveBus(l);
}

const showBusExpenses = ref(false);
</script>
