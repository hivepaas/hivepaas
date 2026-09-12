import type { HivePaaSLoggingExcludedApp, HivePaaSLoggingStatus } from "~/system-settings/domain";

import { InfoBlock } from "@application/shared/components";

import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui";

import { SectionTitle } from "./section-title.com";

function reasonText(app: HivePaaSLoggingExcludedApp) {
    switch (app.reason) {
        case "driver-unreadable": {
            // An empty driver is the daemon's default, not a missing value.
            const driver = app.driver === undefined || app.driver === "" ? "daemon default" : app.driver;
            return `Its log driver (${driver}) cannot be read. Switch it to json-file in the app's container settings.`;
        }
        case "identity-missing":
            return "Created before logging existed. Save its container settings once to add its identity.";
        default:
            return app.reason;
    }
}

export function LoggingStatusSection({ status }: { status: HivePaaSLoggingStatus }) {
    return (
        <section className="flex flex-col gap-4">
            <SectionTitle>Status</SectionTitle>
            <InfoBlock
                titleWidth={220}
                title="Backend running"
            >
                {status.backendReady ? "Yes" : "No"}
            </InfoBlock>
            {status.excludedApps.length > 0 && (
                <InfoBlock
                    titleWidth={220}
                    title="Not collected"
                    description="These apps' logs will not show up in stored logs."
                >
                    <Table>
                        <TableHeader>
                            <TableRow>
                                <TableHead>App</TableHead>
                                <TableHead>Reason</TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {status.excludedApps.map(app => (
                                <TableRow key={app.appId}>
                                    <TableCell className="font-medium">{app.name}</TableCell>
                                    <TableCell className="whitespace-normal">{reasonText(app)}</TableCell>
                                </TableRow>
                            ))}
                        </TableBody>
                    </Table>
                </InfoBlock>
            )}
        </section>
    );
}
