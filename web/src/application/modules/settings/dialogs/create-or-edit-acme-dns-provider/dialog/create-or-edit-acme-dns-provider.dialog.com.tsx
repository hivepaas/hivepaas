import { useEffect, useState } from "react";

import { Dialog, DialogBody, DialogFixedContent, DialogHeader, DialogTitle } from "@components/ui/dialog";
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
import { useSettingsScopePermissions } from "~/settings/module-shared/hooks";

import { AppLoader } from "@application/shared/components";
import { EAcmeDnsProviderKind } from "@application/shared/enums";

import { CreateOrEditAcmeDnsProviderForm, DEFAULT_ACME_DNS_PROVIDER_FORM_VALUES } from "../form";
import { useCreateOrEditAcmeDnsProviderDialogState } from "../hooks";
import type { CreateOrEditAcmeDnsProviderFormInput, CreateOrEditAcmeDnsProviderFormOutput } from "../schemas";

function secretValue(value: string | undefined, secretMasked?: boolean) {
    return secretMasked ? "" : (value ?? "");
}

const ACME_DNS_ALLOW_LIST_SEPARATOR = ",";

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

export function CreateOrEditAcmeDnsProviderDialog() {
    const {
        state,
        props: dialogOptions,
        close: closeDialog,
        clear: clearDialog,
    } = useCreateOrEditAcmeDnsProviderDialogState();
    const [hasChanges, setHasChanges] = useState(false);
    const [testStatus, setTestStatus] = useState<"idle" | "success" | "error">("idle");

    const permissionScope = state.mode === "closed" ? ({ type: "settings" } as const) : state.scope;
    const { canWrite } = useSettingsScopePermissions(permissionScope);

    const { mutate: createSettingAcmeDnsProvider, isPending: isCreatingSetting } =
        AcmeDnsProviderCommands.useCreateOne({
            onSuccess: () => {
                toast.success("ACME DNS provider created successfully");
                closeDialog();
                dialogOptions?.onSuccess?.();
            },
            onError: dialogOptions?.onError,
        });
    const { mutate: updateSettingAcmeDnsProvider, isPending: isUpdatingSetting } =
        AcmeDnsProviderCommands.useUpdateOne({
            onSuccess: () => {
                toast.success("ACME DNS provider updated successfully");
                closeDialog();
                dialogOptions?.onSuccess?.();
            },
            onError: dialogOptions?.onError,
        });
    const { mutate: createProjectAcmeDnsProvider, isPending: isCreatingProject } =
        ProjectAcmeDnsProviderCommands.useCreateOne({
            onSuccess: () => {
                toast.success("Project ACME DNS provider created successfully");
                closeDialog();
                dialogOptions?.onSuccess?.();
            },
            onError: dialogOptions?.onError,
        });
    const { mutate: updateProjectAcmeDnsProvider, isPending: isUpdatingProject } =
        ProjectAcmeDnsProviderCommands.useUpdateOne({
            onSuccess: () => {
                toast.success("Project ACME DNS provider updated successfully");
                closeDialog();
                dialogOptions?.onSuccess?.();
            },
            onError: dialogOptions?.onError,
        });
    const { mutate: testAccess, isPending: isTesting } = AcmeDnsProviderCommands.useTestAccess({
        onSuccess: () => {
            setTestStatus("success");
        },
        onError: error => {
            setTestStatus("error");
            dialogOptions?.onError?.(error);
        },
    });

    useEffect(() => {
        if (state.mode === "closed") {
            setHasChanges(false);
            setTestStatus("idle");
            clearDialog();
        }
    }, [clearDialog, state.mode]);

    const detailId = state.mode === "edit" ? state.id : "";
    const settingDetailQuery = AcmeDnsProviderQueries.useFindOneById(
        { id: detailId },
        {
            enabled: state.mode === "edit" && state.scope.type === "settings",
        },
    );
    const projectDetailQuery = ProjectAcmeDnsProviderQueries.useFindOneById(
        {
            projectID: state.mode === "edit" && state.scope.type === "project" ? state.scope.projectId : "",
            id: detailId,
        },
        {
            enabled: state.mode === "edit" && state.scope.type === "project",
        },
    );
    const detailQuery =
        state.mode === "edit" && state.scope.type === "project" ? projectDetailQuery : settingDetailQuery;
    const acmeDnsProvider = detailQuery.data?.data;

    function onSubmit(values: CreateOrEditAcmeDnsProviderFormOutput) {
        if (state.mode === "closed") {
            return;
        }

        const payload = createPayload(values, state.scope.type === "project");

        if (state.mode === "edit" && acmeDnsProvider) {
            const updatePayload = {
                ...payload,
                updateVer: acmeDnsProvider.updateVer,
            };

            if (state.scope.type === "project") {
                updateProjectAcmeDnsProvider({
                    projectID: state.scope.projectId,
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

        if (state.scope.type === "project") {
            createProjectAcmeDnsProvider({
                projectID: state.scope.projectId,
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
        if (isPending) {
            return;
        }

        if (
            !readOnlyInherited &&
            canWrite &&
            hasChanges &&
            !window.confirm("Are you sure you want to close without saving changes?")
        ) {
            return;
        }

        closeDialog();
        dialogOptions?.onClose?.();
    }

    const open = state.mode !== "closed";
    const resolvedDialogOptions = dialogOptions ?? {};
    const readOnlyInherited = resolvedDialogOptions.readOnlyInherited === true;
    const dialogTitle = readOnlyInherited
        ? (resolvedDialogOptions.entityTitle ?? "ACME DNS Provider")
        : "Create or update a DNS provider";
    const isPending = isCreatingSetting || isUpdatingSetting || isCreatingProject || isUpdatingProject;
    const showAvailableInProjects = state.mode !== "closed" && state.scope.type === "settings";
    const showTestAccess = state.mode !== "closed";
    const initialValues =
        state.mode === "edit" && acmeDnsProvider
            ? createInitialValues(acmeDnsProvider)
            : createInitialValues(undefined);
    const isDetailLoading = state.mode === "edit" && detailQuery.isFetching;
    const canRenderForm = state.mode === "open" || (state.mode === "edit" && !!acmeDnsProvider);

    return (
        <Dialog
            open={open}
            onOpenChange={handleClose}
        >
            <DialogFixedContent className="min-w-[390px] w-[860px]">
                <DialogHeader>
                    <DialogTitle>{dialogTitle}</DialogTitle>
                </DialogHeader>
                {isDetailLoading && (
                    <DialogBody>
                        <AppLoader />
                    </DialogBody>
                )}
                {state.mode !== "closed" && !isDetailLoading && canRenderForm && (
                    <CreateOrEditAcmeDnsProviderForm
                        isPending={isPending}
                        isTesting={isTesting}
                        testStatus={testStatus}
                        onSubmit={onSubmit}
                        onTestAccess={onTestAccess}
                        onHasChanges={setHasChanges}
                        initialValues={initialValues}
                        showAvailableInProjects={showAvailableInProjects}
                        showTestAccess={showTestAccess}
                        isEdit={state.mode === "edit"}
                        readOnlyInherited={readOnlyInherited}
                        readOnly={!canWrite}
                        onClose={handleClose}
                    />
                )}
            </DialogFixedContent>
        </Dialog>
    );
}
