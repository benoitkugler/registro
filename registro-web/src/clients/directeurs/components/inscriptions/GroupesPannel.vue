<template>
  <v-card title="Groupes" subtitle="Définir des groupes d'âge">
    <template #append>
      <v-btn prepend-icon="mdi-cog" class="mr-2" @click="showEditPlages = true"
        >Plages d'âge...</v-btn
      >
      <v-btn prepend-icon="mdi-plus" color="green" @click="createGroupe"
        >Ajouter</v-btn
      >
    </template>
    <v-card-text>
      <v-list>
        <v-list-item
          v-for="groupe in Object.values(props.groupes.Groupes || {})"
          :title="groupe.Nom"
          :subtitle="formatPlage(plages[groupe.Id])"
        >
          <template #prepend>
            <v-badge :color="groupe.Couleur" inline></v-badge>
          </template>
          <template #append>
            <v-btn
              icon="mdi-pencil"
              size="small"
              @click="toEdit = copy(groupe)"
            ></v-btn>
            <v-btn
              class="ml-2"
              icon="mdi-delete"
              size="small"
              color="red"
              @click="toDelete = groupe"
            ></v-btn>
          </template>
        </v-list-item>
      </v-list>
    </v-card-text>

    <!-- edit -->
    <v-dialog
      :model-value="toEdit != null"
      @update:model-value="toEdit = null"
      max-width="400px"
    >
      <v-card title="Modifier le groupe" v-if="toEdit">
        <v-card-text>
          <v-row>
            <v-col>
              <v-text-field
                autofocus
                v-model="toEdit.Nom"
                label="Nom"
                variant="outlined"
                density="comfortable"
              ></v-text-field>
            </v-col>
          </v-row>
          <v-row
            ><v-col>
              <ColorField
                label="Couleur"
                v-model="toEdit.Couleur"
                :swatches="[]"
              ></ColorField> </v-col
          ></v-row>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn @click="updateGroupe"> Enregistrer </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- plages -->
    <v-dialog v-model="showEditPlages" max-width="800px">
      <GroupesPlagesEditPannel
        :groupes="props.groupes"
        :participants="props.participants"
        @save="updatePlages"
      ></GroupesPlagesEditPannel>
    </v-dialog>

    <!-- delete -->
    <v-dialog
      :model-value="toDelete != null"
      @update:model-value="toDelete = null"
      max-width="600px"
    >
      <v-card title="Confirmer la suppression" v-if="toDelete">
        <v-card-text>
          Confirmez-vous la suppression du groupe {{ toDelete.Nom }} ?
          <br /><br />
          Attention, cette action est irréversible.
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn @click="deleteGroupe" color="red">Supprimer</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-card>
</template>

<script lang="ts" setup>
import { controller } from "../../logic/logic";
import {
  type Groupe,
  type Groupes,
  type GroupesOut,
  type IdGroupe,
  type ParticipantExt,
} from "../../logic/api";
import { computed, ref } from "vue";
import { copy } from "@/utils";
import { formatPlage, groupesPlages } from "./groupes";
import GroupesPlagesEditPannel from "./GroupesPlagesEditPannel.vue";

const props = defineProps<{
  groupes: GroupesOut;
  participants: ParticipantExt[];
}>();

const emit = defineEmits<{
  (e: "refresh"): void;
}>();

async function createGroupe() {
  const res = await controller.GroupeCreate();
  if (res === undefined) return;
  controller.showMessage("Groupe créé avec succès.");
  emit("refresh");
  toEdit.value = res;
}

const toEdit = ref<Groupe | null>(null);
async function updateGroupe() {
  if (toEdit.value == null) return;
  const res = await controller.GroupeUpdate(toEdit.value);
  toEdit.value = null;
  if (res === undefined) return;
  controller.showMessage("Groupe modifié avec succès.");
  emit("refresh");
}

const toDelete = ref<Groupe | null>(null);
async function deleteGroupe() {
  if (toDelete.value == null) return;
  const res = await controller.GroupeDelete({ id: toDelete.value.Id });
  toDelete.value = null;
  if (res === undefined) return;
  controller.showMessage("Groupe supprimé avec succès.");
  emit("refresh");
}

const plages = computed(() => groupesPlages(props.groupes.Groupes));

const showEditPlages = ref(false);
async function updatePlages(groupes: Groupes, erase: boolean) {
  showEditPlages.value = false;
  const res = await controller.GroupeUpdatePlages({
    Fins: Object.fromEntries(
      Object.values(groupes || {}).map((g) => [g.Id, g.Fin])
    ),
    OverrideManuel: erase,
  });
  if (res === undefined) return;
  controller.showMessage("Plages modifiées avec succès.");
  emit("refresh");
}
</script>
