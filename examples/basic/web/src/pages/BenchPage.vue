<script setup>

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
  concurrentProp: {
    type: String,
    default: "",
  },
  concurrentProp2: {
    type: String,
    default: "",
  },
})
</script>

<template>
  <h1>Test Page</h1>
  <pre>
    {{ JSON.stringify(props, null, 2) }}
  </pre>
</template>
