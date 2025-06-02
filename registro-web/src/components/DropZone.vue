<template>
  <!-- drop zone -->
  <div
    v-if="props.isDragging"
    style="height: 15px"
    :class="{
      'bg-blue-lighten-1': isOver,
      'bg-blue-lighten-4': !isOver,
      rounded: true,
    }"
    @dragenter="onDragEnter"
    @dragover="(ev) => ev.preventDefault()"
    @dragleave="isOver = false"
    @drop="onDrop"
  ></div>
  <div v-else style="height: 15px"></div>
</template>

<script lang="ts" setup>
import { ref } from "vue";

const props = defineProps<{
  isDragging: boolean;
}>();

const emit = defineEmits<{
  (e: "onDrop", event: DragEvent): void;
}>();

const isOver = ref(false);

function onDragEnter(event: DragEvent) {
  isOver.value = true;
  event.preventDefault();
}

function onDrop(event: DragEvent) {
  isOver.value = false;
  event.preventDefault();
  emit("onDrop", event);
}
</script>
