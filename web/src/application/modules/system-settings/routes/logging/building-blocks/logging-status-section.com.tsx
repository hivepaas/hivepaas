import type { HivePaaSLoggingApp, HivePaaSLoggingStatus } from "~/system-settings/domain";
import { SectionHeader } from "~/system-settings/module-shared";

import { AppLink, InfoBlock, LabelWithInfo } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";

/**
 * One app of the stack: how many of its containers run, and a way to its own
 * screen - it is an ordinary app, with its logs, terminal and resources there.
 */
function LoggingAppStatus({ app, ready }: { app?: HivePaaSLoggingApp; ready: boolean }) {
    if (!app) {
        return <span className="text-muted-foreground">Not run by HivePaaS</span>;
    }
    return (
        <span className="flex flex-wrap items-center gap-x-3 gap-y-1">
            <span>
                {ready ? "Running" : "Not running"} ({app.runningTasks}/{app.desiredTasks})
            </span>
            {app.projectEnv && (
                <AppLink.Basic
                    to={ROUTE.projects.single.apps.single.configuration.general.$route(
                        app.projectId,
                        app.projectEnv,
                        app.appId,
                    )}
                    className="text-link underline-offset-4 hover:underline"
                >
                    Open app
                </AppLink.Basic>
            )}
        </span>
    );
}

export function LoggingStatusSection({ loggingStatus }: { loggingStatus: HivePaaSLoggingStatus }) {
    return (
        <>
            <SectionHeader>Status</SectionHeader>
            <div className="flex flex-col gap-6 px-3">
                <InfoBlock
                    titleWidth={220}
                    title={
                        <LabelWithInfo
                            label="Collector"
                            content="The app that collects logs, one container on every node."
                        />
                    }
                >
                    <LoggingAppStatus
                        app={loggingStatus.collector}
                        ready={loggingStatus.collectorReady}
                    />
                </InfoBlock>
                <InfoBlock
                    titleWidth={220}
                    title={
                        <LabelWithInfo
                            label="Backend"
                            content="The app that stores the logs and answers the apps' log screens."
                        />
                    }
                >
                    <LoggingAppStatus
                        app={loggingStatus.backend}
                        ready={loggingStatus.backendReady}
                    />
                </InfoBlock>
            </div>
        </>
    );
}
