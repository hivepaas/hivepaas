import js from "@eslint/js";
import stylistic from "@stylistic/eslint-plugin";
import tanstackQuery from "@tanstack/eslint-plugin-query";
import prettierConfig from "eslint-config-prettier";
import importPlugin from "eslint-plugin-import";
import jsxA11y from "eslint-plugin-jsx-a11y";
import prettier from "eslint-plugin-prettier";
import react from "eslint-plugin-react";
import reactHooks from "eslint-plugin-react-hooks";
import reactRefresh from "eslint-plugin-react-refresh";
import globals from "globals";
import tseslint from "typescript-eslint";

export default tseslint.config(
    {
        ignores: [
            "dist",
            ".eslintrc.cjs",
            "vite.config.ts",
            "vite.config.js",
            "eslint.config.js",
            "node_modules",
            "coverage",
            "src/infrastructure/api/be-api/**",
            "src/components/ui/**",
            "openapi-ts.config.ts",
            "vite.config.d.ts",
        ],
    },
    js.configs.recommended,
    ...tseslint.configs.strictTypeChecked.map(config => ({
        ...config,
        files: ["**/*.{ts,tsx}"],
    })),
    ...tseslint.configs.stylisticTypeChecked.map(config => ({
        ...config,
        files: ["**/*.{ts,tsx}"],
    })),
    {
        files: ["**/*.{ts,tsx}"],
        plugins: {
            react,
            "react-hooks": reactHooks,
            "react-refresh": reactRefresh,
            "import": importPlugin,
            prettier,
            "@stylistic": stylistic,
            "@tanstack/query": tanstackQuery,
            "jsx-a11y": jsxA11y,
        },
        languageOptions: {
            ecmaVersion: "latest",
            sourceType: "module",
            globals: {
                ...globals.browser,
                ...globals.es2020,
            },
            parserOptions: {
                project: ["./tsconfig.app.json"],
                tsconfigRootDir: import.meta.dirname,
            },
        },
        settings: {
            "react": {
                version: "detect",
            },
            "import/parsers": {
                "@typescript-eslint/parser": [".ts", ".tsx"],
            },
            "import/resolver": {
                typescript: {
                    alwaysTryTypes: true,
                    project: "tsconfig.json",
                },
            },
        },
        rules: {
            ...prettierConfig.rules,
            // TypeScript rules
            "@typescript-eslint/no-namespace": "off",
            "@typescript-eslint/prefer-namespace-keyword": "off",
            "@typescript-eslint/no-non-null-assertion": "off",
            "@typescript-eslint/explicit-function-return-type": "off",
            "@typescript-eslint/explicit-module-boundary-types": "error",
            "@typescript-eslint/no-inferrable-types": "off",
            "@typescript-eslint/no-unused-vars": [
                "warn",
                {
                    argsIgnorePattern: "^_",
                    ignoreRestSiblings: true,
                },
            ],
            "@typescript-eslint/typedef": [
                "error",
                {
                    arrowParameter: false,
                    propertyDeclaration: true,
                    parameter: true,
                },
            ],
            "@typescript-eslint/consistent-type-imports": [
                "error",
                {
                    prefer: "type-imports",
                    disallowTypeAnnotations: false,
                    fixStyle: "inline-type-imports",
                },
            ],
            "@typescript-eslint/restrict-template-expressions": ["error", { allowNumber: true }],
            "@typescript-eslint/consistent-type-definitions": "off",
            "@stylistic/arrow-parens": ["error", "as-needed"],

            // React & Hooks
            ...react.configs.recommended.rules,
            ...reactHooks.configs.recommended.rules,
            ...tanstackQuery.configs.recommended.rules,
            ...jsxA11y.configs.recommended.rules,
            "react/react-in-jsx-scope": "off", // Not needed in React 17+
            "react/jsx-uses-react": "off", // Not needed in React 17+
            "react/jsx-indent": ["error", 4],
            "react/jsx-indent-props": ["error", 4],
            "react/jsx-first-prop-new-line": ["error", "multiline"],
            "react/jsx-max-props-per-line": ["error", { maximum: 1, when: "multiline" }],
            "react/jsx-boolean-value": ["error", "never", { always: ["personal"] }],
            "react/jsx-wrap-multilines": "error",
            "react/jsx-child-element-spacing": "error",
            "react/jsx-curly-brace-presence": ["error", { props: "never", children: "never" }],
            "react/jsx-no-useless-fragment": "error",
            "react/self-closing-comp": "error",
            "react/void-dom-elements-no-children": "error",
            "react/no-danger": "error",
            "react/no-unused-prop-types": "error",
            "react/default-props-match-prop-types": ["error", { allowRequiredDefaults: false }],
            "react/button-has-type": "error",
            "react/function-component-definition": ["error", { namedComponents: "function-declaration" }],
            "react-hooks/exhaustive-deps": "error",
            "react-refresh/only-export-components": ["warn", { allowConstantExport: true }],

            // a11y rules
            "jsx-a11y/click-events-have-key-events": "off",
            "jsx-a11y/control-has-associated-label": "off",
            "jsx-a11y/label-has-associated-control": [
                "error",
                {
                    required: {
                        some: ["nesting", "id"],
                    },
                },
            ],
            "jsx-a11y/label-has-for": [
                "error",
                {
                    required: {
                        some: ["nesting", "id"],
                    },
                },
            ],

            // Import rules
            "import/named": "off",
            "import/no-named-as-default-member": "off",
            "import/extensions": ["error", { ignorePackages: true }],

            // Restricted imports
            "no-restricted-imports": ["error"],

            "no-restricted-syntax": ["error", { selector: "TSEnumDeclaration", message: "Don't declare enums" }],

            // Style rules
            "spaced-comment": ["error", "always", { markers: ["/"] }],
            "no-trailing-spaces": "error",
            "no-shadow": "error",
            "linebreak-style": "off",
            "quotes": ["error", "double"],
            "quote-props": ["error", "consistent-as-needed"],
            "comma-dangle": ["error", "always-multiline"],
            "semi": ["error", "always"],
            "prefer-destructuring": ["error", { object: true, array: false }],
            "object-shorthand": ["error", "always"],
            "dot-notation": "off",
            "@typescript-eslint/dot-notation": "off",
            "no-param-reassign": [
                "error",
                {
                    props: true,
                    ignorePropertyModificationsFor: ["state", "draft"],
                },
            ],
            "object-curly-spacing": ["error", "always"],
        },
    },
    {
        files: ["*.d.ts"],
        rules: {
            "@typescript-eslint/restrict-template-expressions": "off",
            "@typescript-eslint/explicit-module-boundary-types": "off",
        },
    },
    {
        files: ["**/*.tsx", "**/*api.client.ts"],
        rules: {
            "@typescript-eslint/explicit-module-boundary-types": "off",
        },
    },
);
