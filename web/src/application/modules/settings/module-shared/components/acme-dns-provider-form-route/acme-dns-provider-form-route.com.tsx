import { useState } from "react";

import { toast } from "sonner";
import { ProjectAcmeDnsProviderCommands } from "~/projects/data/commands";
import { ProjectAcmeDnsProviderQueries } from "~/projects/data/queries";
import type {
    AcmeDnsProvider_ConfigPayload,
    AcmeDnsProvider_CreateOne_Payload,
} from "~/settings/api/services/acme-dns-provider-services";
import { AcmeDnsProviderCommands } from "~/settings/data/commands";
import { AcmeDnsProviderQueries } from "~/settings/data/queries";
import type { SettingAcmeDnsProvider } from "~/settings/domain";
import {
    CreateOrEditAcmeDnsProviderForm,
    DEFAULT_ACME_DNS_PROVIDER_FORM_VALUES,
} from "~/settings/module-shared/components/acme-dns-provider-form";
import type {
    CreateOrEditAcmeDnsProviderFormInput,
    CreateOrEditAcmeDnsProviderFormOutput,
} from "~/settings/module-shared/components/acme-dns-provider-form";
import { useSettingsScopePermissions } from "~/settings/module-shared/hooks";
import { SettingsFormRouteHeader } from "~/settings/module-shared/components/settings-form-route-header";

import { AppLoader } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";
import { EAcmeDnsProviderKind } from "@application/shared/enums";
import { useAppNavigate } from "@application/shared/hooks/router";

import type { AcmeDnsProviderTableScope } from "../acme-dns-provider-table";

type AcmeDnsProviderFormRouteMode = "create" | "edit";
type TestStatus = "idle" | "success" | "error";

const ACME_DNS_ALLOW_LIST_SEPARATOR = ",";

export function AcmeDnsProviderFormRoute({ mode, scope, acmeDnsProviderId }: Props) {
    const [hasChanges, setHasChanges] = useState(false);
    const [saveRevision, setSaveRevision] = useState(0);
    const [testStatus, setTestStatus] = useState<TestStatus>("idle");
    const { canWrite } = useSettingsScopePermissions(scope);
    const { navigate } = useAppNavigate();

    const listRoute = getAcmeDnsProviderListRoute(scope);
    const isEditMode = mode === "edit";
    const detailId = isEditMode ? (acmeDnsProviderId ?? "") : "";

    function navigateToList() {
        navigate.modules(listRoute, { ignorePrevPath: true });
    }

    function markSaved() {
        setHasChanges(false);
        setSaveRevision(revision => revision + 1);
    }

    const { mutate: createSettingAcmeDnsProvider, isPending: isCreatingSetting } = AcmeDnsProviderCommands.useCreateOne(
        {
            onSuccess: () => {
                toast.success("ACME DNS provider created successfully");
                markSaved();
            },
        },
    );
    const { mutate: updateSettingAcmeDnsProvider, isPending: isUpdatingSetting } = AcmeDnsProviderCommands.useUpdateOne(
        {
            onSuccess: () => {
                toast.success("ACME DNS provider updated successfully");
                markSaved();
            },
        },
    );
    const { mutate: createProjectAcmeDnsProvider, isPending: isCreatingProject } =
        ProjectAcmeDnsProviderCommands.useCreateOne({
            onSuccess: () => {
                toast.success("Project ACME DNS provider created successfully");
                markSaved();
            },
        });
    const { mutate: updateProjectAcmeDnsProvider, isPending: isUpdatingProject } =
        ProjectAcmeDnsProviderCommands.useUpdateOne({
            onSuccess: () => {
                toast.success("Project ACME DNS provider updated successfully");
                markSaved();
            },
        });
    const { mutate: testAccess, isPending: isTesting } = AcmeDnsProviderCommands.useTestAccess({
        onSuccess: () => {
            setTestStatus("success");
        },
        onError: () => {
            setTestStatus("error");
        },
    });

    const settingDetailQuery = AcmeDnsProviderQueries.useFindOneById(
        { id: detailId },
        { enabled: isEditMode && scope.type === "settings" },
    );
    const projectDetailQuery = ProjectAcmeDnsProviderQueries.useFindOneById(
        {
            projectID: scope.type === "project" ? scope.projectId : "",
            id: detailId,
        },
        { enabled: isEditMode && scope.type === "project" },
    );
    const detailQuery = scope.type === "project" ? projectDetailQuery : settingDetailQuery;
    const acmeDnsProvider = detailQuery.data?.data;
    const readOnlyInherited = scope.type === "project" && acmeDnsProvider?.inherited === true;

    function onSubmit(values: CreateOrEditAcmeDnsProviderFormOutput) {
        const payload = createPayload(values, scope.type === "project");

        if (isEditMode && acmeDnsProvider) {
            const updatePayload = {
                ...payload,
                updateVer: acmeDnsProvider.updateVer,
            };

            if (scope.type === "project") {
                updateProjectAcmeDnsProvider({
                    projectID: scope.projectId,
                    id: acmeDnsProvider.id,
                    payload: updatePayload,
                });
                return;
            }

            updateSettingAcmeDnsProvider({
                id: acmeDnsProvider.id,
                payload: updatePayload,
            });
            return;
        }

        if (scope.type === "project") {
            createProjectAcmeDnsProvider({
                projectID: scope.projectId,
                payload,
            });
            return;
        }

        createSettingAcmeDnsProvider({ payload });
    }

    function onTestAccess(values: CreateOrEditAcmeDnsProviderFormOutput, testDomain: string) {
        setTestStatus("idle");
        testAccess({
            payload: {
                ...createConfigPayload(values),
                testDomain,
            },
        });
    }

    function handleClose() {
        if (isPending || isTesting) return;
        if (
            !readOnlyInherited &&
            canWrite &&
            hasChanges &&
            !window.confirm("Are you sure you want to close without saving changes?")
        )
            return;

        navigateToList();
    }

    const isPending = isCreatingSetting || isUpdatingSetting || isCreatingProject || isUpdatingProject;
    const isDetailLoading = isEditMode && detailQuery.isFetching;
    const initialValues =
        isEditMode && acmeDnsProvider ? createInitialValues(acmeDnsProvider) : createInitialValues(undefined);
    const shouldRenderForm = mode === "create" || !!acmeDnsProvider;
    const title = mode === "create" ? "Create ACME DNS Provider" : "Edit ACME DNS Provider";

    return (
        <div className="flex w-full flex-col">
            <SettingsFormRouteHeader title={title} />

            {isDetailLoading && (
                <div className="flex min-h-[220px] items-center justify-center">
                    <AppLoader />
                </div>
            )}

            {!isDetailLoading && shouldRenderForm && (
                <CreateOrEditAcmeDnsProviderForm
                    isPending={isPending}
                    isTesting={isTesting}
                    testStatus={testStatus}
                    onSubmit={onSubmit}
                    onTestAccess={onTestAccess}
                    onHasChanges={setHasChanges}
                    savedVersion={saveRevision}
                    initialValues={initialValues}
                    showAvailableInProjects={scope.type === "settings"}
                    showTestAccess
                    isEdit={isEditMode}
                    readOnlyInherited={readOnlyInherited}
                    readOnly={!canWrite}
                    onClose={handleClose}
                />
            )}
        </div>
    );
}

function secretValue(value: string | undefined, secretMasked?: boolean) {
    return secretMasked ? "" : (value ?? "");
}

function createConfigPayload(values: CreateOrEditAcmeDnsProviderFormOutput): AcmeDnsProvider_ConfigPayload {
    const isAcmeDns = values.kind === EAcmeDnsProviderKind.AcmeDNS;
    const isAzure = values.kind === EAcmeDnsProviderKind.Azure;
    const isBaiduCloud = values.kind === EAcmeDnsProviderKind.BaiduCloud;
    const isCloudflare = values.kind === EAcmeDnsProviderKind.Cloudflare;
    const isDigitalOcean = values.kind === EAcmeDnsProviderKind.DigitalOcean;
    const isGCloud = values.kind === EAcmeDnsProviderKind.GCloud;
    const isGoDaddy = values.kind === EAcmeDnsProviderKind.GoDaddy;
    const isHetzner = values.kind === EAcmeDnsProviderKind.Hetzner;
    const isHuaweiCloud = values.kind === EAcmeDnsProviderKind.HuaweiCloud;
    const isNamecheap = values.kind === EAcmeDnsProviderKind.Namecheap;
    const isRFC2136 = values.kind === EAcmeDnsProviderKind.RFC2136;
    const isRoute53 = values.kind === EAcmeDnsProviderKind.Route53;
    const isTencentCloud = values.kind === EAcmeDnsProviderKind.TencentCloud;

    return {
        kind: values.kind,
        name: values.name,
        acmeDns: isAcmeDns
            ? {
                  apiBase: values.acmeDnsApiBase,
                  allowList: values.acmeDnsAllowList
                      .split(ACME_DNS_ALLOW_LIST_SEPARATOR)
                      .map(item => item.trim())
                      .filter(Boolean),
                  storagePath: values.acmeDnsStoragePath,
                  storageBaseUrl: values.acmeDnsStorageBaseUrl,
              }
            : null,
        azure: isAzure
            ? {
                  clientId: values.azureClientId,
                  clientSecret: values.azureClientSecret,
                  subscriptionId: values.azureSubscriptionId,
                  tenantId: values.azureTenantId,
                  resourceGroupName: values.azureResourceGroupName,
              }
            : null,
        baiduCloud: isBaiduCloud
            ? {
                  accessKey: values.baiduCloudAccessKey,
                  secretKey: values.baiduCloudSecretKey,
              }
            : null,
        cloudflare: isCloudflare
            ? {
                  authToken: values.cloudflareAuthToken,
              }
            : null,
        digitalOcean: isDigitalOcean
            ? {
                  authToken: values.digitalOceanAuthToken,
              }
            : null,
        gCloud: isGCloud
            ? {
                  serviceAccount: values.gCloudServiceAccount,
                  projectId: values.gCloudProjectId,
              }
            : null,
        goDaddy: isGoDaddy
            ? {
                  apiKey: values.goDaddyApiKey,
                  apiSecret: values.goDaddyApiSecret,
              }
            : null,
        hetzner: isHetzner
            ? {
                  apiToken: values.hetznerApiToken,
              }
            : null,
        huaweiCloud: isHuaweiCloud
            ? {
                  accessKey: values.huaweiCloudAccessKey,
                  secretKey: values.huaweiCloudSecretKey,
                  region: values.huaweiCloudRegion,
              }
            : null,
        namecheap: isNamecheap
            ? {
                  apiUser: values.namecheapApiUser,
                  apiKey: values.namecheapApiKey,
              }
            : null,
        rfc2136: isRFC2136
            ? {
                  nameserver: values.rfc2136Nameserver,
                  tsigKeyName: values.rfc2136TsigKeyName,
                  tsigSecret: values.rfc2136TsigSecret,
                  tsigAlgorithm: values.rfc2136TsigAlgorithm,
              }
            : null,
        route53: isRoute53
            ? {
                  accessKeyId: values.route53AccessKeyId,
                  secretAccessKey: values.route53SecretAccessKey,
                  hostedZoneId: values.route53HostedZoneId,
                  region: values.route53Region,
              }
            : null,
        tencentCloud: isTencentCloud
            ? {
                  secretId: values.tencentCloudSecretId,
                  secretKey: values.tencentCloudSecretKey,
                  region: values.tencentCloudRegion,
              }
            : null,
    };
}

function createPayload(
    values: CreateOrEditAcmeDnsProviderFormOutput,
    isProjectScope: boolean,
): AcmeDnsProvider_CreateOne_Payload {
    return {
        ...createConfigPayload(values),
        availableInProjects: isProjectScope ? false : values.availableInProjects,
        default: values.default,
    };
}

function createInitialValues(acmeDnsProvider?: SettingAcmeDnsProvider): Partial<CreateOrEditAcmeDnsProviderFormInput> {
    if (!acmeDnsProvider) {
        return {
            kind: EAcmeDnsProviderKind.AcmeDNS,
            availableInProjects: false,
            default: false,
        };
    }

    return {
        ...DEFAULT_ACME_DNS_PROVIDER_FORM_VALUES,
        name: acmeDnsProvider.name,
        kind: acmeDnsProvider.kind,
        acmeDnsApiBase: acmeDnsProvider.acmeDns?.apiBase ?? "",
        acmeDnsAllowList: acmeDnsProvider.acmeDns?.allowList.join(ACME_DNS_ALLOW_LIST_SEPARATOR) ?? "",
        acmeDnsStoragePath: acmeDnsProvider.acmeDns?.storagePath ?? "",
        acmeDnsStorageBaseUrl: acmeDnsProvider.acmeDns?.storageBaseUrl ?? "",
        azureClientId: acmeDnsProvider.azure?.clientId ?? "",
        azureClientSecret: secretValue(acmeDnsProvider.azure?.clientSecret, acmeDnsProvider.secretMasked),
        azureSubscriptionId: acmeDnsProvider.azure?.subscriptionId ?? "",
        azureTenantId: acmeDnsProvider.azure?.tenantId ?? "",
        azureResourceGroupName: acmeDnsProvider.azure?.resourceGroupName ?? "",
        baiduCloudAccessKey: acmeDnsProvider.baiduCloud?.accessKey ?? "",
        baiduCloudSecretKey: secretValue(acmeDnsProvider.baiduCloud?.secretKey, acmeDnsProvider.secretMasked),
        cloudflareAuthToken: secretValue(acmeDnsProvider.cloudflare?.authToken, acmeDnsProvider.secretMasked),
        digitalOceanAuthToken: secretValue(acmeDnsProvider.digitalOcean?.authToken, acmeDnsProvider.secretMasked),
        gCloudServiceAccount: secretValue(acmeDnsProvider.gCloud?.serviceAccount, acmeDnsProvider.secretMasked),
        gCloudProjectId: acmeDnsProvider.gCloud?.projectId ?? "",
        goDaddyApiKey: acmeDnsProvider.goDaddy?.apiKey ?? "",
        goDaddyApiSecret: secretValue(acmeDnsProvider.goDaddy?.apiSecret, acmeDnsProvider.secretMasked),
        hetznerApiToken: secretValue(acmeDnsProvider.hetzner?.apiToken, acmeDnsProvider.secretMasked),
        huaweiCloudAccessKey: acmeDnsProvider.huaweiCloud?.accessKey ?? "",
        huaweiCloudSecretKey: secretValue(acmeDnsProvider.huaweiCloud?.secretKey, acmeDnsProvider.secretMasked),
        huaweiCloudRegion: acmeDnsProvider.huaweiCloud?.region ?? "",
        namecheapApiUser: acmeDnsProvider.namecheap?.apiUser ?? "",
        namecheapApiKey: secretValue(acmeDnsProvider.namecheap?.apiKey, acmeDnsProvider.secretMasked),
        rfc2136Nameserver: acmeDnsProvider.rfc2136?.nameserver ?? "",
        rfc2136TsigKeyName: acmeDnsProvider.rfc2136?.tsigKeyName ?? "",
        rfc2136TsigSecret: secretValue(acmeDnsProvider.rfc2136?.tsigSecret, acmeDnsProvider.secretMasked),
        rfc2136TsigAlgorithm: acmeDnsProvider.rfc2136?.tsigAlgorithm ?? "",
        route53AccessKeyId: acmeDnsProvider.route53?.accessKeyId ?? "",
        route53SecretAccessKey: secretValue(acmeDnsProvider.route53?.secretAccessKey, acmeDnsProvider.secretMasked),
        route53HostedZoneId: acmeDnsProvider.route53?.hostedZoneId ?? "",
        route53Region: acmeDnsProvider.route53?.region ?? "",
        tencentCloudSecretId: acmeDnsProvider.tencentCloud?.secretId ?? "",
        tencentCloudSecretKey: secretValue(acmeDnsProvider.tencentCloud?.secretKey, acmeDnsProvider.secretMasked),
        tencentCloudRegion: acmeDnsProvider.tencentCloud?.region ?? "",
        availableInProjects: acmeDnsProvider.availableInProjects ?? false,
        default: acmeDnsProvider.default ?? false,
    };
}

function getAcmeDnsProviderListRoute(scope: AcmeDnsProviderTableScope) {
    if (scope.type === "project") {
        return ROUTE.projects.single.providerConfiguration.acmeDnsProviders.$route(scope.projectId);
    }

    return ROUTE.settings.acmeDnsProviders.$route;
}

interface Props {
    mode: AcmeDnsProviderFormRouteMode;
    scope: AcmeDnsProviderTableScope;
    acmeDnsProviderId?: string;
}
