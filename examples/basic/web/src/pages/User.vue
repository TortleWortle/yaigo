<script setup>
const props = defineProps({
    user: {
        type: String,
        default: "",
    },
    age: {
        type: Number,
        default: 0,
    },
    slow: {
      type: String,
      default: "",
    },
});
</script>

<template>
    <h1>Welcome {{ user }}</h1>
    <pre>
      {{ JSON.stringify(props, null, 2) }}
    </pre>
</template>
