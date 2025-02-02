// @ts-check
import { defineConfig } from 'astro/config';

// https://astro.build/config
export default defineConfig({
  vite: {
    server: {
      proxy: {
        '/v1': {
          target: 'http://localhost:4000/',
          changeOrigin: true,
          secure: false,
        },
      },
      allowedHosts: ['hiisi'],
    },
  },
});
