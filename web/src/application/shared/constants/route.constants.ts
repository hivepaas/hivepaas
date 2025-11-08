export const ROUTE = {
    /**
     * Modules
     */
    modules: {
        /**
         * Users & Roles
         */
        usersAndRoles: {
            $pattern: "users-and-roles",

            users: {
                $pattern: "users-and-roles/users",
                $route: "/users-and-roles/users/",
            },
        },
    },
    /**
     * Authentication
     */
    auth: {
        twoFA: {
            $pattern: "auth/2fa",
            $route: "/auth/2fa/",
        },

        signIn: {
            $pattern: "auth/sign-in",
            $route: "/auth/sign-in/",
        },

        signUp: {
            $pattern: "auth/sign-up",
            $route: "/auth/sign-up/",
        },
        resetPassword: {
            $pattern: "auth/reset-password",
            $route: "/auth/reset-password/",
        },
        forgotPassword: {
            $pattern: "auth/forgot-password",
            $route: "/auth/forgot-password/",
        },
    },
} as const;
