<template>
  <v-row class="my-1">
    <v-col align-self="center" cols="1" class="text-center">
      <v-tooltip
        :text="isNew ? 'Marquer comme lu' : 'Marquer comme non lu'"
        location="left"
        v-if="!isFromUs"
      >
        <template #activator="{ props: tooltipProps }">
          <v-btn
            v-bind="tooltipProps"
            size="small"
            icon
            flat
            @click="emit('setSeen', isNew)"
          >
            <v-icon
              :color="isNew ? 'pink' : ''"
              :icon="isNew ? 'mdi-message-badge' : 'mdi-message-check'"
            ></v-icon>
          </v-btn>
        </template>
      </v-tooltip>
    </v-col>
    <v-col
      align-self="center"
      :class="{
        'rounded-lg': true,
        'font-weight-bold': isNew,
        'bg-purple-lighten-4': isFromUs,
        'bg-green-lighten-4': isFromBackoffice,
        'bg-blue-lighten-4': !(isFromUs || isFromBackoffice),
      }"
      style="font-size: smaller"
    >
      <pre>{{ props.message.Content.Message.Contenu }}</pre>
    </v-col>
    <v-col align-self="center" cols="2" class="text-grey">
      {{ Formatters.time(props.message.Event.Created) }}
    </v-col>
  </v-row>
</template>

<script lang="ts" setup>
import { Formatters } from "@/utils";
import { Acteur, type EventExt_MessageEvt } from "../../logic/api";
import { isMessageFromUs, isMessageNew } from "../../logic/logic";
import { computed } from "vue";

const props = defineProps<{
  message: EventExt_MessageEvt;
}>();

const emit = defineEmits<{
  (e: "setSeen", seen: boolean): void;
  (e: "startReply"): void;
}>();

// contrary of isSeen
const isNew = computed(() => isMessageNew(props.message));
const isFromUs = computed(() => isMessageFromUs(props.message));
const isFromBackoffice = computed(() => {
  const o = props.message.Content.Message.Origine;
  return o == Acteur.Backoffice || o == Acteur.FondSoutien;
});
</script>
