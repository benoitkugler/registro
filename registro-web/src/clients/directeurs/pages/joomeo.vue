<template>
  <NavBar :title="`${controller.camp?.Label} - Album Joomeo`"> </NavBar>

  <div v-if="data == null" class="text-center my-6">
    <v-progress-circular indeterminate></v-progress-circular>
  </div>
  <v-alert v-else-if="data.SpaceURL == ''" type="warning" class="ma-2"
    >Aucun album n'est associé au séjour.</v-alert
  >
  <v-card v-else title="Album et contacts" class="ma-2">
    <template #append>
      <v-btn @click="showInvite = true" :disabled="isInviting">
        <template #prepend>
          <v-progress-circular
            indeterminate
            v-if="isInviting"
            size="small"
          ></v-progress-circular>
          <v-icon v-else color="green">mdi-plus</v-icon>
        </template>
        Inviter...
      </v-btn>
    </template>
    <v-card-text>
      <v-alert density="comfortable">
        <v-row>
          <v-col align-self="center"
            >Nom de l'album : <b>{{ data.Album.Label }}</b></v-col
          >
          <v-col align-self="center"
            >Créé le : <b>{{ Formatters.date(data.Album.Date) }}</b></v-col
          >
          <v-col align-self="center"
            >Nombre de photos : <b>{{ data.Album.FilesCount }}</b></v-col
          >
          <v-col align-self="center" cols="auto">
            <v-btn
              size="small"
              :href="data.SpaceURL"
              link
              target="_blank"
              class="mx-1"
              prepend-icon="mdi-open-in-new"
            >
              Accéder à Joomeo</v-btn
            >
          </v-col>
        </v-row>
      </v-alert>

      <v-list max-width="800px" class="mx-auto">
        <v-list-item
          v-for="contact in list"
          :title="contact.login"
          :subtitle="contact.email"
        >
          <template #append>
            <v-row>
              <v-col align-self="center">
                <v-chip prepend-icon="mdi-key">{{ contact.password }}</v-chip>
              </v-col>
              <v-col align-self="center">
                <v-chip
                  v-if="contact.albumAccessRules.allowUpload"
                  prepend-icon="mdi-account-arrow-up"
                  color="blue"
                >
                  Droit d'ajout
                </v-chip>
              </v-col>
              <v-col align-self="center" cols="auto">
                <v-btn icon size="small" flat>
                  <v-icon>mdi-dots-vertical</v-icon>
                  <v-menu activator="parent">
                    <v-list dense>
                      <v-list-item
                        title="Permettre l'ajout"
                        subtitle="Donne le droit d'ajouter des photos"
                        prepend-icon="mdi-account-arrow-up"
                        @click="setUploader(contact.contactid)"
                      ></v-list-item>
                      <v-list-item
                        title="Supprimer"
                        subtitle="Retire l'accès à l'album du séjour"
                        prepend-icon="mdi-delete"
                        @click="unlink(contact.contactid)"
                      ></v-list-item>
                    </v-list>
                  </v-menu>
                </v-btn>
              </v-col>
            </v-row>
          </template>
        </v-list-item>
      </v-list>
    </v-card-text>

    <v-dialog v-model="showInvite" max-width="800px">
      <v-card title="Ajouter des contacts à l'album">
        <v-card-text>
          <v-row>
            <v-col>
              <v-switch
                color="primary"
                label="Inclure les responsables"
                v-model="addResponsables"
              ></v-switch>
              <v-switch
                color="primary"
                label="Inclure les inscrits"
                v-model="addInscrits"
              ></v-switch>
              <v-switch
                color="primary"
                label="Inclure les équipiers"
                v-model="addEquipiers"
              ></v-switch>
              <v-text-field
                variant="outlined"
                density="comfortable"
                label="Ajouter une adresse"
                v-model="customMail"
                :rules="[mailRule]"
              >
                <template #append-inner>
                  <v-btn
                    size="small"
                    icon="mdi-plus"
                    @click="
                      otherMails.push(customMail!);
                      customMail = '';
                    "
                    :disabled="customMail == '' || mailRule(customMail) != true"
                  ></v-btn>
                </template>
              </v-text-field>
            </v-col>
            <v-col>
              <v-list>
                <v-list-subheader>Adresses mails</v-list-subheader>
                <v-list-item
                  v-for="mail in allSelectedMails"
                  :title="mail.mail"
                  :subtitle="mail.kind"
                >
                  <template #prepend v-if="mailRule(mail.mail) != true">
                    <v-icon color="orange">mdi-warning</v-icon>
                  </template>
                  <template #append v-if="mail.kind == 'Autre'">
                    <v-btn
                      size="small"
                      flat
                      icon
                      @click="
                        otherMails = otherMails.filter((m) => m != mail.mail)
                      "
                    >
                      <v-icon color="red">mdi-delete</v-icon>
                    </v-btn>
                  </template>
                </v-list-item>
              </v-list>
            </v-col>
          </v-row>
        </v-card-text>
        <v-card-actions>
          <v-btn @click="invite(false)" :disabled="!allSelectedMails.length">
            Inviter sans envoi de mail
          </v-btn>
          <v-spacer></v-spacer>
          <v-btn
            color="green"
            @click="invite(true)"
            :disabled="!allSelectedMails.length"
            >Inviter par mail</v-btn
          >
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-card>
</template>

<script lang="ts" setup>
import { computed, onMounted, ref } from "vue";
import NavBar from "../components/NavBar.vue";
import type { Joomeo } from "../logic/api";
import { controller } from "../logic/logic";
import { copy, Formatters, FormRules } from "@/utils";

onMounted(loadData);

const list = computed(() => {
  const l = copy(data.value?.Album.Contacts || []);
  return l.sort((a, b) =>
    a.albumAccessRules.allowUpload == b.albumAccessRules.allowUpload
      ? a.login.localeCompare(b.login)
      : compareBool(
          a.albumAccessRules.allowUpload,
          b.albumAccessRules.allowUpload
        )
  );
});

function compareBool(a: boolean, b: boolean) {
  if (a && !b) {
    return -1;
  }
  if (!a && b) {
    return 1;
  }
  return 0;
}

const data = ref<Joomeo | null>(null);
async function loadData() {
  const res = await controller.JoomeoLoad();
  if (res === undefined) return;

  data.value = res;
}

const showInvite = ref(false);
const addResponsables = ref(false);
const addInscrits = ref(false);
const addEquipiers = ref(false);
const otherMails = ref<string[]>([]);
const customMail = ref<string>("");

type kind = "Responsable" | "Inscrit" | "Equipier" | "Autre";

const allSelectedMails = computed(() => {
  const out: { mail: string; kind: kind }[] = otherMails.value.map((p) => ({
    mail: p,
    kind: "Autre",
  }));
  if (addResponsables.value)
    out.push(
      ...((data.value?.MailsResponsables || []).map((p) => ({
        mail: p,
        kind: "Responsable",
      })) satisfies typeof out)
    );
  if (addInscrits.value)
    out.push(
      ...((data.value?.MailsInscrits || []).map((p) => ({
        mail: p,
        kind: "Inscrit",
      })) satisfies typeof out)
    );
  if (addEquipiers.value)
    out.push(
      ...((data.value?.MailsEquipiers || []).map((p) => ({
        mail: p,
        kind: "Equipier",
      })) satisfies typeof out)
    );
  return out;
});

const mailRule = FormRules.mailJoomeo();

const isInviting = ref(false);
async function invite(sendMail: boolean) {
  if (!data.value) return;
  const mails = allSelectedMails.value
    .filter((m) => mailRule(m.mail) == true)
    .map((m) => m.mail);
  if (!mails.length) return;

  showInvite.value = false;
  isInviting.value = true;
  const res = await controller.JoomeoInvite({
    Mails: mails,
    SendMail: sendMail,
  });
  isInviting.value = false;
  if (res === undefined) return;
  controller.showMessage("Contacts invités avec succès.");
  data.value!.Album.Contacts = res || [];
}

async function setUploader(joomeoId: string) {
  const l = data.value?.Album.Contacts || [];
  if (!l) return;
  const res = await controller.JoomeoSetUploader({ joomeoId });
  if (res === undefined) return;
  const index = l.findIndex((rec) => rec.contactid == joomeoId);
  l[index] = res;
  controller.showMessage("Permissions accordées avec succès.");
}

async function unlink(joomeoId: string) {
  const l = data.value?.Album.Contacts || [];
  if (!l) return;
  const res = await controller.JoomeoUnlinkContact({ joomeoId });
  if (res === undefined) return;
  const index = l.findIndex((rec) => rec.contactid == joomeoId);
  l.splice(index, 1);
  controller.showMessage("Contact retiré avec succès.");
}
</script>
