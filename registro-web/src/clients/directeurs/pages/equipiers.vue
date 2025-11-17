<template>
  <NavBar :title="`${controller.camp?.Label} - Equipiers`">
    <v-btn @click="showDocuments = true" prepend-icon="mdi-file">
      Documents
    </v-btn>
  </NavBar>

  <v-card class="mt-2 mx-auto" max-width="800px">
    <template #title>
      <v-btn flat icon>
        <v-icon>mdi-information</v-icon>
        <v-tooltip activator="parent" content-class="ma-0 pa-0">
          <StatistiquesCard :equipiers="equipiers"></StatistiquesCard>
        </v-tooltip>
      </v-btn>
      Équipiers
    </template>

    <template #append>
      <v-btn
        color="green"
        @click="showCreateEquipier = true"
        prepend-icon="mdi-plus"
      >
        Ajouter un équipier</v-btn
      >
      <v-menu>
        <template #activator="{ props: menuProps }">
          <v-btn
            v-bind="menuProps"
            class="ml-2"
            icon="mdi-dots-vertical"
            size="small"
          ></v-btn>
        </template>
        <v-list>
          <v-list-item>
            <v-list-item
              title="Inviter tous"
              subtitle="Envoi un mail à tous les équipiers en attente"
              prepend-icon="mdi-email"
              @click="inviteAll()"
            ></v-list-item>
          </v-list-item>
        </v-list>
      </v-menu>
    </template>
    <v-card-text>
      <v-list>
        <v-list-item
          v-for="equipier in equipiers"
          :title="equipier.Personne"
          :subtitle="formatRoles(equipier.Equipier.Roles)"
        >
          <template #append>
            <v-row>
              <v-col cols="2" align-self="center">
                <v-tooltip
                  :text="`${equipier.Personne} a son anniveraire pendant le séjour !`"
                >
                  <template #activator="{ props }">
                    <v-icon v-bind="props" color="pink"
                      >mdi-cake-variant</v-icon
                    >
                  </template>
                </v-tooltip>
              </v-col>
              <v-col align-self="center">
                <v-chip
                  :prepend-icon="
                    formatStatut(equipier.Equipier.FormStatus).icon
                  "
                  :color="formatStatut(equipier.Equipier.FormStatus).color"
                >
                  {{ FormStatusEquipierLabels[equipier.Equipier.FormStatus] }}
                </v-chip>
              </v-col>

              <v-col
                align-self="center"
                v-if="
                  equipier.Equipier.FormStatus == FormStatusEquipier.Answered &&
                  !equipier.Equipier.AccepteCharte.Bool
                "
              >
                <v-chip color="warning">Charte non acceptée</v-chip>
              </v-col>

              <v-col align-self="center">
                <v-menu>
                  <template #activator="{ props: menuProps }">
                    <v-btn
                      v-bind="menuProps"
                      icon="mdi-dots-vertical"
                      size="x-small"
                    ></v-btn>
                  </template>
                  <v-list>
                    <v-list-item
                      title="Inviter"
                      subtitle="Envoi un mail avec le lien vers le formulaire"
                      prepend-icon="mdi-email"
                      @click="inviteOne(equipier.Equipier.Id)"
                    ></v-list-item>
                    <v-list-item
                      title="Aller au formulaire"
                      prepend-icon="mdi-open-in-new"
                      :href="equipier.FormURL"
                      target="_blank"
                    ></v-list-item>
                    <v-divider></v-divider>
                    <v-list-item
                      title="Supprimer"
                      prepend-icon="mdi-delete"
                      @click="equipierToDelete = equipier"
                    ></v-list-item>
                  </v-list>
                </v-menu>
              </v-col>
            </v-row>
          </template>
        </v-list-item>
      </v-list>
    </v-card-text>
  </v-card>

  <v-dialog v-model="showDocuments">
    <DocumentsTable :equipiers="equipiers"></DocumentsTable>
  </v-dialog>

  <v-dialog v-model="showCreateEquipier" max-width="600px">
    <AddEquipierCard @create="createEquipier"></AddEquipierCard>
  </v-dialog>

  <!-- confirme delete -->
  <v-dialog
    :model-value="equipierToDelete != null"
    @update:model-value="equipierToDelete = null"
    max-width="400px"
  >
    <v-card title="Confirmation" v-if="equipierToDelete">
      <v-card-text>
        Confirmez-vous la suppression de l'équipier
        {{ equipierToDelete.Personne }} ? <br /><br />
        Attention, cette opération est irréversible.
      </v-card-text>
      <v-card-actions>
        <v-spacer></v-spacer>
        <v-btn color="red" @click="deleteEquipier"> Supprimer </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script lang="ts" setup>
import { useRouter } from "vue-router";
import NavBar from "../components/NavBar.vue";
import { onMounted, ref } from "vue";
import { controller } from "../logic/logic";
import {
  FormStatusEquipier,
  FormStatusEquipierLabels,
  RoleLabels,
  type EquipierExt,
  type EquipiersCreateIn,
  type IdEquipier,
  type Roles,
} from "../logic/api";
import DocumentsTable from "../components/equipiers/DocumentsTable.vue";
import AddEquipierCard from "../components/equipiers/AddEquipierCard.vue";
import { nullableToOpt } from "@/utils";
import StatistiquesCard from "../components/equipiers/StatistiquesCard.vue";

const router = useRouter();

onMounted(fetchEquipiers);

const equipiers = ref<EquipierExt[]>([]);
async function fetchEquipiers() {
  const res = await controller.EquipiersGet();
  if (res === undefined) return;
  res?.sort((a, b) => a.Equipier.Id - b.Equipier.Id);
  equipiers.value = res || [];
}

const showDocuments = ref(false);

function formatRoles(rs: Roles) {
  return (rs || []).map((r) => RoleLabels[r]).join(", ");
}

function formatStatut(s: FormStatusEquipier) {
  switch (s) {
    case FormStatusEquipier.NotSend:
      return { icon: "mdi-form-select", color: "" };
    case FormStatusEquipier.Pending:
      return { icon: "mdi-clock", color: "yellow-darken-2" };
    case FormStatusEquipier.Answered:
      return { icon: "mdi-check", color: "green" };
  }
}

const showCreateEquipier = ref(false);
async function createEquipier(args: EquipiersCreateIn) {
  showCreateEquipier.value = false;
  const res = await controller.EquipiersCreate(args);
  if (res === undefined) return;
  controller.showMessage("Equipier crée avec succès.");
  equipiers.value.push(res);
}

async function inviteOne(id: IdEquipier) {
  const res = await controller.EquipiersInvite({ OnlyOne: nullableToOpt(id) });
  if (res === undefined) return;
  controller.showMessage("Equipier invité avec succès.");
  res?.sort((a, b) => a.Equipier.Id - b.Equipier.Id);
  equipiers.value = res || [];
}
async function inviteAll() {
  const res = await controller.EquipiersInvite({
    OnlyOne: nullableToOpt<IdEquipier>(null),
  });
  if (res === undefined) return;
  controller.showMessage("Equipiers invités avec succès.");
  res?.sort((a, b) => a.Equipier.Id - b.Equipier.Id);
  equipiers.value = res || [];
}

const equipierToDelete = ref<EquipierExt | null>(null);
async function deleteEquipier() {
  if (equipierToDelete.value == null) return;
  const id = equipierToDelete.value.Equipier.Id;
  equipierToDelete.value = null;
  const res = await controller.EquipiersDelete({ id });
  if (res === undefined) return;
  controller.showMessage("Equipier supprimé avec succès.");
  // remove from view
  equipiers.value = equipiers.value.filter((eq) => eq.Equipier.Id != id);
}
</script>
