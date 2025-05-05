import { reactRouter } from '@react-router/dev/vite';
import { defineConfig } from 'vite';
import tsconfigPaths from 'vite-tsconfig-paths';

export default defineConfig({
  plugins: [reactRouter(), tsconfigPaths()],
  server: {
    proxy: {
      '/v1': {
        target: 'http://localhost:4000/',
        changeOrigin: true,
        secure: false,
      },
    },
    allowedHosts: ['sunny-oddly-ape.ngrok-free.app'],
  },
});
