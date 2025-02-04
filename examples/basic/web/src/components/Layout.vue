<script setup>
import { Link } from "@inertiajs/vue3";
</script>
<template>
    <nav>
        <Link href="/">Home</Link>
        <Link href="/news">News</Link>
        <Link href="/about">About</Link>
        <Link href="/jeff">Jeff</Link>
        <Link href="/test">Test</Link>
    </nav>
    <div>
        <slot />
    </div>
</template>

<style scoped>
nav {
    display: flex;
    gap: 2rem;
}
</style>
