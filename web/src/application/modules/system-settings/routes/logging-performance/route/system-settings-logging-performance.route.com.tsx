import { Button } from "@components/ui";
import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { toast } from "sonner";
import { HivePaaSLoggingPerformanceCommands, HivePaaSLoggingPerformanceQueries } from "~/system-settings/data";

import { AppLink, AppLoader, FormActionBar } from "@application/shared/components";
import { MODULE_IDS, ROUTE } from "@application/shared/constants";
import { PageError } from "@application/shared/pages";
import { PermissionTooltipAction, useConditionalModule } from "@application/shared/permissions";

import { LoggingPerformanceForm, toLoggingPerformancePayload } from "../form";
import type { LoggingPerformanceFormOutput } from "../schemas";

/**
 * Which nodes measure apps' routes and calls, by OBI, and at which capacity. Saved apart from the logging
 * settings' page, with their version: nothing is deployed, each node's agent applies it on its own.
 */
export function SystemSettingsLoggingPerformanceRoute() {
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.System });
    const { data, isLoading, error, refetch } = HivePaaSLoggingPerformanceQueries.useFindOne();
    const settings = data?.data;

    const { mutate: update, isPending } = HivePaaSLoggingPerformanceCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Routes and calls settings saved: each node's agent applies them within 30 seconds");
        },
    });

    function handleSubmit(values: LoggingPerformanceFormOutput) {
        if (!canWrite || !settings) {
            return;
        }
        update({ payload: toLoggingPerformancePayload(values, settings.updateVer) });
    }

    if (error) {
        return (
            <PageError
                error={error}
                onRetry={refetch}
            />
        );
    }
    if (isLoading || !settings) {
        return <AppLoader />;
    }

    if (!settings.configured) {
        return (
            <div className={cn(dashedBorderBox, "mt-2 flex flex-col gap-1")}>
                <span>Routes and calls are stored with the logs. Turn logging on and save its settings first.</span>
                <AppLink.Modules
                    to={ROUTE.systemSettings.logging.configuration.$route}
                    className="text-link"
                >
                    Logging configuration
                </AppLink.Modules>
            </div>
        );
    }

    return (
        <>
            {!settings.logsStored && (
                <div className={cn(dashedBorderBox, "mt-2 flex flex-col gap-1")}>
                    <span>
                        <span className="font-semibold text-orange-500">Note:</span> Logs are not stored, so OBI does
                        not run: its numbers are kept with them. Turn logging on, with container logs, to start it.
                    </span>
                    <AppLink.Modules
                        to={ROUTE.systemSettings.logging.configuration.$route}
                        className="text-link"
                    >
                        Logging configuration
                    </AppLink.Modules>
                </div>
            )}
            <LoggingPerformanceForm
                settings={settings}
                onSubmit={handleSubmit}
                readOnly={!canWrite}
            >
                <FormActionBar>
                    <PermissionTooltipAction
                        id={MODULE_IDS.System}
                        action="write"
                    >
                        {({ isDenied }) => (
                            <Button
                                type="submit"
                                className="min-w-[100px]"
                                disabled={isPending || isDenied}
                                isLoading={isPending}
                            >
                                Save
                            </Button>
                        )}
                    </PermissionTooltipAction>
                </FormActionBar>
            </LoggingPerformanceForm>
        </>
    );
}
