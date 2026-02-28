export const ROUTE = {
    /**
     * Authentication
     */
    auth: {
        sso: {
            success: {
                $pattern: "auth/sso/success",
                $route: "/auth/sso/success/",
            },
        },

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

    /**
     * Users
     */
    userManagement: {
        $pattern: "user-management",

        users: {
            $pattern: "user-management/users",
            $route: "/user-management/users/",

            single: {
                $pattern: "user-management/users/:id",
                $route: (id: string) => `/user-management/users/${id}/`,
            },
        },
    },

    /**
     * Current User
     */
    currentUser: {
        $pattern: "current-user",

        profile: {
            $pattern: "current-user/profile",
            $route: "/current-user/profile/",
        },
        profileApiKeys: {
            $pattern: "current-user/profile/api-keys",
            $route: "/current-user/profile/api-keys/",
        },
    },

    /**
     * Clusters
     */
    cluster: {
        $pattern: "cluster",

        nodes: {
            $pattern: "cluster/nodes",
            $route: "/cluster/nodes/",

            single: {
                $pattern: "cluster/nodes/:id",
                $route: (id: string) => `/cluster/nodes/${id}/`,
            },
        },
    },

    /**
     * Projects
     */
    projects: {
        $pattern: "projects",

        list: {
            $pattern: "projects",
            $route: "/projects/",
        },

        single: {
            $pattern: "projects/:id",

            apps: {
                $pattern: "projects/:id/apps",
                $route: (id: string) => `/projects/${id}/apps/`,

                single: {
                    $pattern: "projects/:id/apps/:appId",

                    configuration: {
                        $pattern: "projects/:id/apps/:appId/configuration",

                        general: {
                            $pattern: "projects/:id/apps/:appId/general",
                            $route: (id: string, appId: string) => `/projects/${id}/apps/${appId}/general/`,
                        },

                        deploymentSettings: {
                            $pattern: "projects/:id/apps/:appId/deployment-settings",
                            $route: (id: string, appId: string) => `/projects/${id}/apps/${appId}/deployment-settings/`,
                        },

                        containerSettings: {
                            $pattern: "projects/:id/apps/:appId/container-settings",
                            $route: (id: string, appId: string) => `/projects/${id}/apps/${appId}/container-settings/`,
                        },

                        httpSettings: {
                            $pattern: "projects/:id/apps/:appId/http-settings",
                            $route: (id: string, appId: string) => `/projects/${id}/apps/${appId}/http-settings/`,
                        },

                        envVariables: {
                            $pattern: "projects/:id/apps/:appId/env-variables",
                            $route: (id: string, appId: string) => `/projects/${id}/apps/${appId}/env-variables/`,
                        },

                        secrets: {
                            $pattern: "projects/:id/apps/:appId/secrets",
                            $route: (id: string, appId: string) => `/projects/${id}/apps/${appId}/secrets/`,
                        },

                        availabilityAndScaling: {
                            $pattern: "projects/:id/apps/:appId/availability-and-scaling",
                            $route: (id: string, appId: string) =>
                                `/projects/${id}/apps/${appId}/availability-and-scaling/`,
                        },

                        presistentStorage: {
                            $pattern: "projects/:id/apps/:appId/presistent-storage",
                            $route: (id: string, appId: string) => `/projects/${id}/apps/${appId}/presistent-storage/`,
                        },

                        networks: {
                            $pattern: "projects/:id/apps/:appId/networks",
                            $route: (id: string, appId: string) => `/projects/${id}/apps/${appId}/networks/`,
                        },

                        resources: {
                            $pattern: "projects/:id/apps/:appId/resources",
                            $route: (id: string, appId: string) => `/projects/${id}/apps/${appId}/resources/`,
                        },

                        dangerZone: {
                            $pattern: "projects/:id/apps/:appId/danger-zone",
                            $route: (id: string, appId: string) => `/projects/${id}/apps/${appId}/danger-zone/`,
                        },
                    },

                    deployments: {
                        $pattern: "projects/:id/apps/:appId/deployments",
                        $route: (id: string, appId: string) => `/projects/${id}/apps/${appId}/deployments/`,
                    },

                    logs: {
                        $pattern: "projects/:id/apps/:appId/logs",
                        $route: (id: string, appId: string) => `/projects/${id}/apps/${appId}/logs/`,
                    },

                    previewDeployments: {
                        $pattern: "projects/:id/apps/:appId/preview-deployments",
                        $route: (id: string, appId: string) => `/projects/${id}/apps/${appId}/preview-deployments/`,
                    },
                },
            },

            configuration: {
                $pattern: "projects/:id/configuration",

                general: {
                    $pattern: "projects/:id/configuration/general",
                    $route: (id: string) => `/projects/${id}/configuration/general/`,
                },

                envVariables: {
                    $pattern: "projects/:id/configuration/env-variables",
                    $route: (id: string) => `/projects/${id}/configuration/env-variables/`,
                },

                secrets: {
                    $pattern: "projects/:id/configuration/secrets",
                    $route: (id: string) => `/projects/${id}/configuration/secrets/`,
                },
            },
        },
    },
} as const;
