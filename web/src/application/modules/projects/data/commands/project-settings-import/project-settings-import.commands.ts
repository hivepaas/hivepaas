import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useProjectSettingsImportApi } from "~/projects/api/hooks";
import type { ProjectSettingsImport_Import_Req, ProjectSettingsImport_Import_Res } from "~/projects/api/services";
import { QK } from "~/projects/data/constants";

export const PROJECT_SETTINGS_IMPORT_KIND = {
    BasicAuth: "basicAuth",
    RegistryAuth: "registryAuth",
    SslCert: "sslCert",
    SslProvider: "sslProvider",
    Email: "email",
    ImService: "imService",
    SSHKey: "sshKey",
    AccessToken: "accessToken",
    AcmeDnsProvider: "acmeDnsProvider",
    CloudStorage: "cloudStorage",
    Notification: "notification",
    GithubApp: "githubApp",
    RepoWebhook: "repoWebhook",
    ClusterNetwork: "clusterNetwork",
    ClusterVolume: "clusterVolume",
} as const;

export type ProjectSettingsImportKind =
    (typeof PROJECT_SETTINGS_IMPORT_KIND)[keyof typeof PROJECT_SETTINGS_IMPORT_KIND];

const PROJECT_SETTINGS_IMPORT_LIST_QUERY_KEYS = {
    [PROJECT_SETTINGS_IMPORT_KIND.BasicAuth]: QK["projects.basic-auth.$.find-many-paginated"],
    [PROJECT_SETTINGS_IMPORT_KIND.RegistryAuth]: QK["projects.registry-auth.$.find-many-paginated"],
    [PROJECT_SETTINGS_IMPORT_KIND.SslCert]: QK["projects.ssl-cert.$.find-many-paginated"],
    [PROJECT_SETTINGS_IMPORT_KIND.SslProvider]: QK["projects.ssl-provider.$.find-many-paginated"],
    [PROJECT_SETTINGS_IMPORT_KIND.Email]: QK["projects.email.$.find-many-paginated"],
    [PROJECT_SETTINGS_IMPORT_KIND.ImService]: QK["projects.im-service.$.find-many-paginated"],
    [PROJECT_SETTINGS_IMPORT_KIND.SSHKey]: QK["projects.ssh-key.$.find-many-paginated"],
    [PROJECT_SETTINGS_IMPORT_KIND.AccessToken]: QK["projects.access-token.$.find-many-paginated"],
    [PROJECT_SETTINGS_IMPORT_KIND.AcmeDnsProvider]: QK["projects.acme-dns-provider.$.find-many-paginated"],
    [PROJECT_SETTINGS_IMPORT_KIND.CloudStorage]: QK["projects.cloud-storage.$.find-many-paginated"],
    [PROJECT_SETTINGS_IMPORT_KIND.Notification]: QK["projects.notifications.$.find-many-paginated"],
    [PROJECT_SETTINGS_IMPORT_KIND.GithubApp]: QK["projects.github-app.$.find-many-paginated"],
    [PROJECT_SETTINGS_IMPORT_KIND.RepoWebhook]: QK["projects.repo-webhook.$.find-many-paginated"],
    [PROJECT_SETTINGS_IMPORT_KIND.ClusterNetwork]: QK["projects.networks.$.find-many-paginated"],
    [PROJECT_SETTINGS_IMPORT_KIND.ClusterVolume]: QK["projects.cluster-volumes.$.find-many-paginated"],
} as const satisfies Record<ProjectSettingsImportKind, (typeof QK)[keyof typeof QK]>;

type ImportSettingsReq = ProjectSettingsImport_Import_Req["data"] & {
    settingKind: ProjectSettingsImportKind;
};
type ImportSettingsRes = ProjectSettingsImport_Import_Res;
type ImportSettingsOptions = Omit<UseMutationOptions<ImportSettingsRes, Error, ImportSettingsReq>, "mutationFn">;

function useImportSettings({ onSuccess, ...options }: ImportSettingsOptions = {}) {
    const { mutations } = useProjectSettingsImportApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: ({ settingKind: _settingKind, ...data }) => mutations.importSettings(data),
        onSuccess: (response, request, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [PROJECT_SETTINGS_IMPORT_LIST_QUERY_KEYS[request.settingKind]],
            });

            onSuccess?.(response, request, ...rest);
        },
        ...options,
    });
}

export const ProjectSettingsImportCommands = Object.freeze({
    useImportSettings,
});
