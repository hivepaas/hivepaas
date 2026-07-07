import { useState } from "react";

import { toast } from "sonner";
import { ProjectNotificationCommands } from "~/projects/data/commands";
import { ProjectNotificationQueries } from "~/projects/data/queries";
import { NotificationCommands } from "~/settings/data/commands";
import { NotificationQueries } from "~/settings/data/queries";
import { CreateOrEditNotificationTargetForm } from "~/settings/module-shared/components/notification-target-form";
import type {
    CreateOrEditNotificationTargetFormInput,
    CreateOrEditNotificationTargetFormOutput,
} from "~/settings/module-shared/components/notification-target-form";
import { useSettingsScopePermissions } from "~/settings/module-shared/hooks";
import { SettingsFormRouteHeader } from "~/settings/module-shared/components/settings-form-route-header";

import { AppLoader } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";

import type { NotificationTargetTableScope } from "../notification-target-table";

type NotificationTargetFormRouteMode = "create" | "edit";

function splitCommaSeparated(value: string): string[] {
    return value
        .split(",")
        .map(item => item.trim())
        .filter(Boolean);
}

function joinAddresses(value?: string[]): string {
    return value?.join(", ") ?? "";
}

export function NotificationTargetFormRoute({ mode, scope, notificationTargetId }: Props) {
    const [hasChanges, setHasChanges] = useState(false);
    const [saveRevision, setSaveRevision] = useState(0);
    const { canWrite } = useSettingsScopePermissions(scope);
    const { navigate } = useAppNavigate();

    const listRoute = getNotificationTargetListRoute(scope);
    const isEditMode = mode === "edit";
    const detailId = isEditMode ? (notificationTargetId ?? "") : "";

    function navigateToList() {
        navigate.modules(listRoute, { ignorePrevPath: true });
    }

    function markSaved() {
        setHasChanges(false);
        setSaveRevision(revision => revision + 1);
    }

    const { mutate: createSettingNotification, isPending: isCreatingSetting } = NotificationCommands.useCreateOne({
        onSuccess: () => {
            toast.success("Notification target created successfully");
            markSaved();
        },
    });
    const { mutate: updateSettingNotification, isPending: isUpdatingSetting } = NotificationCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Notification target updated successfully");
            markSaved();
        },
    });
    const { mutate: createProjectNotification, isPending: isCreatingProject } =
        ProjectNotificationCommands.useCreateOne({
            onSuccess: () => {
                toast.success("Project notification target created successfully");
                markSaved();
            },
        });
    const { mutate: updateProjectNotification, isPending: isUpdatingProject } =
        ProjectNotificationCommands.useUpdateOne({
            onSuccess: () => {
                toast.success("Project notification target updated successfully");
                markSaved();
            },
        });

    const settingDetailQuery = NotificationQueries.useFindOneById(
        { id: detailId },
        { enabled: isEditMode && scope.type === "settings" },
    );
    const projectDetailQuery = ProjectNotificationQueries.useFindOneById(
        {
            projectID: scope.type === "project" ? scope.projectId : "",
            id: detailId,
        },
        { enabled: isEditMode && scope.type === "project" },
    );
    const detailQuery = scope.type === "project" ? projectDetailQuery : settingDetailQuery;
    const notificationTarget = detailQuery.data?.data;
    const readOnlyInherited = scope.type === "project" && notificationTarget?.inherited === true;

    function createPayload(values: CreateOrEditNotificationTargetFormOutput) {
        const viaEmail = values.emailEnabled
            ? {
                  enabled: values.emailEnabled,
                  useDefault: values.emailUseDefault,
                  sender: { id: values.emailUseDefault ? "" : values.senderEmailAccountId },
                  toProjectMembers: values.notifyProjectMembers,
                  toProjectOwners: values.notifyProjectOwners,
                  toAllAdmins: values.notifyAdmins,
                  toAddresses: splitCommaSeparated(values.customAddresses),
              }
            : null;
        const viaSlack = values.slackEnabled
            ? {
                  enabled: values.slackEnabled,
                  useDefault: values.slackUseDefault,
                  webhook: { id: values.slackUseDefault ? "" : values.slackWebhookId },
              }
            : null;
        const viaDiscord = values.discordEnabled
            ? {
                  enabled: values.discordEnabled,
                  useDefault: values.discordUseDefault,
                  webhook: { id: values.discordUseDefault ? "" : values.discordWebhookId },
              }
            : null;
        const viaTelegram = values.telegramEnabled
            ? {
                  enabled: values.telegramEnabled,
                  useDefault: values.telegramUseDefault,
                  setting: { id: values.telegramUseDefault ? "" : values.telegramSettingId },
              }
            : null;

        return {
            availableInProjects: scope.type === "project" ? false : values.availableInProjects,
            default: values.default,
            name: values.name,
            viaEmail,
            viaSlack,
            viaDiscord,
            viaTelegram,
            minSendInterval: values.minSendInterval,
        };
    }

    function onSubmit(values: CreateOrEditNotificationTargetFormOutput) {
        const payload = createPayload(values);

        if (isEditMode && notificationTarget) {
            const updatePayload = { ...payload, updateVer: notificationTarget.updateVer };

            if (scope.type === "project") {
                updateProjectNotification({
                    projectID: scope.projectId,
                    id: notificationTarget.id,
                    payload: updatePayload,
                });
                return;
            }

            updateSettingNotification({ id: notificationTarget.id, payload: updatePayload });
            return;
        }

        if (scope.type === "project") {
            createProjectNotification({ projectID: scope.projectId, payload });
            return;
        }

        createSettingNotification({ payload });
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
    const initialValues: Partial<CreateOrEditNotificationTargetFormInput> | undefined = notificationTarget
        ? {
              name: notificationTarget.name,
              emailEnabled: notificationTarget.viaEmail?.enabled ?? false,
              emailUseDefault: notificationTarget.viaEmail?.useDefault ?? true,
              senderEmailAccountId: notificationTarget.viaEmail?.sender?.id ?? "",
              notifyAdmins: notificationTarget.viaEmail?.toAllAdmins ?? false,
              notifyProjectOwners: notificationTarget.viaEmail?.toProjectOwners ?? false,
              notifyProjectMembers: notificationTarget.viaEmail?.toProjectMembers ?? false,
              customAddresses: joinAddresses(notificationTarget.viaEmail?.toAddresses),
              slackEnabled: notificationTarget.viaSlack?.enabled ?? false,
              slackUseDefault: notificationTarget.viaSlack?.useDefault ?? true,
              slackWebhookId: notificationTarget.viaSlack?.webhook?.id ?? "",
              discordEnabled: notificationTarget.viaDiscord?.enabled ?? false,
              discordUseDefault: notificationTarget.viaDiscord?.useDefault ?? true,
              discordWebhookId: notificationTarget.viaDiscord?.webhook?.id ?? "",
              telegramEnabled: notificationTarget.viaTelegram?.enabled ?? false,
              telegramUseDefault: notificationTarget.viaTelegram?.useDefault ?? true,
              telegramSettingId: notificationTarget.viaTelegram?.setting?.id ?? "",
              minSendInterval: notificationTarget.minSendInterval || "3m",
              availableInProjects: notificationTarget.availableInProjects ?? false,
              default: notificationTarget.default ?? false,
          }
        : undefined;
    const shouldRenderForm = mode === "create" || initialValues;
    const title = mode === "create" ? "Create Notification Target" : "Edit Notification Target";

    return (
        <div className="flex w-full flex-col overflow-hidden">
            <SettingsFormRouteHeader title={title} />

            {isDetailLoading && (
                <div className="flex min-h-[220px] items-center justify-center">
                    <AppLoader />
                </div>
            )}

            {!isDetailLoading && shouldRenderForm && (
                <CreateOrEditNotificationTargetForm
                    scope={scope}
                    isPending={isPending}
                    onSubmit={onSubmit}
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

function getNotificationTargetListRoute(scope: NotificationTargetTableScope) {
    if (scope.type === "project") {
        return ROUTE.projects.single.providerConfiguration.notificationTargets.$route(scope.projectId);
    }

    return ROUTE.settings.notificationTargets.$route;
}

interface Props {
    mode: NotificationTargetFormRouteMode;
    scope: NotificationTargetTableScope;
    notificationTargetId?: string;
}
