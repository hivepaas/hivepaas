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
        $pattern: "integrations",

        githubApps: {
            $pattern: "integrations/github-apps",
            $route: "/integrations/github-apps/",

            create: {
                $pattern: "integrations/github-apps/create",
                $route: "/integrations/github-apps/create/",
            },

            edit: {
                $pattern: "integrations/github-apps/:githubAppId/edit",
                $route: (githubAppId: string) => `/integrations/github-apps/${githubAppId}/edit/`,
            },
        },

        webhooks: {
            $pattern: "integrations/webhooks",
            $route: "/integrations/webhooks/",

            create: {
                $pattern: "integrations/webhooks/create",
                $route: "/integrations/webhooks/create/",
            },

            edit: {
                $pattern: "integrations/webhooks/:repoWebhookId/edit",
                $route: (repoWebhookId: string) => `/integrations/webhooks/${repoWebhookId}/edit/`,
            },
        },

        basicAuth: {
            $pattern: "integrations/basic-auth",
            $route: "/integrations/basic-auth/",

            create: {
                $pattern: "integrations/basic-auth/create",
                $route: "/integrations/basic-auth/create/",
            },

            edit: {
                $pattern: "integrations/basic-auth/:basicAuthId/edit",
                $route: (basicAuthId: string) => `/integrations/basic-auth/${basicAuthId}/edit/`,
            },
        },

        registryAuth: {
            $pattern: "integrations/registry-auth",
            $route: "/integrations/registry-auth/",

            create: {
                $pattern: "integrations/registry-auth/create",
                $route: "/integrations/registry-auth/create/",
            },

            edit: {
                $pattern: "integrations/registry-auth/:registryAuthId/edit",
                $route: (registryAuthId: string) => `/integrations/registry-auth/${registryAuthId}/edit/`,
            },
        },

        sslProviders: {
            $pattern: "integrations/ssl-providers",
            $route: "/integrations/ssl-providers/",

            create: {
                $pattern: "integrations/ssl-providers/create",
                $route: "/integrations/ssl-providers/create/",
            },

            edit: {
                $pattern: "integrations/ssl-providers/:sslProviderId/edit",
                $route: (sslProviderId: string) => `/integrations/ssl-providers/${sslProviderId}/edit/`,
            },
        },

        sslCertificates: {
            $pattern: "integrations/ssl-certificates",
            $route: "/integrations/ssl-certificates/",

            create: {
                $pattern: "integrations/ssl-certificates/create",
                $route: "/integrations/ssl-certificates/create/",
            },

            edit: {
                $pattern: "integrations/ssl-certificates/:sslCertId/edit",
                $route: (sslCertId: string) => `/integrations/ssl-certificates/${sslCertId}/edit/`,
            },
        },

        emailAccounts: {
            $pattern: "integrations/email-accounts",
            $route: "/integrations/email-accounts/",

            create: {
                $pattern: "integrations/email-accounts/create",
                $route: "/integrations/email-accounts/create/",
            },

            edit: {
                $pattern: "integrations/email-accounts/:emailAccountId/edit",
                $route: (emailAccountId: string) => `/integrations/email-accounts/${emailAccountId}/edit/`,
            },
        },

        imPlatforms: {
            $pattern: "integrations/im-platforms",
            $route: "/integrations/im-platforms/",

            create: {
                $pattern: "integrations/im-platforms/create",
                $route: "/integrations/im-platforms/create/",
            },

            edit: {
                $pattern: "integrations/im-platforms/:imPlatformId/edit",
                $route: (imPlatformId: string) => `/integrations/im-platforms/${imPlatformId}/edit/`,
            },
        },

        sshKeys: {
            $pattern: "integrations/ssh-keys",
            $route: "/integrations/ssh-keys/",

            create: {
                $pattern: "integrations/ssh-keys/create",
                $route: "/integrations/ssh-keys/create/",
            },

            edit: {
                $pattern: "integrations/ssh-keys/:sshKeyId/edit",
                $route: (sshKeyId: string) => `/integrations/ssh-keys/${sshKeyId}/edit/`,
            },
        },

        accessTokens: {
            $pattern: "integrations/access-tokens",
            $route: "/integrations/access-tokens/",

            create: {
                $pattern: "integrations/access-tokens/create",
                $route: "/integrations/access-tokens/create/",
            },

            edit: {
                $pattern: "integrations/access-tokens/:accessTokenId/edit",
                $route: (accessTokenId: string) => `/integrations/access-tokens/${accessTokenId}/edit/`,
            },
        },

        acmeDnsProviders: {
            $pattern: "integrations/acme-dns-providers",
            $route: "/integrations/acme-dns-providers/",

            create: {
                $pattern: "integrations/acme-dns-providers/create",
                $route: "/integrations/acme-dns-providers/create/",
            },

            edit: {
                $pattern: "integrations/acme-dns-providers/:acmeDnsProviderId/edit",
                $route: (acmeDnsProviderId: string) => `/integrations/acme-dns-providers/${acmeDnsProviderId}/edit/`,
            },
        },

        cloudStorages: {
            $pattern: "integrations/cloud-storages",
            $route: "/integrations/cloud-storages/",

            create: {
                $pattern: "integrations/cloud-storages/create",
                $route: "/integrations/cloud-storages/create/",
            },

            edit: {
                $pattern: "integrations/cloud-storages/:cloudStorageId/edit",
                $route: (cloudStorageId: string) => `/integrations/cloud-storages/${cloudStorageId}/edit/`,
            },
        },

        oauth: {
            $pattern: "integrations/oauth",
            $route: "/integrations/oauth/",

            create: {
                $pattern: "integrations/oauth/create",
                $route: "/integrations/oauth/create/",
            },

            edit: {
                $pattern: "integrations/oauth/:oauthId/edit",
                $route: (oauthId: string) => `/integrations/oauth/${oauthId}/edit/`,
            },
        },

        notificationTargets: {
            $pattern: "integrations/notification-targets",
            $route: "/integrations/notification-targets/",

            create: {
                $pattern: "integrations/notification-targets/create",
                $route: "/integrations/notification-targets/create/",
            },

            edit: {
                $pattern: "integrations/notification-targets/:notificationTargetId/edit",
                $route: (notificationTargetId: string) =>
                    `/integrations/notification-targets/${notificationTargetId}/edit/`,
            },
        },

        backupRepos: {
            $pattern: "integrations/backup-repos",
            $route: "/integrations/backup-repos/",

            create: {
                $pattern: "integrations/backup-repos/create",
                $route: "/integrations/backup-repos/create/",
            },

            edit: {
                $pattern: "integrations/backup-repos/:backupRepoId/edit",
                $route: (backupRepoId: string) => `/integrations/backup-repos/${backupRepoId}/edit/`,
            },
        },
    },

    /**
     * App Settings
     */
    appSettings: {
        $pattern: "settings",
        $route: "/settings/image-build/",

        imageBuild: {
            $pattern: "settings/image-build",
            $route: "/settings/image-build/",
        },

        appPlacement: {
            $pattern: "settings/app-placement",
            $route: "/settings/app-placement/",
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

            routingSettings: {
                $pattern: "system/hivepaas/routing-settings",
                $route: "/system/hivepaas/routing-settings/",
            },
        },

        traefik: {
            $pattern: "system/traefik",
            $route: "/system/traefik/general/",

            general: {
                $pattern: "system/traefik/general",
                $route: "/system/traefik/general/",
            },

            configOptions: {
                $pattern: "system/traefik/config-options",
                $route: "/system/traefik/config-options/",
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

        backupRepoCleanup: {
            $pattern: "system/backup-repo-cleanup",
            $route: "/system/backup-repo-cleanup/configuration/",

            configuration: {
                $pattern: "system/backup-repo-cleanup/configuration",
                $route: "/system/backup-repo-cleanup/configuration/",
            },

            actions: {
                $pattern: "system/backup-repo-cleanup/actions",
                $route: "/system/backup-repo-cleanup/actions/",
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
                    $pattern: "projects/:id/:env/apps/:appId",

                    configuration: {
                        $pattern: "projects/:id/:env/apps/:appId/configuration",

                        general: {
                            $pattern: "projects/:id/:env/apps/:appId/general",
                            $route: (id: string, env: string, appId: string) =>
                                `/projects/${id}/${env}/apps/${appId}/general/`,
                        },

                        deploymentSettings: {
                            $pattern: "projects/:id/:env/apps/:appId/deployment-settings",
                            $route: (id: string, env: string, appId: string) =>
                                `/projects/${id}/${env}/apps/${appId}/deployment-settings/`,
                        },

                        containerSettings: {
                            $pattern: "projects/:id/:env/apps/:appId/container-settings",
                            $route: (id: string, env: string, appId: string) =>
                                `/projects/${id}/${env}/apps/${appId}/container-settings/`,
                        },

                        routingSettings: {
                            $pattern: "projects/:id/:env/apps/:appId/routing-settings",
                            $route: (id: string, env: string, appId: string) =>
                                `/projects/${id}/${env}/apps/${appId}/routing-settings/`,
                        },

                        periodicJobs: {
                            $pattern: "projects/:id/:env/apps/:appId/periodic-jobs",
                            $route: (id: string, env: string, appId: string) =>
                                `/projects/${id}/${env}/apps/${appId}/periodic-jobs/`,

                            create: {
                                $pattern: "projects/:id/:env/apps/:appId/periodic-jobs/create",
                                $route: (id: string, env: string, appId: string) =>
                                    `/projects/${id}/${env}/apps/${appId}/periodic-jobs/create/`,
                            },

                            edit: {
                                $pattern: "projects/:id/:env/apps/:appId/periodic-jobs/:healthCheckId/edit",
                                $route: (id: string, env: string, appId: string, healthCheckId: string) =>
                                    `/projects/${id}/${env}/apps/${appId}/periodic-jobs/${healthCheckId}/edit/`,
                            },
                        },

                        scheduledJobs: {
                            $pattern: "projects/:id/:env/apps/:appId/sched-jobs",
                            $route: (id: string, env: string, appId: string) =>
                                `/projects/${id}/${env}/apps/${appId}/sched-jobs/`,

                            create: {
                                $pattern: "projects/:id/:env/apps/:appId/sched-jobs/create",
                                $route: (id: string, env: string, appId: string) =>
                                    `/projects/${id}/${env}/apps/${appId}/sched-jobs/create/`,
                            },

                            edit: {
                                $pattern: "projects/:id/:env/apps/:appId/sched-jobs/:scheduledJobId/edit",
                                $route: (id: string, env: string, appId: string, scheduledJobId: string) =>
                                    `/projects/${id}/${env}/apps/${appId}/sched-jobs/${scheduledJobId}/edit/`,
                            },
                        },

                        envVariables: {
                            $pattern: "projects/:id/:env/apps/:appId/env-variables",
                            $route: (id: string, env: string, appId: string) =>
                                `/projects/${id}/${env}/apps/${appId}/env-variables/`,
                        },

                        secrets: {
                            $pattern: "projects/:id/:env/apps/:appId/secrets",
                            $route: (id: string, env: string, appId: string) =>
                                `/projects/${id}/${env}/apps/${appId}/secrets/`,

                            create: {
                                $pattern: "projects/:id/:env/apps/:appId/secrets/create",
                                $route: (id: string, env: string, appId: string) =>
                                    `/projects/${id}/${env}/apps/${appId}/secrets/create/`,
                            },

                            edit: {
                                $pattern: "projects/:id/:env/apps/:appId/secrets/:secretId/edit",
                                $route: (id: string, env: string, appId: string, secretId: string) =>
                                    `/projects/${id}/${env}/apps/${appId}/secrets/${secretId}/edit/`,
                            },
                        },

                        configFiles: {
                            $pattern: "projects/:id/:env/apps/:appId/config-files",
                            $route: (id: string, env: string, appId: string) =>
                                `/projects/${id}/${env}/apps/${appId}/config-files/`,

                            create: {
                                $pattern: "projects/:id/:env/apps/:appId/config-files/create",
                                $route: (id: string, env: string, appId: string) =>
                                    `/projects/${id}/${env}/apps/${appId}/config-files/create/`,
                            },

                            edit: {
                                $pattern: "projects/:id/:env/apps/:appId/config-files/:configFileId/edit",
                                $route: (id: string, env: string, appId: string, configFileId: string) =>
                                    `/projects/${id}/${env}/apps/${appId}/config-files/${configFileId}/edit/`,
                            },
                        },

                        dataFiles: {
                            $pattern: "projects/:id/:env/apps/:appId/data-files",
                            $route: (id: string, env: string, appId: string) =>
                                `/projects/${id}/${env}/apps/${appId}/data-files/`,

                            create: {
                                $pattern: "projects/:id/:env/apps/:appId/data-files/create",
                                $route: (id: string, env: string, appId: string) =>
                                    `/projects/${id}/${env}/apps/${appId}/data-files/create/`,
                            },
                        },

                        availabilityAndScaling: {
                            $pattern: "projects/:id/:env/apps/:appId/availability-and-scaling",
                            $route: (id: string, env: string, appId: string) =>
                                `/projects/${id}/${env}/apps/${appId}/availability-and-scaling/`,
                        },

                        presistentStorage: {
                            $pattern: "projects/:id/:env/apps/:appId/presistent-storage",
                            $route: (id: string, env: string, appId: string) =>
                                `/projects/${id}/${env}/apps/${appId}/presistent-storage/`,

                            create: {
                                $pattern: "projects/:id/:env/apps/:appId/presistent-storage/create",
                                $route: (id: string, env: string, appId: string) =>
                                    `/projects/${id}/${env}/apps/${appId}/presistent-storage/create/`,
                            },

                            edit: {
                                $pattern: "projects/:id/:env/apps/:appId/presistent-storage/:mountId/edit",
                                $route: (id: string, env: string, appId: string, mountId: string) =>
                                    `/projects/${id}/${env}/apps/${appId}/presistent-storage/${mountId}/edit/`,
                            },
                        },

                        networks: {
                            $pattern: "projects/:id/:env/apps/:appId/networks",
                            $route: (id: string, env: string, appId: string) =>
                                `/projects/${id}/${env}/apps/${appId}/networks/`,
                        },

                        resources: {
                            $pattern: "projects/:id/:env/apps/:appId/resources",
                            $route: (id: string, env: string, appId: string) =>
                                `/projects/${id}/${env}/apps/${appId}/resources/`,
                        },

                        featureSettings: {
                            $pattern: "projects/:id/:env/apps/:appId/feature-settings",
                            $route: (id: string, env: string, appId: string) =>
                                `/projects/${id}/${env}/apps/${appId}/feature-settings/`,
                        },

                        appClone: {
                            $pattern: "projects/:id/:env/apps/:appId/app-clone",
                            $route: (id: string, env: string, appId: string) =>
                                `/projects/${id}/${env}/apps/${appId}/app-clone/`,
                        },

                        dangerZone: {
                            $pattern: "projects/:id/:env/apps/:appId/danger-zone",
                            $route: (id: string, env: string, appId: string) =>
                                `/projects/${id}/${env}/apps/${appId}/danger-zone/`,
                        },
                    },

                    instances: {
                        $pattern: "projects/:id/:env/apps/:appId/instances",
                        $route: (id: string, env: string, appId: string) =>
                            `/projects/${id}/${env}/apps/${appId}/instances/`,
                    },

                    deployments: {
                        $pattern: "projects/:id/:env/apps/:appId/deployments",
                        $route: (id: string, env: string, appId: string) =>
                            `/projects/${id}/${env}/apps/${appId}/deployments/`,
                        details: {
                            $pattern: "projects/:id/:env/apps/:appId/deployments/:deploymentId",
                            $route: (id: string, env: string, appId: string, deploymentId: string) =>
                                `/projects/${id}/${env}/apps/${appId}/deployments/${deploymentId}/`,
                        },
                    },

                    scheduledJobTasks: {
                        $pattern: "projects/:id/:env/apps/:appId/sched-jobs/:scheduledJobId/tasks",
                        $route: (id: string, env: string, appId: string, scheduledJobId: string) =>
                            `/projects/${id}/${env}/apps/${appId}/sched-jobs/${scheduledJobId}/tasks/`,
                        details: {
                            $pattern: "projects/:id/:env/apps/:appId/sched-jobs/:scheduledJobId/tasks/:taskId",
                            $route: (id: string, env: string, appId: string, scheduledJobId: string, taskId: string) =>
                                `/projects/${id}/${env}/apps/${appId}/sched-jobs/${scheduledJobId}/tasks/${taskId}/`,
                        },
                    },

                    logs: {
                        $pattern: "projects/:id/:env/apps/:appId/logs",
                        $route: (id: string, env: string, appId: string) =>
                            `/projects/${id}/${env}/apps/${appId}/logs/`,
                    },

                    terminal: {
                        $pattern: "projects/:id/:env/apps/:appId/terminal",
                        $route: (id: string, env: string, appId: string) =>
                            `/projects/${id}/${env}/apps/${appId}/terminal/`,
                    },

                    previewDeployments: {
                        $pattern: "projects/:id/:env/apps/:appId/preview-deployments",
                        $route: (id: string, env: string, appId: string) =>
                            `/projects/${id}/${env}/apps/${appId}/preview-deployments/`,

                        create: {
                            $pattern: "projects/:id/:env/apps/:appId/preview-deployments/create",
                            $route: (id: string, env: string, appId: string) =>
                                `/projects/${id}/${env}/apps/${appId}/preview-deployments/create/`,
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
                $pattern: "projects/:id/integrations",
                $route: (id: string) => `/projects/${id}/integrations/`,

                accessTokens: {
                    $pattern: "projects/:id/integrations/access-tokens",
                    $route: (id: string) => `/projects/${id}/integrations/access-tokens/`,

                    create: {
                        $pattern: "projects/:id/integrations/access-tokens/create",
                        $route: (id: string) => `/projects/${id}/integrations/access-tokens/create/`,
                    },

                    edit: {
                        $pattern: "projects/:id/integrations/access-tokens/:accessTokenId/edit",
                        $route: (id: string, accessTokenId: string) =>
                            `/projects/${id}/integrations/access-tokens/${accessTokenId}/edit/`,
                    },
                },

                acmeDnsProviders: {
                    $pattern: "projects/:id/integrations/acme-dns-providers",
                    $route: (id: string) => `/projects/${id}/integrations/acme-dns-providers/`,

                    create: {
                        $pattern: "projects/:id/integrations/acme-dns-providers/create",
                        $route: (id: string) => `/projects/${id}/integrations/acme-dns-providers/create/`,
                    },

                    edit: {
                        $pattern: "projects/:id/integrations/acme-dns-providers/:acmeDnsProviderId/edit",
                        $route: (id: string, acmeDnsProviderId: string) =>
                            `/projects/${id}/integrations/acme-dns-providers/${acmeDnsProviderId}/edit/`,
                    },
                },

                basicAuth: {
                    $pattern: "projects/:id/integrations/basic-auth",
                    $route: (id: string) => `/projects/${id}/integrations/basic-auth/`,

                    create: {
                        $pattern: "projects/:id/integrations/basic-auth/create",
                        $route: (id: string) => `/projects/${id}/integrations/basic-auth/create/`,
                    },

                    edit: {
                        $pattern: "projects/:id/integrations/basic-auth/:basicAuthId/edit",
                        $route: (id: string, basicAuthId: string) =>
                            `/projects/${id}/integrations/basic-auth/${basicAuthId}/edit/`,
                    },
                },

                cloudStorages: {
                    $pattern: "projects/:id/integrations/cloud-storages",
                    $route: (id: string) => `/projects/${id}/integrations/cloud-storages/`,

                    create: {
                        $pattern: "projects/:id/integrations/cloud-storages/create",
                        $route: (id: string) => `/projects/${id}/integrations/cloud-storages/create/`,
                    },

                    edit: {
                        $pattern: "projects/:id/integrations/cloud-storages/:cloudStorageId/edit",
                        $route: (id: string, cloudStorageId: string) =>
                            `/projects/${id}/integrations/cloud-storages/${cloudStorageId}/edit/`,
                    },
                },

                commandPipes: {
                    $pattern: "projects/:id/integrations/command-pipes",
                    $route: (id: string) => `/projects/${id}/integrations/command-pipes/`,

                    create: {
                        $pattern: "projects/:id/integrations/command-pipes/create",
                        $route: (id: string) => `/projects/${id}/integrations/command-pipes/create/`,
                    },

                    edit: {
                        $pattern: "projects/:id/integrations/command-pipes/:commandPipeId/edit",
                        $route: (id: string, commandPipeId: string) =>
                            `/projects/${id}/integrations/command-pipes/${commandPipeId}/edit/`,
                    },
                },

                commandTemplates: {
                    $pattern: "projects/:id/integrations/command-templates",
                    $route: (id: string) => `/projects/${id}/integrations/command-templates/`,

                    create: {
                        $pattern: "projects/:id/integrations/command-templates/create",
                        $route: (id: string) => `/projects/${id}/integrations/command-templates/create/`,
                    },

                    edit: {
                        $pattern: "projects/:id/integrations/command-templates/:commandTemplateId/edit",
                        $route: (id: string, commandTemplateId: string) =>
                            `/projects/${id}/integrations/command-templates/${commandTemplateId}/edit/`,
                    },
                },

                backupRepos: {
                    $pattern: "projects/:id/integrations/backup-repos",
                    $route: (id: string) => `/projects/${id}/integrations/backup-repos/`,

                    create: {
                        $pattern: "projects/:id/integrations/backup-repos/create",
                        $route: (id: string) => `/projects/${id}/integrations/backup-repos/create/`,
                    },

                    edit: {
                        $pattern: "projects/:id/integrations/backup-repos/:backupRepoId/edit",
                        $route: (id: string, backupRepoId: string) =>
                            `/projects/${id}/integrations/backup-repos/${backupRepoId}/edit/`,
                    },
                },

                emailAccounts: {
                    $pattern: "projects/:id/integrations/email-accounts",
                    $route: (id: string) => `/projects/${id}/integrations/email-accounts/`,

                    create: {
                        $pattern: "projects/:id/integrations/email-accounts/create",
                        $route: (id: string) => `/projects/${id}/integrations/email-accounts/create/`,
                    },

                    edit: {
                        $pattern: "projects/:id/integrations/email-accounts/:emailAccountId/edit",
                        $route: (id: string, emailAccountId: string) =>
                            `/projects/${id}/integrations/email-accounts/${emailAccountId}/edit/`,
                    },
                },

                envVariables: {
                    $pattern: "projects/:id/integrations/env-variables",
                    $route: (id: string) => `/projects/${id}/integrations/env-variables/`,
                },

                githubApps: {
                    $pattern: "projects/:id/integrations/github-apps",
                    $route: (id: string) => `/projects/${id}/integrations/github-apps/`,

                    create: {
                        $pattern: "projects/:id/integrations/github-apps/create",
                        $route: (id: string) => `/projects/${id}/integrations/github-apps/create/`,
                    },

                    edit: {
                        $pattern: "projects/:id/integrations/github-apps/:githubAppId/edit",
                        $route: (id: string, githubAppId: string) =>
                            `/projects/${id}/integrations/github-apps/${githubAppId}/edit/`,
                    },
                },

                imPlatforms: {
                    $pattern: "projects/:id/integrations/im-platforms",
                    $route: (id: string) => `/projects/${id}/integrations/im-platforms/`,

                    create: {
                        $pattern: "projects/:id/integrations/im-platforms/create",
                        $route: (id: string) => `/projects/${id}/integrations/im-platforms/create/`,
                    },

                    edit: {
                        $pattern: "projects/:id/integrations/im-platforms/:imPlatformId/edit",
                        $route: (id: string, imPlatformId: string) =>
                            `/projects/${id}/integrations/im-platforms/${imPlatformId}/edit/`,
                    },
                },

                notificationTargets: {
                    $pattern: "projects/:id/integrations/notification-targets",
                    $route: (id: string) => `/projects/${id}/integrations/notification-targets/`,

                    create: {
                        $pattern: "projects/:id/integrations/notification-targets/create",
                        $route: (id: string) => `/projects/${id}/integrations/notification-targets/create/`,
                    },

                    edit: {
                        $pattern: "projects/:id/integrations/notification-targets/:notificationTargetId/edit",
                        $route: (id: string, notificationTargetId: string) =>
                            `/projects/${id}/integrations/notification-targets/${notificationTargetId}/edit/`,
                    },
                },

                registryAuth: {
                    $pattern: "projects/:id/integrations/registry-auth",
                    $route: (id: string) => `/projects/${id}/integrations/registry-auth/`,

                    create: {
                        $pattern: "projects/:id/integrations/registry-auth/create",
                        $route: (id: string) => `/projects/${id}/integrations/registry-auth/create/`,
                    },

                    edit: {
                        $pattern: "projects/:id/integrations/registry-auth/:registryAuthId/edit",
                        $route: (id: string, registryAuthId: string) =>
                            `/projects/${id}/integrations/registry-auth/${registryAuthId}/edit/`,
                    },
                },

                secrets: {
                    $pattern: "projects/:id/integrations/secrets",
                    $route: (id: string) => `/projects/${id}/integrations/secrets/`,

                    create: {
                        $pattern: "projects/:id/integrations/secrets/create",
                        $route: (id: string) => `/projects/${id}/integrations/secrets/create/`,
                    },

                    edit: {
                        $pattern: "projects/:id/integrations/secrets/:secretId/edit",
                        $route: (id: string, secretId: string) =>
                            `/projects/${id}/integrations/secrets/${secretId}/edit/`,
                    },
                },

                sshKeys: {
                    $pattern: "projects/:id/integrations/ssh-keys",
                    $route: (id: string) => `/projects/${id}/integrations/ssh-keys/`,

                    create: {
                        $pattern: "projects/:id/integrations/ssh-keys/create",
                        $route: (id: string) => `/projects/${id}/integrations/ssh-keys/create/`,
                    },

                    edit: {
                        $pattern: "projects/:id/integrations/ssh-keys/:sshKeyId/edit",
                        $route: (id: string, sshKeyId: string) =>
                            `/projects/${id}/integrations/ssh-keys/${sshKeyId}/edit/`,
                    },
                },

                sslProviders: {
                    $pattern: "projects/:id/integrations/ssl-providers",
                    $route: (id: string) => `/projects/${id}/integrations/ssl-providers/`,

                    create: {
                        $pattern: "projects/:id/integrations/ssl-providers/create",
                        $route: (id: string) => `/projects/${id}/integrations/ssl-providers/create/`,
                    },

                    edit: {
                        $pattern: "projects/:id/integrations/ssl-providers/:sslProviderId/edit",
                        $route: (id: string, sslProviderId: string) =>
                            `/projects/${id}/integrations/ssl-providers/${sslProviderId}/edit/`,
                    },
                },

                sslCertificates: {
                    $pattern: "projects/:id/integrations/ssl-certificates",
                    $route: (id: string) => `/projects/${id}/integrations/ssl-certificates/`,

                    create: {
                        $pattern: "projects/:id/integrations/ssl-certificates/create",
                        $route: (id: string) => `/projects/${id}/integrations/ssl-certificates/create/`,
                    },

                    edit: {
                        $pattern: "projects/:id/integrations/ssl-certificates/:sslCertId/edit",
                        $route: (id: string, sslCertId: string) =>
                            `/projects/${id}/integrations/ssl-certificates/${sslCertId}/edit/`,
                    },
                },

                webhooks: {
                    $pattern: "projects/:id/integrations/webhooks",
                    $route: (id: string) => `/projects/${id}/integrations/webhooks/`,

                    create: {
                        $pattern: "projects/:id/integrations/webhooks/create",
                        $route: (id: string) => `/projects/${id}/integrations/webhooks/create/`,
                    },

                    edit: {
                        $pattern: "projects/:id/integrations/webhooks/:repoWebhookId/edit",
                        $route: (id: string, repoWebhookId: string) =>
                            `/projects/${id}/integrations/webhooks/${repoWebhookId}/edit/`,
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
