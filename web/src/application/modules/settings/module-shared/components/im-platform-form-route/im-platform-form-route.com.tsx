import { useState } from "react";

import { toast } from "sonner";
import { ProjectImServiceCommands } from "~/projects/data/commands";
import { ProjectImServiceQueries } from "~/projects/data/queries";
import { ImServiceCommands } from "~/settings/data/commands";
import { ImServiceQueries } from "~/settings/data/queries";
import type { SettingImService } from "~/settings/domain";
import { CreateOrEditImPlatformForm } from "~/settings/module-shared/components/im-platform-form";
import type {
    CreateOrEditImPlatformFormInput,
    CreateOrEditImPlatformFormOutput,
} from "~/settings/module-shared/components/im-platform-form";
import { SettingsFormRouteHeader } from "~/settings/module-shared/components/settings-form-route-header";
import { useSettingRevealSecrets, useSettingsScopePermissions } from "~/settings/module-shared/hooks";

import { AppLoader } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";
import { EImServiceKind } from "@application/shared/enums";
import { useAppNavigate } from "@application/shared/hooks/router";

import { RevealSecretsProvider } from "@/components/ui/input-password";

import { ConfirmRevealSecretsDialog } from "../confirm-reveal-secrets-dialog";
import type { ImPlatformTableScope } from "../im-platform-table";
import { RevealSecretsButton } from "../reveal-secrets-button";

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
            env: scope.type === "project" ? scope.env : undefined,
            id: detailId,
        },
        { enabled: isEditMode && scope.type === "project" },
    );
    const detailQuery = scope.type === "project" ? projectDetailQuery : settingDetailQuery;
    const imPlatform = detailQuery.data?.data;
    const readOnlyInherited = scope.type === "project" && imPlatform?.inherited === true;

    function createPayload(values: CreateOrEditImPlatformFormOutput) {
        const basePayload = {
            inheritable: values.inheritable,
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
                lark: null,
            };
        }

        if (values.kind === EImServiceKind.Discord) {
            return {
                ...basePayload,
                slack: null,
                discord: { webhook: values.webhook },
                telegram: null,
                lark: null,
            };
        }

        if (values.kind === EImServiceKind.Lark) {
            return {
                ...basePayload,
                slack: null,
                discord: null,
                telegram: null,
                lark: {
                    webhook: values.webhook,
                    secret: values.secret ?? null,
                },
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
            lark: null,
        };
    }

    function onSubmit(values: CreateOrEditImPlatformFormOutput) {
        const payload = createPayload(values);

        if (isEditMode && imPlatform) {
            const updatePayload = { ...payload, updateVer: imPlatform.updateVer };

            if (scope.type === "project") {
                updateProjectImPlatform({
                    projectID: scope.projectId,
                    env: scope.env,
                    id: imPlatform.id,
                    payload: updatePayload,
                });
                return;
            }

            updateSettingImPlatform({ id: imPlatform.id, payload: updatePayload });
            return;
        }

        if (scope.type === "project") {
            createProjectImPlatform({
                projectID: scope.projectId,
                env: scope.env,
                payload,
            });
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

        navigateToList();
    }

    const {
        canShowRevealButton,
        isDialogOpen,
        setIsDialogOpen,
        isRevealing,
        isRevealed,
        revealedData,
        revealRevision,
        handleConfirmReveal,
    } = useSettingRevealSecrets<SettingImService>({
        settingType: "im-services",
        settingId: detailId,
        scope,
        isInherited: readOnlyInherited || imPlatform?.inherited === true,
        mode,
    });

    const activeImPlatform = revealedData ?? imPlatform;

    const isPending = isCreatingSetting || isUpdatingSetting || isCreatingProject || isUpdatingProject;
    const isDetailLoading = isEditMode && detailQuery.isFetching;
    const initialValues: Partial<CreateOrEditImPlatformFormInput> | undefined = activeImPlatform
        ? {
              name: activeImPlatform.name,
              kind: activeImPlatform.kind,
              webhook:
                  activeImPlatform.kind === EImServiceKind.Slack
                      ? (activeImPlatform.slack?.webhook ?? "")
                      : activeImPlatform.kind === EImServiceKind.Discord
                        ? (activeImPlatform.discord?.webhook ?? "")
                        : activeImPlatform.kind === EImServiceKind.Lark
                          ? (activeImPlatform.lark?.webhook ?? "")
                          : "",
              secret: activeImPlatform.kind === EImServiceKind.Lark ? (activeImPlatform.lark?.secret ?? "") : "",
              botToken:
                  activeImPlatform.kind === EImServiceKind.Telegram ? (activeImPlatform.telegram?.botToken ?? "") : "",
              chatId:
                  activeImPlatform.kind === EImServiceKind.Telegram ? (activeImPlatform.telegram?.chatId ?? "") : "",
              inheritable: Boolean(activeImPlatform.inheritable),
              default: activeImPlatform.default ?? false,
          }
        : undefined;
    const shouldRenderForm = mode === "create" || Boolean(initialValues);
    const title = mode === "create" ? "Create IM Platform" : "Edit IM Platform";

    return (
        <div className="flex w-full flex-col">
            <SettingsFormRouteHeader
                title={title}
                actions={
                    canShowRevealButton ? (
                        <RevealSecretsButton
                            onClick={() => {
                                setIsDialogOpen(true);
                            }}
                            isLoading={isRevealing}
                        />
                    ) : undefined
                }
            />

            <ConfirmRevealSecretsDialog
                open={isDialogOpen}
                onOpenChange={setIsDialogOpen}
                onConfirm={handleConfirmReveal}
                isPending={isRevealing}
            />

            {isDetailLoading && (
                <div className="flex min-h-[220px] items-center justify-center">
                    <AppLoader />
                </div>
            )}

            {!isDetailLoading && shouldRenderForm && (
                <RevealSecretsProvider value={{ isRevealed }}>
                    <CreateOrEditImPlatformForm
                        key={`${detailId}-${revealRevision}`}
                        isPending={isPending}
                        isTesting={isTesting}
                        testStatus={testStatus}
                        onSubmit={onSubmit}
                        onTestSendMsg={onTestSendMsg}
                        onHasChanges={setHasChanges}
                        savedVersion={saveRevision}
                        initialValues={initialValues}
                        showAvailableInProjects
                        isProjectScope={scope.type === "project"}
                        readOnlyInherited={readOnlyInherited}
                        readOnly={!canWrite}
                        onClose={handleClose}
                    />
                </RevealSecretsProvider>
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
