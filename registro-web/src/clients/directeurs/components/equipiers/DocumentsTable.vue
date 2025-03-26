<template>
  <v-card
    title="Documents de l'équipe"
    subtitle="Demandes et fichiers déposés par l'équipe"
  >
    <template #append> </template>
    <v-card-text class="mt-4">
      <v-skeleton-loader type="table" v-if="!demandes"></v-skeleton-loader>
      <div v-else>
        <v-table class="text-center">
          <!-- header -->
          <tr>
            <td></td>
            <td v-for="demande in demandes">
              {{ CategorieLabels[demande.Categorie] }}
            </td>
          </tr>
          <tr
            v-for="(equipier, i) in props.equipiers"
            :class="i % 2 == 0 ? 'bg-grey-lighten-4 rounded' : ''"
          >
            <td class="text-left">{{ equipier.Personne }}</td>
            <td v-for="demande in demandes">
              <DemandeChip
                :demande="itemAt(equipier.Equipier.Id, demande.Id)"
                @update-state="
                  (s) => updateState(equipier.Equipier.Id, demande.Id, s)
                "
              ></DemandeChip>
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
  DemandeState,
  type Demande,
  type EquipierDemande,
  type EquipierExt,
  type IdDemande,
  type IdEquipier,
} from "../../logic/api";
import DemandeChip from "./DemandeChip.vue";

const props = defineProps<{
  equipiers: EquipierExt[];
}>();

onMounted(fetchDocuments);

const demandes = ref<Demande[]>([]);
const equipierMatrix = ref<EquipierDemande[]>([]);
async function fetchDocuments() {
  const res = await controller.EquipiersDemandesGet();
  if (res === undefined) return;
  demandes.value = res.Demandes || [];
  equipierMatrix.value = res.Equipiers || [];
}

function itemAt(idEquipier: IdEquipier, idDemande: IdDemande) {
  return equipierMatrix.value.find(
    (p) => p.Key.IdDemande == idDemande && p.Key.IdEquipier == idEquipier
  )!;
}

async function updateState(
  idEquipier: IdEquipier,
  idDemande: IdDemande,
  state: DemandeState
) {
  const res = await controller.EquipiersDemandeSet({
    IdEquipier: idEquipier,
    IdDemande: idDemande,
    State: state,
  });
  if (res === undefined) return;
  controller.showMessage("Demande modifiée avec succès.");
  itemAt(idEquipier, idDemande).State = state;
}
</script>
