import reactHooks from "eslint-plugin-react-hooks";
import { defineConfig } from "eslint/config";
import { dirname } from "node:path";
import { fileURLToPath } from "node:url";
import tseslint from "typescript-eslint";

const __dirname = dirname(fileURLToPath(import.meta.url));

export default defineConfig(
  {
    ignores: ["**/*.js", "**/*.cjs"],
  },
  reactHooks.configs.flat["recommended-latest"],
  {
    files: ["**/*.{js,ts,jsx,tsx}"],
    ignores: ["**/.next/**", "components/shadcn/**"],

    languageOptions: {
      parser: tseslint.parser,
      parserOptions: {
        projectService: true,
        tsconfigRootDir: __dirname,
      },
    },
    plugins: {
      "@typescript-eslint": tseslint.plugin,
    },
    rules: {
      "@typescript-eslint/no-unnecessary-condition": "warn",
      "@typescript-eslint/no-unnecessary-type-assertion": "warn",
    },
  },
);
