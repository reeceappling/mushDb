import { dirname } from "node:path";
import { fileURLToPath } from "node:url";
import { FlatCompat } from "@eslint/eslintrc";
import unusedImports from "eslint-plugin-unused-imports";

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

const compat = new FlatCompat({
  baseDirectory: __dirname,
});

const eslintConfig = [
  //...compat.extends("next/core-web-vitals", "next/typescript"), // TODO: reenable?
];

export default [
  ...compat.extends("next/core-web-vitals", "next/typescript"),
  {
    plugins: {
      "unused-imports": unusedImports,
    },
    rules: {
      // Turn off some things
      '@typescript-eslint/no-unused-expressions': "off",
      "@typescript-eslint/no-explicit-any": "off",
      "no-unused-vars": "off",
      "@typescript-eslint/no-unused-vars": "off",
      "@typescript-eslint/no-empty-object-type": "off",
      "react-hooks/rules-of-hooks": "off", // TODO: reenable
      "react-hooks/exhaustive-deps": "off", // TODO: reenable?
      "@typescript-eslint/no-unsafe-declaration-merging": "off",
      "unused-imports/no-unused-imports": "error",
      "unused-imports/no-unused-vars": "off",
      "@next/next/no-img-element": "off", // TODO: reenable and figure out?
      // "unused-imports/no-unused-vars": [
      //   "warn",
      //   {
      //     vars: "all",
      //     args: "after-used",
      //     varsIgnorePattern: "^_",
      //     argsIgnorePattern: "^_",
      //   },
      // ],
    },
  },
];
