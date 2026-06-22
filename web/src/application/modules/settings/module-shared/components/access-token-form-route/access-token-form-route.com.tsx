import { useState } from "react";

import { ArrowLeft } from "lucide-react";
import { toast } from "sonner";
import { ProjectAccessTokenCommands } from "~/projects/data/commands";
import { ProjectAccessTokenQueries } from "~/projects/data/queries";
import { AccessTokenCommands } from "~/settings/data/commands";
import { AccessTokenQueries } from "~/settings/data/queries";
import { CreateOrEditAccessTokenForm } from "~/settings/module-shared/components/access-token-form";
import type {
    CreateOrEditAccessTokenFormInput,
    CreateOrEditAccessTokenFormOutput,
} from "~/settings/module-shared/components/access-token-form";
import { useSettingsScopePermissions } from "~/settings/module-shared/hooks";

import { AppLoader } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";
import { EAccessTokenKind } from "@application/shared/enums";
import { useAppNavigate } from "@application/shared/hooks/router";

import { Button } from "@/components/ui";

import type { AccessTokenTableScope } from "../access-token-table";

type AccessTokenFormRouteMode = "create" | "edit";

export function AccessTokenFormRoute({ mode, scope, accessTokenId }: Props) {
    const [hasChanges, setHasChanges] = useState(false);
    const [testStatus, setTestStatus] = useState<"idle" | "succeeded" | "failed">("idle");
    const { canWrite } = useSettingsScopePermissions(scope);
    const { navigate } = useAppNavigate();

    const listRoute = getAccessTokenListRoute(scope);
    const isEditMode = mode === "edit";
    const detailId = isEditMode ? (accessTokenId ?? "") : "";

    function navigateToList() {
        navigate.modules(listRoute, { ignorePrevPath: true });
    }

    const { mutate: createSettingAccessToken, isPending: isCreatingSetting } = AccessTokenCommands.useCreateOne({
        onSuccess: () => {
            toast.success("Access token created successfully");
            navigateToList();
        },
    });
    const { mutate: updateSettingAccessToken, isPending: isUpdatingSetting } = AccessTokenCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Access token updated successfully");
            navigateToList();
        },
    });
    const { mutate: createProjectAccessToken, isPending: isCreatingProject } = ProjectAccessTokenCommands.useCreateOne({
        onSuccess: () => {
            toast.success("Project access token created successfully");
            navigateToList();
        },
    });
    const { mutate: updateProjectAccessToken, isPending: isUpdatingProject } = ProjectAccessTokenCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Project access token updated successfully");
            navigateToList();
        },
    });
    const { mutate: testConnection, isPending: isTesting } = AccessTokenCommands.useTestConn({
        onSuccess: () => {
            setTestStatus("succeeded");
        },
        onError: () => {
            setTestStatus("failed");
        },
    });

    const settingDetailQuery = AccessTokenQueries.useFindOneById(
        { id: detailId },
        { enabled: isEditMode && scope.type === "settings" },
    );
    const projectDetailQuery = ProjectAccessTokenQueries.useFindOneById(
        {
            projectID: scope.type === "project" ? scope.projectId : "",
            id: detailId,
        },
        { enabled: isEditMode && scope.type === "project" },
    );
    const detailQuery = scope.type === "project" ? projectDetailQuery : settingDetailQuery;
    const accessToken = detailQuery.data?.data;
    const readOnlyInherited = scope.type === "project" && accessToken?.inherited === true;

    function createPayload(values: CreateOrEditAccessTokenFormOutput) {
        return {
            availableInProjects: scope.type === "project" ? false : values.availableInProjects,
            default: values.default,
            expireAt: values.expireAt,
            kind: values.kind,
            name: values.name,
            user: values.user,
            token: values.token,
            baseURL: values.baseURL,
        };
    }

    function onSubmit(values: CreateOrEditAccessTokenFormOutput) {
        const payload = createPayload(values);

        if (isEditMode && accessToken) {
            const updatePayload = { ...payload, updateVer: accessToken.updateVer };

            if (scope.type === "project") {
                updateProjectAccessToken({
                    projectID: scope.projectId,
                    id: accessToken.id,
                    payload: updatePayload,
                });
                return;
            }

            updateSettingAccessToken({ id: accessToken.id, payload: updatePayload });
            return;
        }

        if (scope.type === "project") {
            createProjectAccessToken({ projectID: scope.projectId, payload });
            return;
        }

        createSettingAccessToken({ payload });
    }

    function onTestConnection(values: CreateOrEditAccessTokenFormOutput) {
        setTestStatus("idle");
        testConnection({
            payload: {
                expireAt: values.expireAt,
                kind: values.kind,
                name: values.name,
                user: values.user,
                token: values.token,
                baseURL: values.baseURL,
            },
        });
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
    const initialValues: Partial<CreateOrEditAccessTokenFormInput> | undefined = accessToken
        ? {
              name: accessToken.name,
              kind: accessToken.kind ?? EAccessTokenKind.Github,
              user: accessToken.user,
              token: accessToken.token,
              baseURL: accessToken.baseURL,
              expireAt: accessToken.expireAt ?? null,
              availableInProjects: accessToken.availableInProjects ?? false,
              default: accessToken.default ?? false,
          }
        : undefined;
    const shouldRenderForm = mode === "create" || initialValues;
    const title = mode === "create" ? "Create Access Token" : "Edit Access Token";

    return (
        <div className="flex w-full flex-col overflow-hidden">
            <div className="flex flex-wrap items-start justify-between gap-3 border-b px-6 py-4">
                <div className="flex min-w-0 flex-col gap-3">
                    <h1 className="text-lg font-semibold text-foreground">{title}</h1>
                </div>
                <Button
                    type="button"
                    variant="outline"
                    onClick={handleClose}
                    disabled={isPending}
                >
                    <ArrowLeft className="size-4" />
                    Back
                </Button>
            </div>

            {isDetailLoading && (
                <div className="flex min-h-[220px] items-center justify-center">
                    <AppLoader />
                </div>
            )}

            {!isDetailLoading && shouldRenderForm && (
                <CreateOrEditAccessTokenForm
                    isPending={isPending}
                    isTesting={isTesting}
                    testStatus={testStatus}
                    onSubmit={onSubmit}
                    onTestConnection={onTestConnection}
                    onHasChanges={setHasChanges}
                    initialValues={initialValues}
                    showAvailableInProjects={scope.type === "settings"}
                    readOnlyInherited={readOnlyInherited}
                    readOnly={!canWrite}
                    onClose={handleClose}
                />
            )}
        </div>
    );
}

function getAccessTokenListRoute(scope: AccessTokenTableScope) {
    if (scope.type === "project") {
        return ROUTE.projects.single.providerConfiguration.accessTokens.$route(scope.projectId);
    }

    return ROUTE.settings.accessTokens.$route;
}

interface Props {
    mode: AccessTokenFormRouteMode;
    scope: AccessTokenTableScope;
    accessTokenId?: string;
}
