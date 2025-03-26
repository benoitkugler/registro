<template>
  <v-card title="Documents de l'équipe">
    <template #append> </template>
    <v-card-text class="mt-4">
      <v-skeleton-loader type="table" v-if="!demandes"></v-skeleton-loader>
      <div v-else>
        <v-table>
          <!-- header -->
          <tr>
            <td></td>
            <td v-for="demande in demandes">
              {{ CategorieLabels[demande.Categorie] }}
            </td>
          </tr>
          <tr v-for="equipier in props.equipiers">
            <td>{{ equipier.Personne }}</td>
            <td v-for="demande in demandes">
              <v-menu>
                <template #activator="{ props: menuProps }">
                  <v-chip v-bind="menuProps">
                    {{
                      DemandeStateLabels[
                        index(equipier.Equipier.Id, demande.Id).State
                      ]
                    }}
                  </v-chip>
                </template>
              </v-menu>
            </td>
          </tr>
        </v-table>
      </div>
    </v-card-text>
  </v-card>
</template>

<script lang="ts" setup>
import { ref, onMounted } from "vue";
import { controller } from "../../logic/logic";
import {
  CategorieLabels,
  DemandeStateLabels,
  type Demande,
  type DemandeKey,
  type EquipierDemandes,
  type EquipierExt,
  type IdDemande,
  type IdEquipier,
} from "../../logic/api";

const props = defineProps<{
  equipiers: EquipierExt[];
}>();

onMounted(fetchDocuments);

const demandes = ref<Demande[]>([]);
const equipierMatrix = ref<EquipierDemandes[]>([]);
async function fetchDocuments() {
  const res = await controller.EquipiersDemandesGet();
  if (res === undefined) return;
  demandes.value = res.Demandes || [];
  equipierMatrix.value = res.Equipiers || [];
}

function index(idEquipier: IdEquipier, idDemande: IdDemande) {
  return equipierMatrix.value.find(
    (p) => p.Key.IdDemande == idDemande && p.Key.IdEquipier == idEquipier
  )!;
}
</script>
