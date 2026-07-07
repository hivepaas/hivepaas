import { useState } from "react";

import { toast } from "sonner";
import { ProjectSslCertCommands } from "~/projects/data/commands";
import { ProjectDomainSettingsQueries, ProjectSslCertQueries } from "~/projects/data/queries";
import type { SslCert_Notification_Payload } from "~/settings/api/services/ssl-cert-services";
import { DomainSettingsQueries, SslCertCommands, SslCertQueries } from "~/settings/data";
import { SettingsFormRouteHeader } from "~/settings/module-shared/components/settings-form-route-header";
import { CreateOrEditSslCertForm } from "~/settings/module-shared/components/ssl-cert-form";
import type {
    CreateOrEditSslCertFormInput,
    CreateOrEditSslCertFormOutput,
} from "~/settings/module-shared/components/ssl-cert-form";
import { useSettingsScopePermissions } from "~/settings/module-shared/hooks";

import { AppLoader } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";
import { ESslCertType, ESslKeyType } from "@application/shared/enums";
import { useAppNavigate } from "@application/shared/hooks/router";

import type { SslCertTableScope } from "../ssl-cert-table";

type SslCertFormRouteMode = "create" | "edit";

function addDays(date: Date, days: number): Date {
    const result = new Date(date);
    result.setDate(result.getDate() + days);
    return result;
}

function normalizeCertType(certType?: string): ESslCertType {
    switch (certType) {
        case ESslCertType.LetsEncrypt:
        case ESslCertType.ZeroSSL:
        case ESslCertType.GoogleTrust:
        case ESslCertType.SelfSigned:
        case ESslCertType.Custom:
            return certType;
        case "googlets":
            return ESslCertType.GoogleTrust;
        default:
            return ESslCertType.Custom;
    }
}

function getGeneratedCertValidityDays(certType: ESslCertType): number {
    switch (certType) {
        case ESslCertType.LetsEncrypt:
        case ESslCertType.ZeroSSL:
        case ESslCertType.GoogleTrust:
            return 90;
        default:
            return 365;
    }
}

function buildNotificationPayload(
    notification: CreateOrEditSslCertFormOutput["notification"],
): SslCert_Notification_Payload {
    return {
        success: { id: notification.successUseDefault ? "" : (notification.success?.id ?? "") },
        successUseDefault: notification.successUseDefault,
        failure: { id: notification.failureUseDefault ? "" : (notification.failure?.id ?? "") },
        failureUseDefault: notification.failureUseDefault,
    };
}

export function SslCertFormRoute({ mode, scope, sslCertId }: Props) {
    const [hasChanges, setHasChanges] = useState(false);
    const [saveRevision, setSaveRevision] = useState(0);
    const { canWrite } = useSettingsScopePermissions(scope);
    const { navigate } = useAppNavigate();

    const listRoute = getSslCertListRoute(scope);
    const isEditMode = mode === "edit";
    const detailId = isEditMode ? (sslCertId ?? "") : "";
    const projectId = scope.type === "project" ? scope.projectId : "";

    function navigateToList() {
        navigate.modules(listRoute, { ignorePrevPath: true });
    }

    function markSaved() {
        setHasChanges(false);
        setSaveRevision(revision => revision + 1);
    }

    const settingsDomainSettingsQuery = DomainSettingsQueries.useFindOne(
        {},
        {
            enabled: mode === "create" && scope.type === "settings",
        },
    );

    const projectDomainSettingsQuery = ProjectDomainSettingsQueries.useFindOne(
        {
            projectID: projectId,
        },
        {
            enabled: mode === "create" && scope.type === "project" && !!projectId,
        },
    );

    const settingDetailQuery = SslCertQueries.useFindOneById(
        { id: detailId },
        { enabled: isEditMode && scope.type === "settings" },
    );
    const projectDetailQuery = ProjectSslCertQueries.useFindOneById(
        {
            projectID: projectId,
            id: detailId,
        },
        { enabled: isEditMode && scope.type === "project" },
    );
    const detailQuery = scope.type === "project" ? projectDetailQuery : settingDetailQuery;
    const sslCert = detailQuery.data?.data;
    const readOnlyInherited = scope.type === "project" && sslCert?.inherited === true;

    const { mutate: createSettingSslCert, isPending: isCreatingSetting } = SslCertCommands.useCreateOne({
        onSuccess: () => {
            toast.success("SSL certificate created successfully");
            markSaved();
        },
    });
    const { mutate: updateSettingSslCert, isPending: isUpdatingSetting } = SslCertCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("SSL certificate updated successfully");
            markSaved();
        },
    });
    const { mutate: createProjectSslCert, isPending: isCreatingProject } = ProjectSslCertCommands.useCreateOne({
        onSuccess: () => {
            toast.success("Project SSL certificate created successfully");
            markSaved();
        },
    });
    const { mutate: updateProjectSslCert, isPending: isUpdatingProject } = ProjectSslCertCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Project SSL certificate updated successfully");
            markSaved();
        },
    });

    function createPayload(values: CreateOrEditSslCertFormOutput) {
        const now = new Date();
        const isCustom = values.certType === ESslCertType.Custom;
        const generatedCertValidityDays = getGeneratedCertValidityDays(values.certType);
        const fallbackExpireAt = addDays(now, generatedCertValidityDays);
        const expireAt = values.expireAt ?? fallbackExpireAt;
        const notifyFrom = values.notifyFrom ?? addDays(expireAt, -30);

        return {
            availableInProjects: scope.type === "project" ? false : values.availableInProjects,
            default: values.default,
            certType: values.certType,
            provider: values.provider?.id ? { id: values.provider.id } : undefined,
            acmeProvider: values.acmeProvider?.id ? { id: values.acmeProvider.id } : undefined,
            domain: values.domain,
            certificate: isCustom ? values.certificate : "",
            privateKey: isCustom ? values.privateKey : "",
            keyType: values.keyType,
            validPeriod: `${generatedCertValidityDays}d`,
            email: values.email,
            autoRenew: isCustom ? false : values.autoRenew,
            expireAt,
            notifyFrom,
            notification: buildNotificationPayload(values.notification),
        };
    }

    function onSubmit(values: CreateOrEditSslCertFormOutput) {
        const payload = createPayload(values);

        if (isEditMode && sslCert) {
            const updatePayload = { ...payload, updateVer: sslCert.updateVer };

            if (scope.type === "project") {
                updateProjectSslCert({
                    projectID: scope.projectId,
                    id: sslCert.id,
                    payload: updatePayload,
                });
                return;
            }

            updateSettingSslCert({ id: sslCert.id, payload: updatePayload });
            return;
        }

        if (scope.type === "project") {
            createProjectSslCert({ projectID: scope.projectId, payload });
            return;
        }

        createSettingSslCert({ payload });
    }

    function handleClose() {
        if (isPending) return;
        if (
            !readOnlyInherited &&
            canWrite &&
            hasChanges &&
            !window.confirm("Are you sure you want to close without saving changes?")
        )
            return;

        navigateToList();
    }

    const domainSettings =
        scope.type === "project"
            ? projectDomainSettingsQuery.data?.data.certSettings
            : settingsDomainSettingsQuery.data?.data.certSettings;
    const isPending = isCreatingSetting || isUpdatingSetting || isCreatingProject || isUpdatingProject;
    const isDetailLoading = isEditMode && detailQuery.isFetching;
    const initialValues: Partial<CreateOrEditSslCertFormInput> | undefined =
        isEditMode && sslCert
            ? {
                  domain: sslCert.domain,
                  certType: normalizeCertType(sslCert.certType),
                  provider: sslCert.provider ?? undefined,
                  acmeProvider: sslCert.acmeProvider ?? undefined,
                  email: sslCert.email,
                  keyType: sslCert.keyType,
                  autoRenew: sslCert.autoRenew,
                  certificate: sslCert.certificate,
                  privateKey: sslCert.privateKey,
                  expireAt: sslCert.expireAt ?? null,
                  notifyFrom: sslCert.notifyFrom ?? null,
                  availableInProjects: sslCert.availableInProjects ?? false,
                  default: sslCert.default ?? false,
                  notification: {
                      successUseDefault: sslCert.notification?.successUseDefault ?? true,
                      success: sslCert.notification?.success ?? undefined,
                      failureUseDefault: sslCert.notification?.failureUseDefault ?? true,
                      failure: sslCert.notification?.failure ?? undefined,
                  },
              }
            : {
                  certType: ESslCertType.LetsEncrypt,
                  email: domainSettings?.email ?? "",
                  keyType: domainSettings?.keyType ?? ESslKeyType.ECP256,
                  autoRenew: domainSettings?.autoRenew ?? true,
                  availableInProjects: false,
                  default: false,
                  notification: {
                      successUseDefault: true,
                      failureUseDefault: true,
                  },
              };
    const shouldRenderForm = mode === "create" || !!sslCert;
    const title = mode === "create" ? "Create SSL Certificate" : "Edit SSL Certificate";

    return (
        <div className="flex w-full flex-col overflow-hidden">
            <SettingsFormRouteHeader title={title} />

            {isDetailLoading && (
                <div className="flex min-h-[220px] items-center justify-center">
                    <AppLoader />
                </div>
            )}

            {!isDetailLoading && shouldRenderForm && (
                <CreateOrEditSslCertForm
                    isPending={isPending}
                    onSubmit={onSubmit}
                    onHasChanges={setHasChanges}
                    savedVersion={saveRevision}
                    initialValues={initialValues}
                    scope={scope}
                    showAvailableInProjects={scope.type === "settings"}
                    readOnlyInherited={readOnlyInherited}
                    readOnly={!canWrite}
                    onClose={handleClose}
                />
            )}
        </div>
    );
}

function getSslCertListRoute(scope: SslCertTableScope) {
    if (scope.type === "project") {
        return ROUTE.projects.single.providerConfiguration.sslCertificates.$route(scope.projectId);
    }

    return ROUTE.settings.sslCertificates.$route;
}

interface Props {
    mode: SslCertFormRouteMode;
    scope: SslCertTableScope;
    sslCertId?: string;
}
