<template>
  <v-card
    title="Structures d'aide"
    subtitle="Organismes pouvant distribuer une aide extérieure"
  >
    <template #append>
      <v-btn @click="create">
        <template #prepend>
          <v-icon color="green">mdi-plus</v-icon>
        </template>
        Ajouter une structure
      </v-btn>
    </template>
    <v-card-text class="overflow-y-auto">
      <v-row v-for="structure in list">
        <v-col align-self="center">
          <v-list-item-title>
            {{ structure.Nom }}
          </v-list-item-title>
          <v-list-item-subtitle>
            {{ structure.Immatriculation }}
          </v-list-item-subtitle>
        </v-col>
        <v-col align-self="center" style="font-size: smaller">
          <div v-for="line in structure.Info.split('\n')">
            {{ line }}
          </div>
        </v-col>
        <v-col align-self="center" cols="auto">
          <v-menu>
            <template #activator="{ props: menuProps }">
              <v-btn icon="mdi-dots-vertical" v-bind="menuProps" size="x-small">
              </v-btn>
            </template>
            <v-list>
              <v-list-item
                title="Modifier"
                prepend-icon="mdi-pencil"
                @click="toEdit = copy(structure)"
              >
              </v-list-item>
              <v-list-item
                title="Supprimer"
                prepend-icon="mdi-delete"
                @click="toDelete = structure"
              >
              </v-list-item>
            </v-list>
          </v-menu>
        </v-col>
      </v-row>
    </v-card-text>

    <!-- editor -->
    <v-dialog
      :model-value="toEdit != null"
      @update:model-value="toEdit = null"
      max-width="600px"
    >
      <v-card title="Modifier la structure" v-if="toEdit != null">
        <v-card-text>
          <v-row>
            <v-col>
              <v-text-field
                variant="outlined"
                density="compact"
                label="Nom"
                v-model="toEdit.Nom"
                hide-details
              >
              </v-text-field>
            </v-col>
            <v-col>
              <v-text-field
                variant="outlined"
                density="compact"
                label="Immatriculation"
                v-model="toEdit.Immatriculation"
                hide-details
              >
              </v-text-field>
            </v-col>
          </v-row>
          <v-row>
            <v-col>
              <v-textarea
                variant="outlined"
                density="compact"
                label="Informations"
                v-model="toEdit.Info"
                hide-details
              ></v-textarea>
            </v-col>
          </v-row>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn @click="update">Enregistrer</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- confirme delete -->
    <v-dialog
      :model-value="toDelete != null"
      @update:model-value="toDelete = null"
      max-width="450px"
    >
      <v-card title="Confirmation" v-if="toDelete != null">
        <v-card-text>
          Confirmez-vous la suppression de la structure {{ toDelete.Nom }} ?
          <br /><br />
          Attention, cette opération est irréversible.
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="red" @click="delete_">Supprimer</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-card>
</template>

<script setup lang="ts">
import type {
  Structureaide,
  Structureaides,
} from "@/clients/backoffice/logic/api";
import { controller } from "@/clients/backoffice/logic/logic";
import { copy } from "@/utils";
import { computed, ref } from "vue";

const structures = defineModel<Structureaides>({ required: true });

const list = computed(() =>
  Object.values(structures.value || {}).sort((a, b) =>
    a.Nom.localeCompare(b.Nom)
  )
);

async function create() {
  const res = await controller.StructureaideCreate();
  if (res === undefined) return;

  controller.showMessage("Structure créée avec succès.");
  const m = structures.value || {};
  m[res.Id] = res;
  structures.value = m;
  toEdit.value = res; // start edit
}

const toEdit = ref<Structureaide | null>(null);
async function update() {
  const st = toEdit.value;
  if (st == null) return;
  toEdit.value = null;
  const res = await controller.StructureaideUpdate(st);
  if (res === undefined) return;

  controller.showMessage("Structure modifiée avec succès.");
  const m = structures.value || {};
  m[st.Id] = st;
  structures.value = m;
}

const toDelete = ref<Structureaide | null>(null);
async function delete_() {
  const st = toDelete.value;
  if (st == null) return;
  toDelete.value = null;
  const res = await controller.StructureaideDelete({ id: st.Id });
  if (res === undefined) return;

  const m = structures.value || {};
  delete m[st.Id];
  structures.value = m;
  controller.showMessage("Structure supprimée avec succès.");
}
</script>
