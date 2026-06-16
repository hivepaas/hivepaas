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
            $pattern: "current-user/api-keys",
            $route: "/current-user/api-keys/",
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

        sslProviders: {
            $pattern: "settings/ssl-providers",
            $route: "/settings/ssl-providers/",
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

        sslRenewal: {
            $pattern: "system-settings/ssl-renewal",
            $route: "/system-settings/ssl-renewal/configuration/",

            configuration: {
                $pattern: "system-settings/ssl-renewal/configuration",
                $route: "/system-settings/ssl-renewal/configuration/",
            },

            actions: {
                $pattern: "system-settings/ssl-renewal/actions",
                $route: "/system-settings/ssl-renewal/actions/",
            },
        },
    },

    /**
     * System Status
     */
    systemStatus: {
        $pattern: "system-status",
        $route: "/system-status/tasks/",

        tasks: {
            $pattern: "system-status/tasks",
            $route: "/system-status/tasks/",

            details: {
                $pattern: "system-status/tasks/:taskId",
                $route: (taskId: string) => `/system-status/tasks/${taskId}/`,
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
                        details: {
                            $pattern: "projects/:id/apps/:appId/deployments/:deploymentId",
                            $route: (id: string, appId: string, deploymentId: string) =>
                                `/projects/${id}/apps/${appId}/deployments/${deploymentId}/`,
                        },
                    },

                    scheduledJobTasks: {
                        $pattern: "projects/:id/apps/:appId/sched-jobs/:scheduledJobId/tasks",
                        $route: (id: string, appId: string, scheduledJobId: string) =>
                            `/projects/${id}/apps/${appId}/sched-jobs/${scheduledJobId}/tasks/`,
                        details: {
                            $pattern: "projects/:id/apps/:appId/sched-jobs/:scheduledJobId/tasks/:taskId",
                            $route: (id: string, appId: string, scheduledJobId: string, taskId: string) =>
                                `/projects/${id}/apps/${appId}/sched-jobs/${scheduledJobId}/tasks/${taskId}/`,
                        },
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
                $pattern: "projects/:id/settings",
                $route: (id: string) => `/projects/${id}/settings/`,

                general: {
                    $pattern: "projects/:id/settings/general",
                    $route: (id: string) => `/projects/${id}/settings/general/`,
                },

                buildSettings: {
                    $pattern: "projects/:id/settings/build-settings",
                    $route: (id: string) => `/projects/${id}/settings/build-settings/`,
                },

                storageSettings: {
                    $pattern: "projects/:id/settings/storage-settings",
                    $route: (id: string) => `/projects/${id}/settings/storage-settings/`,
                },

                domainSettings: {
                    $pattern: "projects/:id/settings/domain-settings",
                    $route: (id: string) => `/projects/${id}/settings/domain-settings/`,
                },

                dangerZone: {
                    $pattern: "projects/:id/settings/danger-zone",
                    $route: (id: string) => `/projects/${id}/settings/danger-zone/`,
                },
            },

            providerConfiguration: {
                $pattern: "projects/:id/provider-settings",
                $route: (id: string) => `/projects/${id}/provider-settings/`,

                accessTokens: {
                    $pattern: "projects/:id/provider-settings/access-tokens",
                    $route: (id: string) => `/projects/${id}/provider-settings/access-tokens/`,
                },

                basicAuth: {
                    $pattern: "projects/:id/provider-settings/basic-auth",
                    $route: (id: string) => `/projects/${id}/provider-settings/basic-auth/`,
                },

                cloudStorages: {
                    $pattern: "projects/:id/provider-settings/cloud-storages",
                    $route: (id: string) => `/projects/${id}/provider-settings/cloud-storages/`,
                },

                emailAccounts: {
                    $pattern: "projects/:id/provider-settings/email-accounts",
                    $route: (id: string) => `/projects/${id}/provider-settings/email-accounts/`,
                },

                envVariables: {
                    $pattern: "projects/:id/provider-settings/env-variables",
                    $route: (id: string) => `/projects/${id}/provider-settings/env-variables/`,
                },

                githubApps: {
                    $pattern: "projects/:id/provider-settings/github-apps",
                    $route: (id: string) => `/projects/${id}/provider-settings/github-apps/`,
                },

                webhooks: {
                    $pattern: "projects/:id/provider-settings/webhooks",
                    $route: (id: string) => `/projects/${id}/provider-settings/webhooks/`,
                },

                imPlatforms: {
                    $pattern: "projects/:id/provider-settings/im-platforms",
                    $route: (id: string) => `/projects/${id}/provider-settings/im-platforms/`,
                },

                notificationTargets: {
                    $pattern: "projects/:id/provider-settings/notification-targets",
                    $route: (id: string) => `/projects/${id}/provider-settings/notification-targets/`,
                },

                registryAuth: {
                    $pattern: "projects/:id/provider-settings/registry-auth",
                    $route: (id: string) => `/projects/${id}/provider-settings/registry-auth/`,
                },

                secrets: {
                    $pattern: "projects/:id/provider-settings/secrets",
                    $route: (id: string) => `/projects/${id}/provider-settings/secrets/`,
                },

                sshKeys: {
                    $pattern: "projects/:id/provider-settings/ssh-keys",
                    $route: (id: string) => `/projects/${id}/provider-settings/ssh-keys/`,
                },

                sslProviders: {
                    $pattern: "projects/:id/provider-settings/ssl-providers",
                    $route: (id: string) => `/projects/${id}/provider-settings/ssl-providers/`,
                },

                sslCertificates: {
                    $pattern: "projects/:id/provider-settings/ssl-certificates",
                    $route: (id: string) => `/projects/${id}/provider-settings/ssl-certificates/`,
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
