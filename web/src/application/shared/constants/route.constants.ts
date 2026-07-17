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

            create: {
                $pattern: "current-user/api-keys/create",
                $route: "/current-user/api-keys/create/",
            },
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

            create: {
                $pattern: "cluster/networks/create",
                $route: "/cluster/networks/create/",
            },

            details: {
                $pattern: "cluster/networks/:networkId",
                $route: (networkId: string) => `/cluster/networks/${networkId}/`,
            },
        },

        volumes: {
            $pattern: "cluster/volumes",
            $route: "/cluster/volumes/",

            create: {
                $pattern: "cluster/volumes/create",
                $route: "/cluster/volumes/create/",
            },

            edit: {
                $pattern: "cluster/volumes/:volumeId/edit",
                $route: (volumeId: string) => `/cluster/volumes/${volumeId}/edit/`,
            },
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

            create: {
                $pattern: "sources/github-apps/create",
                $route: "/sources/github-apps/create/",
            },

            edit: {
                $pattern: "sources/github-apps/:githubAppId/edit",
                $route: (githubAppId: string) => `/sources/github-apps/${githubAppId}/edit/`,
            },
        },

        webhooks: {
            $pattern: "sources/webhooks",
            $route: "/sources/webhooks/",

            create: {
                $pattern: "sources/webhooks/create",
                $route: "/sources/webhooks/create/",
            },

            edit: {
                $pattern: "sources/webhooks/:repoWebhookId/edit",
                $route: (repoWebhookId: string) => `/sources/webhooks/${repoWebhookId}/edit/`,
            },
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

            create: {
                $pattern: "settings/basic-auth/create",
                $route: "/settings/basic-auth/create/",
            },

            edit: {
                $pattern: "settings/basic-auth/:basicAuthId/edit",
                $route: (basicAuthId: string) => `/settings/basic-auth/${basicAuthId}/edit/`,
            },
        },

        registryAuth: {
            $pattern: "settings/registry-auth",
            $route: "/settings/registry-auth/",

            create: {
                $pattern: "settings/registry-auth/create",
                $route: "/settings/registry-auth/create/",
            },

            edit: {
                $pattern: "settings/registry-auth/:registryAuthId/edit",
                $route: (registryAuthId: string) => `/settings/registry-auth/${registryAuthId}/edit/`,
            },
        },

        sslProviders: {
            $pattern: "settings/ssl-providers",
            $route: "/settings/ssl-providers/",

            create: {
                $pattern: "settings/ssl-providers/create",
                $route: "/settings/ssl-providers/create/",
            },

            edit: {
                $pattern: "settings/ssl-providers/:sslProviderId/edit",
                $route: (sslProviderId: string) => `/settings/ssl-providers/${sslProviderId}/edit/`,
            },
        },

        sslCertificates: {
            $pattern: "settings/ssl-certificates",
            $route: "/settings/ssl-certificates/",

            create: {
                $pattern: "settings/ssl-certificates/create",
                $route: "/settings/ssl-certificates/create/",
            },

            edit: {
                $pattern: "settings/ssl-certificates/:sslCertId/edit",
                $route: (sslCertId: string) => `/settings/ssl-certificates/${sslCertId}/edit/`,
            },
        },

        emailAccounts: {
            $pattern: "settings/email-accounts",
            $route: "/settings/email-accounts/",

            create: {
                $pattern: "settings/email-accounts/create",
                $route: "/settings/email-accounts/create/",
            },

            edit: {
                $pattern: "settings/email-accounts/:emailAccountId/edit",
                $route: (emailAccountId: string) => `/settings/email-accounts/${emailAccountId}/edit/`,
            },
        },

        imPlatforms: {
            $pattern: "settings/im-platforms",
            $route: "/settings/im-platforms/",

            create: {
                $pattern: "settings/im-platforms/create",
                $route: "/settings/im-platforms/create/",
            },

            edit: {
                $pattern: "settings/im-platforms/:imPlatformId/edit",
                $route: (imPlatformId: string) => `/settings/im-platforms/${imPlatformId}/edit/`,
            },
        },

        sshKeys: {
            $pattern: "settings/ssh-keys",
            $route: "/settings/ssh-keys/",

            create: {
                $pattern: "settings/ssh-keys/create",
                $route: "/settings/ssh-keys/create/",
            },

            edit: {
                $pattern: "settings/ssh-keys/:sshKeyId/edit",
                $route: (sshKeyId: string) => `/settings/ssh-keys/${sshKeyId}/edit/`,
            },
        },

        accessTokens: {
            $pattern: "settings/access-tokens",
            $route: "/settings/access-tokens/",

            create: {
                $pattern: "settings/access-tokens/create",
                $route: "/settings/access-tokens/create/",
            },

            edit: {
                $pattern: "settings/access-tokens/:accessTokenId/edit",
                $route: (accessTokenId: string) => `/settings/access-tokens/${accessTokenId}/edit/`,
            },
        },

        acmeDnsProviders: {
            $pattern: "settings/acme-dns-providers",
            $route: "/settings/acme-dns-providers/",

            create: {
                $pattern: "settings/acme-dns-providers/create",
                $route: "/settings/acme-dns-providers/create/",
            },

            edit: {
                $pattern: "settings/acme-dns-providers/:acmeDnsProviderId/edit",
                $route: (acmeDnsProviderId: string) => `/settings/acme-dns-providers/${acmeDnsProviderId}/edit/`,
            },
        },

        cloudStorages: {
            $pattern: "settings/cloud-storages",
            $route: "/settings/cloud-storages/",

            create: {
                $pattern: "settings/cloud-storages/create",
                $route: "/settings/cloud-storages/create/",
            },

            edit: {
                $pattern: "settings/cloud-storages/:cloudStorageId/edit",
                $route: (cloudStorageId: string) => `/settings/cloud-storages/${cloudStorageId}/edit/`,
            },
        },

        oauth: {
            $pattern: "settings/oauth",
            $route: "/settings/oauth/",

            create: {
                $pattern: "settings/oauth/create",
                $route: "/settings/oauth/create/",
            },

            edit: {
                $pattern: "settings/oauth/:oauthId/edit",
                $route: (oauthId: string) => `/settings/oauth/${oauthId}/edit/`,
            },
        },

        notificationTargets: {
            $pattern: "settings/notification-targets",
            $route: "/settings/notification-targets/",

            create: {
                $pattern: "settings/notification-targets/create",
                $route: "/settings/notification-targets/create/",
            },

            edit: {
                $pattern: "settings/notification-targets/:notificationTargetId/edit",
                $route: (notificationTargetId: string) =>
                    `/settings/notification-targets/${notificationTargetId}/edit/`,
            },
        },
    },

    /**
     * System Settings
     */
    systemSettings: {
        $pattern: "system",

        hivepaas: {
            $pattern: "system/hivepaas",
            $route: "/system/hivepaas/general/",

            general: {
                $pattern: "system/hivepaas/general",
                $route: "/system/hivepaas/general/",
            },

            httpSettings: {
                $pattern: "system/hivepaas/http-settings",
                $route: "/system/hivepaas/http-settings/",
            },
        },

        traefik: {
            $pattern: "system/traefik",
            $route: "/system/traefik/general/",

            general: {
                $pattern: "system/traefik/general",
                $route: "/system/traefik/general/",
            },
        },

        dataBackup: {
            $pattern: "system/data-backup",
            $route: "/system/data-backup/configuration/",

            configuration: {
                $pattern: "system/data-backup/configuration",
                $route: "/system/data-backup/configuration/",
            },

            backupFiles: {
                $pattern: "system/data-backup/backup-files",
                $route: "/system/data-backup/backup-files/",
            },

            actions: {
                $pattern: "system/data-backup/actions",
                $route: "/system/data-backup/actions/",
            },
        },

        dataCleanup: {
            $pattern: "system/data-cleanup",
            $route: "/system/data-cleanup/configuration/",

            configuration: {
                $pattern: "system/data-cleanup/configuration",
                $route: "/system/data-cleanup/configuration/",
            },

            actions: {
                $pattern: "system/data-cleanup/actions",
                $route: "/system/data-cleanup/actions/",
            },
        },

        sslRenewal: {
            $pattern: "system/ssl-renewal",
            $route: "/system/ssl-renewal/configuration/",

            configuration: {
                $pattern: "system/ssl-renewal/configuration",
                $route: "/system/ssl-renewal/configuration/",
            },

            actions: {
                $pattern: "system/ssl-renewal/actions",
                $route: "/system/ssl-renewal/actions/",
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

                            create: {
                                $pattern: "projects/:id/apps/:appId/health-checks/create",
                                $route: (id: string, appId: string) =>
                                    `/projects/${id}/apps/${appId}/health-checks/create/`,
                            },

                            edit: {
                                $pattern: "projects/:id/apps/:appId/health-checks/:healthCheckId/edit",
                                $route: (id: string, appId: string, healthCheckId: string) =>
                                    `/projects/${id}/apps/${appId}/health-checks/${healthCheckId}/edit/`,
                            },
                        },

                        scheduledJobs: {
                            $pattern: "projects/:id/apps/:appId/sched-jobs",
                            $route: (id: string, appId: string) => `/projects/${id}/apps/${appId}/sched-jobs/`,

                            create: {
                                $pattern: "projects/:id/apps/:appId/sched-jobs/create",
                                $route: (id: string, appId: string) =>
                                    `/projects/${id}/apps/${appId}/sched-jobs/create/`,
                            },

                            edit: {
                                $pattern: "projects/:id/apps/:appId/sched-jobs/:scheduledJobId/edit",
                                $route: (id: string, appId: string, scheduledJobId: string) =>
                                    `/projects/${id}/apps/${appId}/sched-jobs/${scheduledJobId}/edit/`,
                            },
                        },

                        envVariables: {
                            $pattern: "projects/:id/apps/:appId/env-variables",
                            $route: (id: string, appId: string) => `/projects/${id}/apps/${appId}/env-variables/`,
                        },

                        secrets: {
                            $pattern: "projects/:id/apps/:appId/secrets",
                            $route: (id: string, appId: string) => `/projects/${id}/apps/${appId}/secrets/`,

                            create: {
                                $pattern: "projects/:id/apps/:appId/secrets/create",
                                $route: (id: string, appId: string) => `/projects/${id}/apps/${appId}/secrets/create/`,
                            },

                            edit: {
                                $pattern: "projects/:id/apps/:appId/secrets/:secretId/edit",
                                $route: (id: string, appId: string, secretId: string) =>
                                    `/projects/${id}/apps/${appId}/secrets/${secretId}/edit/`,
                            },
                        },

                        configFiles: {
                            $pattern: "projects/:id/apps/:appId/config-files",
                            $route: (id: string, appId: string) => `/projects/${id}/apps/${appId}/config-files/`,

                            create: {
                                $pattern: "projects/:id/apps/:appId/config-files/create",
                                $route: (id: string, appId: string) =>
                                    `/projects/${id}/apps/${appId}/config-files/create/`,
                            },

                            edit: {
                                $pattern: "projects/:id/apps/:appId/config-files/:configFileId/edit",
                                $route: (id: string, appId: string, configFileId: string) =>
                                    `/projects/${id}/apps/${appId}/config-files/${configFileId}/edit/`,
                            },
                        },

                        dataFiles: {
                            $pattern: "projects/:id/apps/:appId/data-files",
                            $route: (id: string, appId: string) => `/projects/${id}/apps/${appId}/data-files/`,

                            create: {
                                $pattern: "projects/:id/apps/:appId/data-files/create",
                                $route: (id: string, appId: string) =>
                                    `/projects/${id}/apps/${appId}/data-files/create/`,
                            },
                        },

                        availabilityAndScaling: {
                            $pattern: "projects/:id/apps/:appId/availability-and-scaling",
                            $route: (id: string, appId: string) =>
                                `/projects/${id}/apps/${appId}/availability-and-scaling/`,
                        },

                        presistentStorage: {
                            $pattern: "projects/:id/apps/:appId/presistent-storage",
                            $route: (id: string, appId: string) => `/projects/${id}/apps/${appId}/presistent-storage/`,

                            create: {
                                $pattern: "projects/:id/apps/:appId/presistent-storage/create",
                                $route: (id: string, appId: string) =>
                                    `/projects/${id}/apps/${appId}/presistent-storage/create/`,
                            },

                            edit: {
                                $pattern: "projects/:id/apps/:appId/presistent-storage/:mountId/edit",
                                $route: (id: string, appId: string, mountId: string) =>
                                    `/projects/${id}/apps/${appId}/presistent-storage/${mountId}/edit/`,
                            },
                        },

                        networks: {
                            $pattern: "projects/:id/apps/:appId/networks",
                            $route: (id: string, appId: string) => `/projects/${id}/apps/${appId}/networks/`,
                        },

                        resources: {
                            $pattern: "projects/:id/apps/:appId/resources",
                            $route: (id: string, appId: string) => `/projects/${id}/apps/${appId}/resources/`,
                        },

                        featureSettings: {
                            $pattern: "projects/:id/apps/:appId/feature-settings",
                            $route: (id: string, appId: string) => `/projects/${id}/apps/${appId}/feature-settings/`,
                        },

                        dangerZone: {
                            $pattern: "projects/:id/apps/:appId/danger-zone",
                            $route: (id: string, appId: string) => `/projects/${id}/apps/${appId}/danger-zone/`,
                        },
                    },

                    instances: {
                        $pattern: "projects/:id/apps/:appId/instances",
                        $route: (id: string, appId: string) => `/projects/${id}/apps/${appId}/instances/`,
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

                    terminal: {
                        $pattern: "projects/:id/apps/:appId/terminal",
                        $route: (id: string, appId: string) => `/projects/${id}/apps/${appId}/terminal/`,
                    },

                    previewDeployments: {
                        $pattern: "projects/:id/apps/:appId/preview-deployments",
                        $route: (id: string, appId: string) => `/projects/${id}/apps/${appId}/preview-deployments/`,

                        create: {
                            $pattern: "projects/:id/apps/:appId/preview-deployments/create",
                            $route: (id: string, appId: string) =>
                                `/projects/${id}/apps/${appId}/preview-deployments/create/`,
                        },
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

                    create: {
                        $pattern: "projects/:id/provider-settings/access-tokens/create",
                        $route: (id: string) => `/projects/${id}/provider-settings/access-tokens/create/`,
                    },

                    edit: {
                        $pattern: "projects/:id/provider-settings/access-tokens/:accessTokenId/edit",
                        $route: (id: string, accessTokenId: string) =>
                            `/projects/${id}/provider-settings/access-tokens/${accessTokenId}/edit/`,
                    },
                },

                acmeDnsProviders: {
                    $pattern: "projects/:id/provider-settings/acme-dns-providers",
                    $route: (id: string) => `/projects/${id}/provider-settings/acme-dns-providers/`,

                    create: {
                        $pattern: "projects/:id/provider-settings/acme-dns-providers/create",
                        $route: (id: string) => `/projects/${id}/provider-settings/acme-dns-providers/create/`,
                    },

                    edit: {
                        $pattern: "projects/:id/provider-settings/acme-dns-providers/:acmeDnsProviderId/edit",
                        $route: (id: string, acmeDnsProviderId: string) =>
                            `/projects/${id}/provider-settings/acme-dns-providers/${acmeDnsProviderId}/edit/`,
                    },
                },

                basicAuth: {
                    $pattern: "projects/:id/provider-settings/basic-auth",
                    $route: (id: string) => `/projects/${id}/provider-settings/basic-auth/`,

                    create: {
                        $pattern: "projects/:id/provider-settings/basic-auth/create",
                        $route: (id: string) => `/projects/${id}/provider-settings/basic-auth/create/`,
                    },

                    edit: {
                        $pattern: "projects/:id/provider-settings/basic-auth/:basicAuthId/edit",
                        $route: (id: string, basicAuthId: string) =>
                            `/projects/${id}/provider-settings/basic-auth/${basicAuthId}/edit/`,
                    },
                },

                cloudStorages: {
                    $pattern: "projects/:id/provider-settings/cloud-storages",
                    $route: (id: string) => `/projects/${id}/provider-settings/cloud-storages/`,

                    create: {
                        $pattern: "projects/:id/provider-settings/cloud-storages/create",
                        $route: (id: string) => `/projects/${id}/provider-settings/cloud-storages/create/`,
                    },

                    edit: {
                        $pattern: "projects/:id/provider-settings/cloud-storages/:cloudStorageId/edit",
                        $route: (id: string, cloudStorageId: string) =>
                            `/projects/${id}/provider-settings/cloud-storages/${cloudStorageId}/edit/`,
                    },
                },

                commandTemplates: {
                    $pattern: "projects/:id/provider-settings/command-templates",
                    $route: (id: string) => `/projects/${id}/provider-settings/command-templates/`,

                    create: {
                        $pattern: "projects/:id/provider-settings/command-templates/create",
                        $route: (id: string) => `/projects/${id}/provider-settings/command-templates/create/`,
                    },

                    edit: {
                        $pattern: "projects/:id/provider-settings/command-templates/:commandTemplateId/edit",
                        $route: (id: string, commandTemplateId: string) =>
                            `/projects/${id}/provider-settings/command-templates/${commandTemplateId}/edit/`,
                    },
                },

                emailAccounts: {
                    $pattern: "projects/:id/provider-settings/email-accounts",
                    $route: (id: string) => `/projects/${id}/provider-settings/email-accounts/`,

                    create: {
                        $pattern: "projects/:id/provider-settings/email-accounts/create",
                        $route: (id: string) => `/projects/${id}/provider-settings/email-accounts/create/`,
                    },

                    edit: {
                        $pattern: "projects/:id/provider-settings/email-accounts/:emailAccountId/edit",
                        $route: (id: string, emailAccountId: string) =>
                            `/projects/${id}/provider-settings/email-accounts/${emailAccountId}/edit/`,
                    },
                },

                envVariables: {
                    $pattern: "projects/:id/provider-settings/env-variables",
                    $route: (id: string) => `/projects/${id}/provider-settings/env-variables/`,
                },

                imPlatforms: {
                    $pattern: "projects/:id/provider-settings/im-platforms",
                    $route: (id: string) => `/projects/${id}/provider-settings/im-platforms/`,

                    create: {
                        $pattern: "projects/:id/provider-settings/im-platforms/create",
                        $route: (id: string) => `/projects/${id}/provider-settings/im-platforms/create/`,
                    },

                    edit: {
                        $pattern: "projects/:id/provider-settings/im-platforms/:imPlatformId/edit",
                        $route: (id: string, imPlatformId: string) =>
                            `/projects/${id}/provider-settings/im-platforms/${imPlatformId}/edit/`,
                    },
                },

                notificationTargets: {
                    $pattern: "projects/:id/provider-settings/notification-targets",
                    $route: (id: string) => `/projects/${id}/provider-settings/notification-targets/`,

                    create: {
                        $pattern: "projects/:id/provider-settings/notification-targets/create",
                        $route: (id: string) => `/projects/${id}/provider-settings/notification-targets/create/`,
                    },

                    edit: {
                        $pattern: "projects/:id/provider-settings/notification-targets/:notificationTargetId/edit",
                        $route: (id: string, notificationTargetId: string) =>
                            `/projects/${id}/provider-settings/notification-targets/${notificationTargetId}/edit/`,
                    },
                },

                registryAuth: {
                    $pattern: "projects/:id/provider-settings/registry-auth",
                    $route: (id: string) => `/projects/${id}/provider-settings/registry-auth/`,

                    create: {
                        $pattern: "projects/:id/provider-settings/registry-auth/create",
                        $route: (id: string) => `/projects/${id}/provider-settings/registry-auth/create/`,
                    },

                    edit: {
                        $pattern: "projects/:id/provider-settings/registry-auth/:registryAuthId/edit",
                        $route: (id: string, registryAuthId: string) =>
                            `/projects/${id}/provider-settings/registry-auth/${registryAuthId}/edit/`,
                    },
                },

                secrets: {
                    $pattern: "projects/:id/provider-settings/secrets",
                    $route: (id: string) => `/projects/${id}/provider-settings/secrets/`,

                    create: {
                        $pattern: "projects/:id/provider-settings/secrets/create",
                        $route: (id: string) => `/projects/${id}/provider-settings/secrets/create/`,
                    },

                    edit: {
                        $pattern: "projects/:id/provider-settings/secrets/:secretId/edit",
                        $route: (id: string, secretId: string) =>
                            `/projects/${id}/provider-settings/secrets/${secretId}/edit/`,
                    },
                },

                sshKeys: {
                    $pattern: "projects/:id/provider-settings/ssh-keys",
                    $route: (id: string) => `/projects/${id}/provider-settings/ssh-keys/`,

                    create: {
                        $pattern: "projects/:id/provider-settings/ssh-keys/create",
                        $route: (id: string) => `/projects/${id}/provider-settings/ssh-keys/create/`,
                    },

                    edit: {
                        $pattern: "projects/:id/provider-settings/ssh-keys/:sshKeyId/edit",
                        $route: (id: string, sshKeyId: string) =>
                            `/projects/${id}/provider-settings/ssh-keys/${sshKeyId}/edit/`,
                    },
                },

                sslProviders: {
                    $pattern: "projects/:id/provider-settings/ssl-providers",
                    $route: (id: string) => `/projects/${id}/provider-settings/ssl-providers/`,

                    create: {
                        $pattern: "projects/:id/provider-settings/ssl-providers/create",
                        $route: (id: string) => `/projects/${id}/provider-settings/ssl-providers/create/`,
                    },

                    edit: {
                        $pattern: "projects/:id/provider-settings/ssl-providers/:sslProviderId/edit",
                        $route: (id: string, sslProviderId: string) =>
                            `/projects/${id}/provider-settings/ssl-providers/${sslProviderId}/edit/`,
                    },
                },

                sslCertificates: {
                    $pattern: "projects/:id/provider-settings/ssl-certificates",
                    $route: (id: string) => `/projects/${id}/provider-settings/ssl-certificates/`,

                    create: {
                        $pattern: "projects/:id/provider-settings/ssl-certificates/create",
                        $route: (id: string) => `/projects/${id}/provider-settings/ssl-certificates/create/`,
                    },

                    edit: {
                        $pattern: "projects/:id/provider-settings/ssl-certificates/:sslCertId/edit",
                        $route: (id: string, sslCertId: string) =>
                            `/projects/${id}/provider-settings/ssl-certificates/${sslCertId}/edit/`,
                    },
                },
            },

            sources: {
                $pattern: "projects/:id/sources",
                $route: (id: string) => `/projects/${id}/sources/`,

                githubApps: {
                    $pattern: "projects/:id/sources/github-apps",
                    $route: (id: string) => `/projects/${id}/sources/github-apps/`,

                    create: {
                        $pattern: "projects/:id/sources/github-apps/create",
                        $route: (id: string) => `/projects/${id}/sources/github-apps/create/`,
                    },

                    edit: {
                        $pattern: "projects/:id/sources/github-apps/:githubAppId/edit",
                        $route: (id: string, githubAppId: string) =>
                            `/projects/${id}/sources/github-apps/${githubAppId}/edit/`,
                    },
                },

                webhooks: {
                    $pattern: "projects/:id/sources/webhooks",
                    $route: (id: string) => `/projects/${id}/sources/webhooks/`,

                    create: {
                        $pattern: "projects/:id/sources/webhooks/create",
                        $route: (id: string) => `/projects/${id}/sources/webhooks/create/`,
                    },

                    edit: {
                        $pattern: "projects/:id/sources/webhooks/:repoWebhookId/edit",
                        $route: (id: string, repoWebhookId: string) =>
                            `/projects/${id}/sources/webhooks/${repoWebhookId}/edit/`,
                    },
                },
            },

            clusterResources: {
                $pattern: "projects/:id/cluster-resources",
                $route: (id: string) => `/projects/${id}/cluster-resources/`,

                networks: {
                    $pattern: "projects/:id/cluster-resources/networks",
                    $route: (id: string) => `/projects/${id}/cluster-resources/networks/`,

                    create: {
                        $pattern: "projects/:id/cluster-resources/networks/create",
                        $route: (id: string) => `/projects/${id}/cluster-resources/networks/create/`,
                    },

                    details: {
                        $pattern: "projects/:id/cluster-resources/networks/:networkId",
                        $route: (id: string, networkId: string) =>
                            `/projects/${id}/cluster-resources/networks/${networkId}/`,
                    },
                },

                volumes: {
                    $pattern: "projects/:id/cluster-resources/volumes",
                    $route: (id: string) => `/projects/${id}/cluster-resources/volumes/`,

                    create: {
                        $pattern: "projects/:id/cluster-resources/volumes/create",
                        $route: (id: string) => `/projects/${id}/cluster-resources/volumes/create/`,
                    },

                    edit: {
                        $pattern: "projects/:id/cluster-resources/volumes/:volumeId/edit",
                        $route: (id: string, volumeId: string) =>
                            `/projects/${id}/cluster-resources/volumes/${volumeId}/edit/`,
                    },
                },
            },
        },
    },
} as const;
