<script setup>
import {router, Deferred} from "@inertiajs/vue3";

const props = defineProps({
  inlineProp: {
    type: String,
    default: "",
  },
  helperProp: {
    type: String,
    default: 0,
  },
  time: {
    type: String,
    default: "",
  },
  deferredProp: {
    type: String,
    default: "",
  }
});

function reloadTime() {
  router.reload({only: ['time']})
}

function reloadAllExceptTime() {
  router.reload({except: ['time']})
}
</script>

<template>
  <h1>Test Page</h1>
  <pre>
    {{ JSON.stringify(props, null, 2) }}
  </pre>
  <Deferred data="deferredProp">
    <template #fallback>
      <div>Loading...</div>
    </template>

    <div>
      {{ deferredProp }}
    </div>
  </Deferred>
  <button @click="reloadTime()">Reload time</button>
  <button @click="reloadAllExceptTime()">Reload all except time</button>
</template>
