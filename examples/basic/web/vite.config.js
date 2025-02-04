import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";

// https://vite.dev/config/
export default defineConfig({
  plugins: [vue()],
  resolve: (name) => {
    const pages = import.meta.glob("./pages/**/*.vue", { eager: true });
    return pages[`./pages/${name}.vue`];
  },
  build: {
    manifest: true,
    rollupOptions: {
      input: "src/main.js",
    },
  },
  server: {
    hmr: {
      host: "localhost",
    },
  },
});
