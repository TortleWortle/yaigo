<script setup>
defineProps({
    user: {
        type: String,
        default: "",
    },
    age: {
        type: Number,
        default: 0,
    },
});
</script>

<template>
    <h1>Welcome {{ user }}</h1>
    <p>{{ age }}</p>
</template>
