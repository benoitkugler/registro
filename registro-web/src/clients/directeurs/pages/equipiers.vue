<template>
  <NavBar :title="`${controller.camp?.Label} - Equipiers`">
    <v-btn @click="showDocuments = true">
      <template #prepend>
        <v-icon>mdi-file</v-icon>
      </template>
      Documents
    </v-btn>
  </NavBar>

  <v-card class="ma-2" title="Liste des équipiers">
    <template #append>
      <v-btn color="green" @click="showCreateEquipier = true">
        <template #prepend>
          <v-icon>mdi-plus</v-icon>
        </template>
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
import { ar } from "vuetify/locale";
import AddEquipierCard from "../components/equipiers/AddEquipierCard.vue";
import { create } from "node_modules/axios/index.cjs";
import { nullableToOpt, zeroableToNullable } from "@/utils";

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
</script>
