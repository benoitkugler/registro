<template>
  <NavBar :title="`${controller.camp?.Label} - Equipiers`"> </NavBar>

  <v-card>
    <template #append>
      <v-btn @click="showDocuments = true">
        <template #prepend>
          <v-icon>mdi-file</v-icon>
        </template>
        Documents</v-btn
      >
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
                  :prepend-icon="formStatusIcon(equipier.Equipier.FormStatus)"
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
                      size="small"
                    ></v-btn>
                  </template>
                  <v-list>
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
  type Roles,
} from "../logic/api";
import DocumentsTable from "../components/equipiers/DocumentsTable.vue";

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

function formStatusIcon(s: FormStatusEquipier) {
  switch (s) {
    case FormStatusEquipier.NotSend:
      return "mdi-form-select";
    case FormStatusEquipier.Pending:
      return "mdi-clock";
    case FormStatusEquipier.Answered:
      return "mdi-checked";
  }
}
</script>
