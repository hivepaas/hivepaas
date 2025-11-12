module.exports = {
    root: true,
    env: {
        browser: true,
        es2020: true,
    },
    parserOptions: {
        project: ["./tsconfig.json"],
        tsconfigRootDir: __dirname,
        ecmaVersion: "latest",
        sourceType: "module",
    },
    parser: "@typescript-eslint/parser",
    plugins: [
        "@typescript-eslint", //
        "@stylistic/js",
        "@stylistic/ts",
        "react",
        "react-hooks",
        "react-refresh",
        "@tanstack/query",
        "jsx-a11y",
        "import",
        "prettier",
    ],
    extends: [
        "eslint:recommended",
        "plugin:@typescript-eslint/strict-type-checked",
        "plugin:@typescript-eslint/stylistic-type-checked",
        "plugin:react/recommended",
        "plugin:react-hooks/recommended",
        "plugin:@tanstack/eslint-plugin-query/recommended",
        "plugin:jsx-a11y/recommended",
        "plugin:import/recommended",
        "prettier",
    ],
    ignorePatterns: ["dist", ".eslintrc.cjs", "vite.config.ts"],
    rules: {
        "@typescript-eslint/no-unused-vars": [
            "error",
            {
                argsIgnorePattern: "^_",
                varsIgnorePattern: "^_",
                caughtErrorsIgnorePattern: "^_",
            },
        ],
        "@typescript-eslint/no-namespace": "off",
        "@typescript-eslint/prefer-namespace-keyword": "off",
        "@typescript-eslint/no-non-null-assertion": "off",
        "@typescript-eslint/explicit-function-return-type": "off",
        "@typescript-eslint/explicit-module-boundary-types": "error",
        "@typescript-eslint/no-inferrable-types": "off",
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

        "@stylistic/js/arrow-parens": ["error", "as-needed"],

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

        "react-refresh/only-export-components": "off",

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

        "import/named": "off",
        "import/no-named-as-default-member": "off",
        "import/extensions": [
            "error",
            {
                ignorePackages: true,
            },
        ],

        // TODO move rules to stylistic
        "spaced-comment": ["error", "always", { markers: ["/"] }],
        "no-trailing-spaces": "error",
        "no-shadow": "error",
        "linebreak-style": ["off"],
        "quotes": ["error", "double"],
        "quote-props": ["error", "consistent-as-needed"],
        "comma-dangle": ["error", "always-multiline"],
        "semi": ["error", "always"],
        "prefer-destructuring": ["error", { object: true, array: false }],
        "object-shorthand": ["error", "always"],
        "no-param-reassign": [
            "error",
            {
                props: true,
                ignorePropertyModificationsFor: [
                    "state", // redux
                    "draft", // immer
                ],
            },
        ],
        "object-curly-spacing": ["error", "always"],
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
    overrides: [
        {
            files: ["*.tsx", "*api.client.ts"],
            rules: {
                "@typescript-eslint/explicit-module-boundary-types": "off",
            },
        },
    ],
};
