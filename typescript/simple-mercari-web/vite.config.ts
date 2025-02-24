import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    port: 3000,
    hmr: {
      port: 3000, // host machine port mapping
    },
  },
  resolve: {
    alias: [{ find: '~', replacement: '/src' }],
  },
});
