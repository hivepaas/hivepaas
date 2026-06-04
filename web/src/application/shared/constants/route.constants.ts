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
     * Cluster
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

        networks: {
            $pattern: "cluster/networks",
            $route: "/cluster/networks/",
        },
    },

    /**
     * Sources
     */
    sources: {
        $pattern: "sources",
        $route: "/sources/github-apps/",

        githubApps: {
            $pattern: "sources/github-apps",
            $route: "/sources/github-apps/",
        },

        webhooks: {
            $pattern: "sources/webhooks",
            $route: "/sources/webhooks/",
        },
    },

    /**
     * Settings
     */
    settings: {
        $pattern: "settings",

        basicAuth: {
            $pattern: "settings/basic-auth",
            $route: "/settings/basic-auth/",
        },

        registryAuth: {
            $pattern: "settings/registry-auth",
            $route: "/settings/registry-auth/",
        },

        sslCertificates: {
            $pattern: "settings/ssl-certificates",
            $route: "/settings/ssl-certificates/",
        },

        emailAccounts: {
            $pattern: "settings/email-accounts",
            $route: "/settings/email-accounts/",
        },

        imPlatforms: {
            $pattern: "settings/im-platforms",
            $route: "/settings/im-platforms/",
        },

        sshKeys: {
            $pattern: "settings/ssh-keys",
            $route: "/settings/ssh-keys/",
        },

        accessTokens: {
            $pattern: "settings/access-tokens",
            $route: "/settings/access-tokens/",
        },

        cloudStorages: {
            $pattern: "settings/cloud-storages",
            $route: "/settings/cloud-storages/",
        },

        oauth: {
            $pattern: "settings/oauth",
            $route: "/settings/oauth/",
        },

        notificationTargets: {
            $pattern: "settings/notification-targets",
            $route: "/settings/notification-targets/",
        },
    },

    /**
     * System Settings
     */
    systemSettings: {
        $pattern: "system-settings",

        localpaas: {
            $pattern: "system-settings/localpaas",
            $route: "/system-settings/localpaas/general/",

            general: {
                $pattern: "system-settings/localpaas/general",
                $route: "/system-settings/localpaas/general/",
            },
        },

        traefik: {
            $pattern: "system-settings/traefik",
            $route: "/system-settings/traefik/general/",

            general: {
                $pattern: "system-settings/traefik/general",
                $route: "/system-settings/traefik/general/",
            },
        },

        dataBackup: {
            $pattern: "system-settings/data-backup",
            $route: "/system-settings/data-backup/configuration/",

            configuration: {
                $pattern: "system-settings/data-backup/configuration",
                $route: "/system-settings/data-backup/configuration/",
            },

            backupFiles: {
                $pattern: "system-settings/data-backup/backup-files",
                $route: "/system-settings/data-backup/backup-files/",
            },

            actions: {
                $pattern: "system-settings/data-backup/actions",
                $route: "/system-settings/data-backup/actions/",
            },
        },

        dataCleanup: {
            $pattern: "system-settings/data-cleanup",
            $route: "/system-settings/data-cleanup/configuration/",

            configuration: {
                $pattern: "system-settings/data-cleanup/configuration",
                $route: "/system-settings/data-cleanup/configuration/",
            },

            actions: {
                $pattern: "system-settings/data-cleanup/actions",
                $route: "/system-settings/data-cleanup/actions/",
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

                        healthChecks: {
                            $pattern: "projects/:id/apps/:appId/health-checks",
                            $route: (id: string, appId: string) => `/projects/${id}/apps/${appId}/health-checks/`,
                        },

                        scheduledJobs: {
                            $pattern: "projects/:id/apps/:appId/sched-jobs",
                            $route: (id: string, appId: string) => `/projects/${id}/apps/${appId}/sched-jobs/`,
                        },

                        envVariables: {
                            $pattern: "projects/:id/apps/:appId/env-variables",
                            $route: (id: string, appId: string) => `/projects/${id}/apps/${appId}/env-variables/`,
                        },

                        secrets: {
                            $pattern: "projects/:id/apps/:appId/secrets",
                            $route: (id: string, appId: string) => `/projects/${id}/apps/${appId}/secrets/`,
                        },

                        configFiles: {
                            $pattern: "projects/:id/apps/:appId/config-files",
                            $route: (id: string, appId: string) => `/projects/${id}/apps/${appId}/config-files/`,
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
                $route: (id: string) => `/projects/${id}/configuration/`,

                general: {
                    $pattern: "projects/:id/configuration/general",
                    $route: (id: string) => `/projects/${id}/configuration/general/`,
                },

                buildSettings: {
                    $pattern: "projects/:id/configuration/build-settings",
                    $route: (id: string) => `/projects/${id}/configuration/build-settings/`,
                },

                storageSettings: {
                    $pattern: "projects/:id/configuration/storage-settings",
                    $route: (id: string) => `/projects/${id}/configuration/storage-settings/`,
                },

                domainSettings: {
                    $pattern: "projects/:id/configuration/domain-settings",
                    $route: (id: string) => `/projects/${id}/configuration/domain-settings/`,
                },

                dangerZone: {
                    $pattern: "projects/:id/configuration/danger-zone",
                    $route: (id: string) => `/projects/${id}/configuration/danger-zone/`,
                },
            },

            providerConfiguration: {
                $pattern: "projects/:id/provider-configuration",
                $route: (id: string) => `/projects/${id}/provider-configuration/`,

                accessTokens: {
                    $pattern: "projects/:id/provider-configuration/access-tokens",
                    $route: (id: string) => `/projects/${id}/provider-configuration/access-tokens/`,
                },

                basicAuth: {
                    $pattern: "projects/:id/provider-configuration/basic-auth",
                    $route: (id: string) => `/projects/${id}/provider-configuration/basic-auth/`,
                },

                cloudStorages: {
                    $pattern: "projects/:id/provider-configuration/cloud-storages",
                    $route: (id: string) => `/projects/${id}/provider-configuration/cloud-storages/`,
                },

                emailAccounts: {
                    $pattern: "projects/:id/provider-configuration/email-accounts",
                    $route: (id: string) => `/projects/${id}/provider-configuration/email-accounts/`,
                },

                envVariables: {
                    $pattern: "projects/:id/provider-configuration/env-variables",
                    $route: (id: string) => `/projects/${id}/provider-configuration/env-variables/`,
                },

                githubApps: {
                    $pattern: "projects/:id/provider-configuration/github-apps",
                    $route: (id: string) => `/projects/${id}/provider-configuration/github-apps/`,
                },

                webhooks: {
                    $pattern: "projects/:id/provider-configuration/webhooks",
                    $route: (id: string) => `/projects/${id}/provider-configuration/webhooks/`,
                },

                imPlatforms: {
                    $pattern: "projects/:id/provider-configuration/im-platforms",
                    $route: (id: string) => `/projects/${id}/provider-configuration/im-platforms/`,
                },

                notificationTargets: {
                    $pattern: "projects/:id/provider-configuration/notification-targets",
                    $route: (id: string) => `/projects/${id}/provider-configuration/notification-targets/`,
                },

                registryAuth: {
                    $pattern: "projects/:id/provider-configuration/registry-auth",
                    $route: (id: string) => `/projects/${id}/provider-configuration/registry-auth/`,
                },

                secrets: {
                    $pattern: "projects/:id/provider-configuration/secrets",
                    $route: (id: string) => `/projects/${id}/provider-configuration/secrets/`,
                },

                sshKeys: {
                    $pattern: "projects/:id/provider-configuration/ssh-keys",
                    $route: (id: string) => `/projects/${id}/provider-configuration/ssh-keys/`,
                },

                sslCertificates: {
                    $pattern: "projects/:id/provider-configuration/ssl-certificates",
                    $route: (id: string) => `/projects/${id}/provider-configuration/ssl-certificates/`,
                },
            },

            clusterResources: {
                $pattern: "projects/:id/cluster-resources",
                $route: (id: string) => `/projects/${id}/cluster-resources/`,

                networks: {
                    $pattern: "projects/:id/cluster-resources/networks",
                    $route: (id: string) => `/projects/${id}/cluster-resources/networks/`,
                },
            },
        },
    },
} as const;
