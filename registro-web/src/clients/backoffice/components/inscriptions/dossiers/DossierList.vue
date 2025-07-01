<template>
  <div>
    <v-row>
      <v-col align-self="center">
        <DebounceField
          density="compact"
          variant="outlined"
          prepend-inner-icon="mdi-magnify"
          label="Rechercher un dossier"
          hide-details
          clearable
          v-model="query.Pattern"
          @update:model-value="searchDossiers"
        >
          <template #append-inner>
            <v-chip
              class="ml-1"
              size="small"
              v-if="query.IdCamp.Valid"
              prepend-icon="mdi-bed"
            >
              {{ props.camps.find((c) => c.Id == query.IdCamp.Id)?.Label }}
            </v-chip>
            <v-chip
              class="ml-1"
              size="small"
              v-if="query.Attente != QueryAttente.EmptyQA"
              prepend-icon="mdi-clock"
            >
              {{ QueryAttenteLabels[query.Attente] }}
            </v-chip>
            <v-chip
              class="ml-1"
              size="small"
              v-if="query.Reglement != QueryReglement.EmptyQR"
              prepend-icon="mdi-currency-eur"
            >
              {{ QueryReglementLabels[query.Reglement] }}
            </v-chip>
            <v-chip class="ml-1" size="small" v-if="query.OnlyFondSoutien">
              Fond de soutien
            </v-chip>
          </template>

          <template #append>
            <v-menu location="left top" :close-on-content-click="false">
              <template #activator="{ props: menuProps }">
                <v-btn v-bind="menuProps" size="small" variant="flat" icon>
                  <v-icon>mdi-filter-cog</v-icon>
                  <v-tooltip
                    activator="parent"
                    text="Autres critères de recherche"
                  >
                  </v-tooltip>
                </v-btn>
              </template>
              <v-card min-width="400px" title="Filtrer les dossiers">
                <v-card-text>
                  <v-row>
                    <v-col>
                      <SelectCamp
                        label="Camp"
                        :camps="props.camps"
                        :model-value="
                          nullableToZeroable(optToNullable(query.IdCamp))
                        "
                        @update:model-value="
                          (id) => {
                            query.IdCamp = nullableToOpt(
                              zeroableToNullable(id)
                            );
                            searchDossiers();
                          }
                        "
                      ></SelectCamp>
                    </v-col>
                  </v-row>
                  <v-row>
                    <v-col>
                      <v-select
                        density="compact"
                        variant="outlined"
                        label="Liste d'attente"
                        v-model="query.Attente"
                        :items="selectItems(QueryAttenteLabels)"
                        @update:model-value="searchDossiers"
                        hide-details
                      ></v-select>
                    </v-col>
                  </v-row>
                  <v-row>
                    <v-col>
                      <v-select
                        density="compact"
                        variant="outlined"
                        label="Réglement"
                        v-model="query.Reglement"
                        :items="selectItems(QueryReglementLabels)"
                        @update:model-value="searchDossiers"
                        hide-details
                      ></v-select>
                    </v-col>
                  </v-row>
                  <v-row>
                    <v-col>
                      <v-switch
                        label="Trier selon les nouveaux messages"
                        hint="Affiche les nouveaux messages en premier."
                        persistent-hint
                        density="compact"
                        color="light-blue-darken-3"
                        v-model="query.SortByNewMessages"
                        @update:model-value="searchDossiers"
                      >
                      </v-switch>
                    </v-col>
                  </v-row>
                  <v-row>
                    <v-col>
                      <v-switch
                        label="Fond de soutien"
                        hint="N'affiche que les dossiers demandant le fond de soutien."
                        persistent-hint
                        density="compact"
                        color="orange"
                        v-model="query.OnlyFondSoutien"
                        @update:model-value="searchDossiers"
                      >
                      </v-switch>
                    </v-col>
                  </v-row>
                </v-card-text>
              </v-card>
            </v-menu>
          </template>
        </DebounceField>
      </v-col>
    </v-row>

    <v-list
      v-if="dossierHeaders != null"
      class="mt-2"
      :selected="props.selected ? [props.selected] : undefined"
    >
      <v-list-item v-if="!dossierHeaders.Dossiers?.length" class="text-center">
        <i>Aucun dossier ne correspond à votre recherche.</i>
      </v-list-item>
      <v-list-item
        rounded
        v-for="(dossier, i) in dossierHeaders.Dossiers"
        :key="i"
        :title="dossier.Responsable"
        :subtitle="dossier.Participants"
        :value="dossier.Id"
        @click="emit('click', dossier)"
        class="my-1"
      >
        <template #append v-if="dossier.NewMessages">
          <v-badge
            color="light-blue-darken-3"
            :content="dossier.NewMessages"
          ></v-badge>
        </template>
      </v-list-item>
    </v-list>
  </div>
</template>

<script setup lang="ts">
import {
  QueryAttente,
  QueryAttenteLabels,
  QueryReglement,
  QueryReglementLabels,
  type CampItem,
  type DossierHeader,
  type IdDossier,
  type SearchDossierIn,
  type SearchDossierOut,
} from "@/clients/backoffice/logic/api";
import { controller } from "@/clients/backoffice/logic/logic";
import {
  nullableToOpt,
  nullableToZeroable,
  optToNullable,
  selectItems,
  zeroableToNullable,
} from "@/utils";
import { ref, watch } from "vue";

const props = defineProps<{
  camps: CampItem[];
  selected?: IdDossier;
}>();

const emit = defineEmits<{
  (e: "click", dossier: DossierHeader): void;
  (e: "update", dossiers: SearchDossierOut): void;
}>();

const query = defineModel<SearchDossierIn>("query", { required: true });

defineExpose({ searchDossiers });

watch(() => query.value, searchDossiers, { immediate: true });

const dossierHeaders = ref<SearchDossierOut | null>(null);
async function searchDossiers() {
  const res = await controller.DossiersSearch(query.value);
  if (res === undefined) return;
  dossierHeaders.value = res;
  emit("update", res);
}
</script>
