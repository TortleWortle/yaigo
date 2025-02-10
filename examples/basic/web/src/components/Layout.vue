<script setup>
import { Link } from "@inertiajs/vue3";
</script>
<template>
    <nav>
        <Link href="/">Home</Link>
        <Link href="/news">News</Link>
        <Link href="/about">About</Link>
        <Link href="/jeff">Jeff (normal api)</Link>
        <Link href="/geoffrey">Geoffrey (helper api)</Link>
        <Link href="/404">404</Link>
        <Link href="/test">Test</Link>
        <Link href="/clear_history">Clear history</Link>
        <Link href="/brokenprop">Broken prop</Link>
        <Link href="/benchme">Bench me</Link>
        <Link href="/redirect">Redirect</Link>
        <Link href="/location">External link</Link>
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
