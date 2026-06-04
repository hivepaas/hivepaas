import type { ReactNode } from "react";

import { useParams } from "react-router";
import invariant from "tiny-invariant";
import { APP_CONFIGURATION_QUERY_OPTIONS } from "~/projects/data/constants";
import { ProjectAppsQueries } from "~/projects/data/queries";
import { AppDangerAction, useConfirmAppDangerActionDialog } from "~/projects/dialogs/confirm-app-danger-action";
import { EProjectAppStatus } from "~/projects/module-shared/enums";

import { AppLoader } from "@application/shared/components";
import { MODULE_IDS } from "@application/shared/constants";
import { PageError } from "@application/shared/pages";
import { PermissionTooltipAction } from "@application/shared/permissions";

import { Button } from "@/components/ui/button";

export function AppConfigDangerZoneRoute() {
    const { id: projectId, appId } = useParams<{ id: string; appId: string }>();
    const { actions: confirmAction } = useConfirmAppDangerActionDialog();

    invariant(projectId, "projectId must be defined");
    invariant(appId, "appId must be defined");

    const { data, isLoading, error, refetch } = ProjectAppsQueries.useFindOneById(
        {
            projectID: projectId,
            appID: appId,
        },
        APP_CONFIGURATION_QUERY_OPTIONS,
    );

    if (isLoading) {
        return <AppLoader />;
    }

    if (error) {
        return (
            <PageError
                error={error}
                onRetry={() => {
                    void refetch();
                }}
            />
        );
    }

    invariant(data, "app data must be defined");

    const app = data.data;
    const isDisabled = app.status === EProjectAppStatus.Disabled;
    const isDeleting = app.status === EProjectAppStatus.Deleting;
    const isStatusActionable = app.status === EProjectAppStatus.Active || isDisabled;
    const statusAction = isDisabled ? AppDangerAction.ReEnable : AppDangerAction.Disable;
    const statusButtonLabel = isDisabled ? "Re-enable App" : "Disable App";
    const statusButtonVariant = isDisabled ? "default" : "destructive";

    const target = {
        projectId,
        appId,
        appName: app.name,
        updateVer: app.updateVer,
    };

    return (
        <div className="flex flex-col gap-5">
            <DangerActionPanel>
                <p className="text-base font-semibold leading-7 text-foreground">
                    Disabling an application means setting its number of instances to 0, so the application will no
                    longer consume system resources such as CPU or memory. However, the application&apos;s information
                    will still remain in the system, and you can restore it at any time.
                </p>

                <PermissionTooltipAction
                    id={MODULE_IDS.Project}
                    action="write"
                >
                    {({ isDenied }) => (
                        <Button
                            variant={statusButtonVariant}
                            disabled={isDenied || !isStatusActionable}
                            className="min-w-[150px]"
                            onClick={() => {
                                if (isDenied || !isStatusActionable) {
                                    return;
                                }

                                confirmAction.open(statusAction, target);
                            }}
                        >
                            {statusButtonLabel}
                        </Button>
                    )}
                </PermissionTooltipAction>
            </DangerActionPanel>

            <DangerActionPanel>
                <p className="text-base font-semibold leading-7 text-foreground">
                    Deleting an application will remove all of its information and release all system resources
                    allocated to it. You will not be able to recover it after deletion.
                </p>

                <PermissionTooltipAction
                    id={MODULE_IDS.Project}
                    action="write"
                >
                    {({ isDenied }) => (
                        <Button
                            variant="destructive"
                            disabled={isDenied || isDeleting}
                            className="min-w-[150px]"
                            onClick={() => {
                                if (isDenied || isDeleting) {
                                    return;
                                }

                                confirmAction.open(AppDangerAction.Delete, target);
                            }}
                        >
                            Delete App
                        </Button>
                    )}
                </PermissionTooltipAction>
            </DangerActionPanel>
        </div>
    );
}

function DangerActionPanel({ children }: { children: ReactNode }) {
    return (
        <section className="flex flex-col items-start gap-6 rounded-lg border bg-background p-5">{children}</section>
    );
}
