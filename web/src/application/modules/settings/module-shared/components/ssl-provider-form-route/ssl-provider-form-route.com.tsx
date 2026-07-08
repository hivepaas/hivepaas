import { useState } from "react";

import { toast } from "sonner";
import { ProjectSslProviderCommands } from "~/projects/data/commands";
import { ProjectSslProviderQueries } from "~/projects/data/queries";
import type { SslProvider_CreateOne_Payload } from "~/settings/api/services/ssl-provider-services";
import { SslProviderCommands } from "~/settings/data/commands";
import { SslProviderQueries } from "~/settings/data/queries";
import type { SettingSslProvider } from "~/settings/domain";
import {
    CreateOrEditSslProviderForm,
    SSL_PROVIDER_UNSPECIFIED_KEY_TYPE,
} from "~/settings/module-shared/components/ssl-provider-form";
import type {
    CreateOrEditSslProviderFormInput,
    CreateOrEditSslProviderFormOutput,
} from "~/settings/module-shared/components/ssl-provider-form";
import { useSettingsScopePermissions } from "~/settings/module-shared/hooks";
import { SettingsFormRouteHeader } from "~/settings/module-shared/components/settings-form-route-header";

import { AppLoader } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";
import { ESslProviderKind } from "@application/shared/enums";
import { useAppNavigate } from "@application/shared/hooks/router";

import type { SslProviderTableScope } from "../ssl-provider-table";

type SslProviderFormRouteMode = "create" | "edit";

function getEabValues(sslProvider?: SettingSslProvider) {
    if (!sslProvider) {
        return {
            eabKid: "",
            eabHmacKey: "",
        };
    }

    if (sslProvider.kind === ESslProviderKind.ZeroSSL) {
        return {
            eabKid: sslProvider.zeroSSL?.eabKid ?? "",
            eabHmacKey: sslProvider.zeroSSL?.eabHmacKey ?? "",
        };
    }

    if (sslProvider.kind === ESslProviderKind.GoogleTrust) {
        return {
            eabKid: sslProvider.googleTrust?.eabKid ?? "",
            eabHmacKey: sslProvider.googleTrust?.eabHmacKey ?? "",
        };
    }

    return {
        eabKid: "",
        eabHmacKey: "",
    };
}

function createPayload(
    values: CreateOrEditSslProviderFormOutput,
    isProjectScope: boolean,
): SslProvider_CreateOne_Payload {
    const defaultKeyType =
        values.defaultKeyType === SSL_PROVIDER_UNSPECIFIED_KEY_TYPE ? undefined : values.defaultKeyType;
    const isLetsEncrypt = values.kind === ESslProviderKind.LetsEncrypt;
    const isZeroSSL = values.kind === ESslProviderKind.ZeroSSL;
    const isGoogleTrust = values.kind === ESslProviderKind.GoogleTrust;
    const eabPayload = {
        eabKid: values.eabKid,
        eabHmacKey: values.eabHmacKey,
    };

    return {
        availableInProjects: isProjectScope ? false : values.availableInProjects,
        default: values.default,
        name: values.name,
        kind: values.kind,
        email: values.email,
        ...(defaultKeyType ? { defaultKeyType } : {}),
        letsEncrypt: isLetsEncrypt ? {} : null,
        zeroSSL: isZeroSSL ? eabPayload : null,
        googleTrust: isGoogleTrust ? eabPayload : null,
    };
}

export function SslProviderFormRoute({ mode, scope, sslProviderId }: Props) {
    const [hasChanges, setHasChanges] = useState(false);
    const [saveRevision, setSaveRevision] = useState(0);
    const { canWrite } = useSettingsScopePermissions(scope);
    const { navigate } = useAppNavigate();

    const listRoute = getSslProviderListRoute(scope);
    const isEditMode = mode === "edit";
    const detailId = isEditMode ? (sslProviderId ?? "") : "";

    function navigateToList() {
        navigate.modules(listRoute, { ignorePrevPath: true });
    }

    function markSaved() {
        setHasChanges(false);
        setSaveRevision(revision => revision + 1);
    }

    const { mutate: createSettingSslProvider, isPending: isCreatingSetting } = SslProviderCommands.useCreateOne({
        onSuccess: () => {
            toast.success("SSL provider created successfully");
            markSaved();
        },
    });
    const { mutate: updateSettingSslProvider, isPending: isUpdatingSetting } = SslProviderCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("SSL provider updated successfully");
            markSaved();
        },
    });
    const { mutate: createProjectSslProvider, isPending: isCreatingProject } = ProjectSslProviderCommands.useCreateOne({
        onSuccess: () => {
            toast.success("Project SSL provider created successfully");
            markSaved();
        },
    });
    const { mutate: updateProjectSslProvider, isPending: isUpdatingProject } = ProjectSslProviderCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Project SSL provider updated successfully");
            markSaved();
        },
    });

    const settingDetailQuery = SslProviderQueries.useFindOneById(
        { id: detailId },
        { enabled: isEditMode && scope.type === "settings" },
    );
    const projectDetailQuery = ProjectSslProviderQueries.useFindOneById(
        {
            projectID: scope.type === "project" ? scope.projectId : "",
            id: detailId,
        },
        { enabled: isEditMode && scope.type === "project" },
    );
    const detailQuery = scope.type === "project" ? projectDetailQuery : settingDetailQuery;
    const sslProvider = detailQuery.data?.data;
    const readOnlyInherited = scope.type === "project" && sslProvider?.inherited === true;

    function onSubmit(values: CreateOrEditSslProviderFormOutput) {
        const payload = createPayload(values, scope.type === "project");

        if (isEditMode && sslProvider) {
            const updatePayload = { ...payload, updateVer: sslProvider.updateVer };

            if (scope.type === "project") {
                updateProjectSslProvider({
                    projectID: scope.projectId,
                    id: sslProvider.id,
                    payload: updatePayload,
                });
                return;
            }

            updateSettingSslProvider({ id: sslProvider.id, payload: updatePayload });
            return;
        }

        if (scope.type === "project") {
            createProjectSslProvider({ projectID: scope.projectId, payload });
            return;
        }

        createSettingSslProvider({ payload });
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

    const isPending = isCreatingSetting || isUpdatingSetting || isCreatingProject || isUpdatingProject;
    const isDetailLoading = isEditMode && detailQuery.isFetching;
    const initialValues: Partial<CreateOrEditSslProviderFormInput> | undefined =
        isEditMode && sslProvider
            ? {
                  name: sslProvider.name,
                  kind: sslProvider.kind,
                  email: sslProvider.email,
                  defaultKeyType: sslProvider.defaultKeyType || SSL_PROVIDER_UNSPECIFIED_KEY_TYPE,
                  ...getEabValues(sslProvider),
                  availableInProjects: sslProvider.availableInProjects ?? false,
                  default: sslProvider.default ?? false,
              }
            : {
                  kind: ESslProviderKind.LetsEncrypt,
                  defaultKeyType: SSL_PROVIDER_UNSPECIFIED_KEY_TYPE,
                  availableInProjects: false,
                  default: false,
              };
    const shouldRenderForm = mode === "create" || !!sslProvider;
    const title = mode === "create" ? "Create SSL Provider" : "Edit SSL Provider";

    return (
        <div className="flex w-full flex-col">
            <SettingsFormRouteHeader title={title} />

            {isDetailLoading && (
                <div className="flex min-h-[220px] items-center justify-center">
                    <AppLoader />
                </div>
            )}

            {!isDetailLoading && shouldRenderForm && (
                <CreateOrEditSslProviderForm
                    isPending={isPending}
                    onSubmit={onSubmit}
                    onHasChanges={setHasChanges}
                    savedVersion={saveRevision}
                    initialValues={initialValues}
                    showAvailableInProjects={scope.type === "settings"}
                    isEdit={isEditMode}
                    readOnlyInherited={readOnlyInherited}
                    readOnly={!canWrite}
                    onClose={handleClose}
                />
            )}
        </div>
    );
}

function getSslProviderListRoute(scope: SslProviderTableScope) {
    if (scope.type === "project") {
        return ROUTE.projects.single.providerConfiguration.sslProviders.$route(scope.projectId);
    }

    return ROUTE.settings.sslProviders.$route;
}

interface Props {
    mode: SslProviderFormRouteMode;
    scope: SslProviderTableScope;
    sslProviderId?: string;
}
