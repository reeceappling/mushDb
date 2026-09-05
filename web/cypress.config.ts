import { defineConfig } from "cypress";

export default defineConfig({
  component: {
    devServer: {
      framework: 'next',
      bundler: 'webpack' // TODO: turbopack
    }
  },
  e2e: {
    setupNodeEvents(on, config) {
      // implement node event listeners here
    },
  },
});
