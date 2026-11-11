<template>
  <div>
    <v-card subtitle="Effectifs et durée" color="deep-orange-lighten-4">
      <v-card-text>
        <v-row>
          <v-col align-self="center">
            Durée : {{ props.resolved.Duree }} jours</v-col
          >
          <v-col align-self="center"
            >Prix du séjour :
            {{ Formatters.montant(props.resolved.Prix) }}</v-col
          >
          <v-col align-self="center">
            <template v-if="props.simulated">
              <IntField
                debounce
                label="Nombre estimé d'inscrits"
                v-model="inner.SimulatedInscrits"
                @update:model-value="update"
                :min="0 as Int"
              ></IntField>
            </template>
            <!-- real -->
            <template v-else>
              Nombre d'inscrits : {{ props.resolved.Inscrits }}
            </template>
          </v-col>
          <v-col align-self="center">
            <template v-if="props.simulated">
              <IntField
                debounce
                label="Nombre estimé d'équipiers"
                v-model="inner.SimulatedEquipiers"
                @update:model-value="update"
                :min="0 as Int"
              ></IntField>
            </template>
            <!-- real -->
            <template v-else>
              Nombre d'équipiers : {{ props.resolved.Equipiers }}
            </template>
          </v-col>
        </v-row>
      </v-card-text>
    </v-card>

    <!-- repas -->
    <v-card subtitle="Repas" class="my-2" color="light-blue-lighten-4">
      <v-card-text>
        <v-row>
          <!-- inscrits -->
          <v-col>
            <v-row>
              <v-col cols="12">
                <MealField
                  v-model="inner.StartMealInscrits"
                  label="Premier repas des inscrits"
                  @update:model-value="update"
                ></MealField>
              </v-col>
              <v-col cols="12">
                <MealField
                  v-model="inner.EndMealInscrits"
                  label="Dernier repas des inscrits"
                  @update:model-value="update"
                ></MealField>
              </v-col>
            </v-row>
          </v-col>
          <v-divider vertical thickness="1"></v-divider>
          <!-- equipiers -->
          <v-col>
            <v-row>
              <v-col>
                <v-select
                  variant="outlined"
                  density="compact"
                  v-model="inner.OffsetStartMealEquipiers"
                  @update:model-value="update"
                  label="Arrivée des équipiers"
                  hide-details
                  :items="[
                    { title: '4 jours avant', value: -4 },
                    { title: '3 jours avant', value: -3 },
                    { title: '2 jours avant', value: -2 },
                    { title: '1 jour avant', value: -1 },
                    { title: 'Premier jour du camp', value: 0 },
                  ]"
                ></v-select>
              </v-col>
              <v-col>
                <MealField
                  v-model="inner.StartMealEquipiers"
                  label="Premier repas des équipiers"
                  @update:model-value="update"
                ></MealField>
              </v-col>
            </v-row>

            <v-row>
              <v-col>
                <v-select
                  variant="outlined"
                  density="compact"
                  v-model="inner.OffsetEndMealEquipiers"
                  @update:model-value="update"
                  label="Départ des équipiers"
                  hide-details
                  :items="[
                    { title: 'Dernier jour du camp', value: 0 },
                    { title: '1 jour après', value: 1 },
                    { title: '2 jours après', value: 2 },
                    { title: '3 jours après', value: 3 },
                    { title: '4 jours après', value: 4 },
                  ]"
                ></v-select>
              </v-col>
              <v-col>
                <MealField
                  v-model="inner.EndMealEquipiers"
                  label="Dernier repas des équipiers"
                  @update:model-value="update"
                ></MealField>
              </v-col>
            </v-row>
          </v-col>
        </v-row>
        <v-divider horizontal thickness="1" class="my-4"></v-divider>
        <v-row>
          <v-col align-self="center">
            Total inscrits :
            {{ Formatters.montant(props.resolved.FoodInscrits) }}
          </v-col>
          <v-col align-self="center">
            Total équipiers :
            {{ Formatters.montant(props.resolved.FoodEquipiers) }}
          </v-col>
          <v-col align-self="center" v-if="props.resolved.HasSmallTeamFactor">
            <v-chip color="info">
              <template v-slot:append>
                <HelpTooltip
                  :text="smallFoodTooltip"
                  class="ml-2"
                ></HelpTooltip>
              </template>
              Petit groupe
            </v-chip>
          </v-col>
          <v-col
            align-self="center"
            class="text-right font-weight-bold"
            cols="auto"
          >
            Total : {{ Formatters.montant(props.resolved.FoodTotal) }}
          </v-col>
        </v-row>
      </v-card-text>
    </v-card>

    <v-card subtitle="Hébergement" class="my-2" color="yellow-accent-1">
      <v-card-text>
        <v-row justify="space-between">
          <template v-if="props.camp.AtMDR">
            <v-col>
              <HelpTooltip :text="chauffageTooltip"></HelpTooltip>
              Participation au chauffage MdR :
              {{ Formatters.montant(props.resolved.ChauffageMDR) }}
            </v-col>
            <v-col>
              <HelpTooltip :text="entretienTooltip"></HelpTooltip>
              Participation à l'entretien MdR :
              {{ Formatters.montant(props.resolved.EntretienMDR) }}
            </v-col>
          </template>
          <template v-else>
            <v-col cols="6">
              <v-row v-for="(_, i) in inner.ExternalHost" :key="i">
                <v-col>
                  <v-text-field
                    label="Description"
                    variant="outlined"
                    density="compact"
                    hide-details
                    v-model="inner.ExternalHost![i].Name"
                    @blur="update"
                  ></v-text-field>
                </v-col>
                <v-col cols="5">
                  <MontantField
                    debounce
                    v-model="inner.ExternalHost![i].Cost"
                    @update:model-value="update"
                    label="Coût"
                    hide-details
                  ></MontantField>
                </v-col>
                <v-col cols="auto" align-self="center">
                  <v-btn icon size="x-small" @click="deleteExternalHost(i)">
                    <v-icon color="red">mdi-delete</v-icon>
                  </v-btn>
                </v-col>
              </v-row>
              <v-row justify="center">
                <v-col cols="auto">
                  <v-btn rounded="pill" @click="addExternalHost" size="small">
                    <template v-slot:prepend>
                      <v-icon color="green">mdi-plus</v-icon>
                    </template>
                    Ajouter un hébergement
                  </v-btn>
                </v-col>
              </v-row>
            </v-col>
          </template>
          <v-col
            class="text-right font-weight-bold"
            cols="auto"
            align-self="center"
            >Total : {{ Formatters.montant(props.resolved.HostTotal) }}</v-col
          >
        </v-row>
      </v-card-text>
    </v-card>

    <v-card subtitle="Activités" class="my-2" color="light-green-lighten-3">
      <v-card-text>
        <v-row justify="space-between">
          <v-col cols="6" align-self="center">
            <v-row v-for="(_, i) in inner.Activities" :key="i">
              <v-col>
                <v-text-field
                  label="Description"
                  variant="outlined"
                  density="compact"
                  hide-details
                  v-model="inner.Activities![i].Name"
                  @blur="update"
                ></v-text-field>
              </v-col>
              <v-col cols="5">
                <MontantField
                  debounce
                  v-model="inner.Activities![i].Cost"
                  @update:model-value="update"
                  label="Coût"
                  hide-details
                ></MontantField>
              </v-col>
              <v-col cols="auto" align-self="center">
                <v-btn icon size="x-small" @click="deleteActivity(i)">
                  <v-icon color="red">mdi-delete</v-icon>
                </v-btn>
              </v-col>
            </v-row>
            <v-row justify="center">
              <v-col cols="auto">
                <v-btn-group density="compact" rounded="pill" divided>
                  <v-btn @click="addActivity" size="small">
                    <template v-slot:prepend>
                      <v-icon color="green">mdi-plus</v-icon>
                    </template>
                    Ajouter une activité
                  </v-btn>
                  <v-menu>
                    <template v-slot:activator="{ props: menuProps }">
                      <v-btn
                        v-bind="menuProps"
                        :disabled="!activityHints.length"
                      >
                        <v-icon>mdi-assistant</v-icon>
                      </v-btn>
                    </template>
                    <v-card
                      title="Importer une activité"
                      subtitle="Les activités ci-dessous proviennent des autres séjours."
                    >
                      <v-card-text>
                        <v-list>
                          <v-list-item
                            v-for="(activity, i) in activityHints"
                            @click="copyActivity(activity.Activity)"
                            :key="i"
                            :title="activity.Activity.Name"
                            :subtitle="Camps.label(activity.Origin)"
                          >
                            <template v-slot:append>
                              {{ Formatters.montant(activity.Activity.Cost) }}
                            </template>
                          </v-list-item>
                        </v-list>
                      </v-card-text>
                    </v-card>
                  </v-menu>
                </v-btn-group>
              </v-col>
            </v-row>
          </v-col>
          <v-divider vertical thickness="1"></v-divider>
          <v-col cols="4">
            <v-row v-if="props.camp.AtMDR">
              <v-col align-self="center" class="text-center">
                Matériel pédagogique MdR :
                {{ Formatters.montant(props.resolved.MaterielMDR) }}
              </v-col>
            </v-row>
            <v-row>
              <v-col align-self="center">
                <MontantField
                  debounce
                  label="Matériel pédagogique"
                  v-model="inner.MaterielPeda"
                  @update:model-value="update"
                  hide-details
                ></MontantField> </v-col
            ></v-row>
            <v-row>
              <v-col align-self="center">
                <MontantField
                  debounce
                  label="Sanitaire et pharmacie"
                  v-model="inner.ProduitPharma"
                  @update:model-value="update"
                  hide-details
                ></MontantField>
              </v-col>
            </v-row>
          </v-col>

          <v-col
            align-self="center"
            class="text-right font-weight-bold"
            cols="auto"
            >Total : {{ Formatters.montant(props.resolved.ActivitiesTotal) }}
          </v-col>
        </v-row>
      </v-card-text>
    </v-card>

    <v-card subtitle="Transport" class="my-2" color="orange-lighten-3">
      <v-card-text>
        <v-row justify="space-between">
          <v-col cols="4" align-self="center">
            <v-row>
              <v-col cols="12">
                <MontantField
                  debounce
                  label="Trajets de l'équipe"
                  v-model="inner.TeamTrips"
                  @update:model-value="update"
                  hide-details
                ></MontantField>
              </v-col>
              <v-col cols="12">
                <MontantField
                  debounce
                  label="Véhicules externes"
                  v-model="inner.ExternalCarRent"
                  @update:model-value="update"
                  hide-details
                  hint="Location, carburant, ..."
                ></MontantField>
              </v-col>
            </v-row>
          </v-col>

          <v-divider vertical thickness="1"></v-divider>

          <v-col cols="6">
            <v-row>
              <v-col cols="12">
                <IntField
                  debounce
                  label="Utilisation du bus de l'asso."
                  v-model="inner.AssoCarKms[0]"
                  @update:model-value="update"
                  suffix="kms"
                  hint="Nombre de kilomètres"
                ></IntField>
              </v-col>
              <v-col cols="6">
                <IntField
                  debounce
                  label="Utilisation des minubus de l'asso."
                  v-model="inner.AssoCarKms[1]"
                  @update:model-value="update"
                  suffix="kms"
                  hint="Nombre de kilomètres (cumulés)"
                ></IntField>
              </v-col>
              <v-col cols="6">
                <IntField
                  debounce
                  label="Utilisation des véhicules légers de l'asso."
                  v-model="inner.AssoCarKms[2]"
                  @update:model-value="update"
                  suffix="kms"
                  hint="Nombre de kilomètres (cumulés)"
                ></IntField>
              </v-col>
            </v-row>
          </v-col>

          <v-col
            align-self="center"
            class="text-right font-weight-bold"
            cols="auto"
            >Total : {{ Formatters.montant(props.resolved.TransportTotal) }}
          </v-col>
        </v-row>
      </v-card-text>
    </v-card>

    <v-card subtitle="Bilan" class="my-2" color="pink-lighten-3">
      <v-card-text>
        <v-row>
          <v-col align-self="center">
            Total des pensions :
            {{ Formatters.montant(props.resolved.Pensions) }}
          </v-col>
          <v-col align-self="center">
            Total des dépenses :
            {{ Formatters.montant(props.resolved.Depenses) }}
          </v-col>
          <v-col align-self="center" class="font-italic">
            Participation aux frais de l'asso. :
            {{ props.resolved.BilanPercentage }}%
          </v-col>
          <v-col align-self="center" class="text-right pr-1">
            <v-chip label class="font-weight-bold mx-0">
              Bilan : {{ Formatters.montant(props.resolved.Bilan) }}
            </v-chip>
          </v-col>
        </v-row>
      </v-card-text>
    </v-card>
  </div>
</template>

<script lang="ts" setup>
import { computed, onMounted, ref } from "vue";
import {
  Currency,
  type ActivityHint,
  type CampHeader,
  type ComptaBudget,
  type ComptaConfig,
  type Int,
  type NamedCost,
  type ResolvedBudget,
} from "../../logic/api.ts";
import { Camps, copy, Formatters } from "@/utils.ts";
import { controller } from "../../logic/logic.ts";
import HelpTooltip from "./HelpTooltip.vue";
import MealField from "./MealField.vue";

const props = defineProps<{
  camp: CampHeader;
  simulated: boolean;
  origin: ComptaBudget;
  resolved: ResolvedBudget;
  config: ComptaConfig;
}>();

const emit = defineEmits<{
  (e: "update", origin: ComptaBudget): void;
}>();

onMounted(fetchActivities);

const inner = ref(copy(props.origin));

function update() {
  emit("update", inner.value);
}

const smallFoodTooltip = computed(
  () =>
    `Effectif des inscrits inférieurs à ${props.config.SmallMealThreshold} : +${props.config.SmallMealMajoration}%`,
);
const chauffageTooltip = computed(
  () =>
    `De Novembre à Avril, ${(props.config.ChauffageMDR / 100).toFixed(
      2,
    )}€ par jour`,
);
const entretienTooltip = computed(
  () =>
    `${(props.config.EntretienMDR / 100).toFixed(2)}€ par jour et par inscrit`,
);

function addActivity() {
  inner.value.Activities = (inner.value.Activities || []).concat({
    Name: "",
    Cost: { Cent: 10000 as Int, Currency: Currency.Euros },
  });
  update();
}
function copyActivity(activity: NamedCost) {
  inner.value.Activities = (inner.value.Activities || []).concat(activity);
  update();
}

function deleteActivity(i: number) {
  (inner.value.Activities || []).splice(i, 1);
  update();
}

function addExternalHost() {
  inner.value.ExternalHost = (inner.value.ExternalHost || []).concat({
    Name: "",
    Cost: { Cent: 10000 as Int, Currency: Currency.Euros },
  });
  update();
}

function deleteExternalHost(i: number) {
  (inner.value.ExternalHost || []).splice(i, 1);
  update();
}

const activityHints = ref<ActivityHint[]>([]);
async function fetchActivities() {
  const res = await controller.GetBudgetActivities({
    "id-compta": props.camp.Compta.Id,
  });
  if (res === undefined) return;
  activityHints.value = res || [];
}
</script>
