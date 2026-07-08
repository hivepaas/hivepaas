import { useState } from "react";

import { toast } from "sonner";
import { ProjectImServiceCommands } from "~/projects/data/commands";
import { ProjectImServiceQueries } from "~/projects/data/queries";
import { ImServiceCommands } from "~/settings/data/commands";
import { ImServiceQueries } from "~/settings/data/queries";
import { CreateOrEditImPlatformForm } from "~/settings/module-shared/components/im-platform-form";
import type {
    CreateOrEditImPlatformFormInput,
    CreateOrEditImPlatformFormOutput,
} from "~/settings/module-shared/components/im-platform-form";
import { useSettingsScopePermissions } from "~/settings/module-shared/hooks";
import { SettingsFormRouteHeader } from "~/settings/module-shared/components/settings-form-route-header";

import { AppLoader } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";
import { EImServiceKind } from "@application/shared/enums";
import { useAppNavigate } from "@application/shared/hooks/router";

import type { ImPlatformTableScope } from "../im-platform-table";

type ImPlatformFormRouteMode = "create" | "edit";

export function ImPlatformFormRoute({ mode, scope, imPlatformId }: Props) {
    const [hasChanges, setHasChanges] = useState(false);
    const [saveRevision, setSaveRevision] = useState(0);
    const [testStatus, setTestStatus] = useState<"idle" | "succeeded" | "failed">("idle");
    const { canWrite } = useSettingsScopePermissions(scope);
    const { navigate } = useAppNavigate();

    const listRoute = getImPlatformListRoute(scope);
    const isEditMode = mode === "edit";
    const detailId = isEditMode ? (imPlatformId ?? "") : "";

    function navigateToList() {
        navigate.modules(listRoute, { ignorePrevPath: true });
    }

    function markSaved() {
        setHasChanges(false);
        setSaveRevision(revision => revision + 1);
    }

    const { mutate: createSettingImPlatform, isPending: isCreatingSetting } = ImServiceCommands.useCreateOne({
        onSuccess: () => {
            toast.success("IM platform created successfully");
            markSaved();
        },
    });
    const { mutate: updateSettingImPlatform, isPending: isUpdatingSetting } = ImServiceCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("IM platform updated successfully");
            markSaved();
        },
    });
    const { mutate: createProjectImPlatform, isPending: isCreatingProject } = ProjectImServiceCommands.useCreateOne({
        onSuccess: () => {
            toast.success("Project IM platform created successfully");
            markSaved();
        },
    });
    const { mutate: updateProjectImPlatform, isPending: isUpdatingProject } = ProjectImServiceCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Project IM platform updated successfully");
            markSaved();
        },
    });
    const { mutate: testSendMsg, isPending: isTesting } = ImServiceCommands.useTestSendMsg({
        onSuccess: () => {
            setTestStatus("succeeded");
        },
        onError: () => {
            setTestStatus("failed");
        },
    });

    const settingDetailQuery = ImServiceQueries.useFindOneById(
        { id: detailId },
        { enabled: isEditMode && scope.type === "settings" },
    );
    const projectDetailQuery = ProjectImServiceQueries.useFindOneById(
        {
            projectID: scope.type === "project" ? scope.projectId : "",
            id: detailId,
        },
        { enabled: isEditMode && scope.type === "project" },
    );
    const detailQuery = scope.type === "project" ? projectDetailQuery : settingDetailQuery;
    const imPlatform = detailQuery.data?.data;
    const readOnlyInherited = scope.type === "project" && imPlatform?.inherited === true;

    function createPayload(values: CreateOrEditImPlatformFormOutput) {
        const basePayload = {
            availableInProjects: scope.type === "project" ? false : values.availableInProjects,
            default: values.default,
            name: values.name,
            kind: values.kind,
        };

        if (values.kind === EImServiceKind.Slack) {
            return {
                ...basePayload,
                slack: { webhook: values.webhook },
                discord: null,
                telegram: null,
            };
        }

        if (values.kind === EImServiceKind.Discord) {
            return {
                ...basePayload,
                slack: null,
                discord: { webhook: values.webhook },
                telegram: null,
            };
        }

        return {
            ...basePayload,
            slack: null,
            discord: null,
            telegram: {
                botToken: values.botToken,
                chatId: values.chatId,
            },
        };
    }

    function onSubmit(values: CreateOrEditImPlatformFormOutput) {
        const payload = createPayload(values);

        if (isEditMode && imPlatform) {
            const updatePayload = { ...payload, updateVer: imPlatform.updateVer };

            if (scope.type === "project") {
                updateProjectImPlatform({
                    projectID: scope.projectId,
                    id: imPlatform.id,
                    payload: updatePayload,
                });
                return;
            }

            updateSettingImPlatform({ id: imPlatform.id, payload: updatePayload });
            return;
        }

        if (scope.type === "project") {
            createProjectImPlatform({ projectID: scope.projectId, payload });
            return;
        }

        createSettingImPlatform({ payload });
    }

    function onTestSendMsg(values: CreateOrEditImPlatformFormOutput) {
        setTestStatus("idle");
        testSendMsg({
            payload: {
                ...createPayload(values),
                testMsg: "test message",
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
    const initialValues: Partial<CreateOrEditImPlatformFormInput> | undefined = imPlatform
        ? {
              name: imPlatform.name,
              kind: imPlatform.kind,
              webhook:
                  imPlatform.kind === EImServiceKind.Slack
                      ? (imPlatform.slack?.webhook ?? "")
                      : imPlatform.kind === EImServiceKind.Discord
                        ? (imPlatform.discord?.webhook ?? "")
                        : "",
              botToken: imPlatform.kind === EImServiceKind.Telegram ? (imPlatform.telegram?.botToken ?? "") : "",
              chatId: imPlatform.kind === EImServiceKind.Telegram ? (imPlatform.telegram?.chatId ?? "") : "",
              availableInProjects: imPlatform.availableInProjects ?? false,
              default: imPlatform.default ?? false,
          }
        : undefined;
    const shouldRenderForm = mode === "create" || initialValues;
    const title = mode === "create" ? "Create IM Platform" : "Edit IM Platform";

    return (
        <div className="flex w-full flex-col">
            <SettingsFormRouteHeader title={title} />

            {isDetailLoading && (
                <div className="flex min-h-[220px] items-center justify-center">
                    <AppLoader />
                </div>
            )}

            {!isDetailLoading && shouldRenderForm && (
                <CreateOrEditImPlatformForm
                    isPending={isPending}
                    isTesting={isTesting}
                    testStatus={testStatus}
                    onSubmit={onSubmit}
                    onTestSendMsg={onTestSendMsg}
                    onHasChanges={setHasChanges}
                    savedVersion={saveRevision}
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

function getImPlatformListRoute(scope: ImPlatformTableScope) {
    if (scope.type === "project") {
        return ROUTE.projects.single.providerConfiguration.imPlatforms.$route(scope.projectId);
    }

    return ROUTE.settings.imPlatforms.$route;
}

interface Props {
    mode: ImPlatformFormRouteMode;
    scope: ImPlatformTableScope;
    imPlatformId?: string;
}
