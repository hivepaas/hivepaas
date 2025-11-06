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
            $pattern: "auth/two-fa",
            $route: "/auth/two-fa/",
        },

        signIn: {
            $pattern: "auth/sign-in",
            $route: "/auth/sign-in/",
        },

        signUp: {
            $pattern: "auth/sign-up",
            $route: "/auth/sign-up/",
        },
    },
} as const;
