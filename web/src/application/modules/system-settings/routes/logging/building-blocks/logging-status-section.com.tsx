import type { HivePaaSLoggingStatus } from "~/system-settings/domain";
import { SectionHeader } from "~/system-settings/module-shared";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";

export function LoggingStatusSection({ loggingStatus }: { loggingStatus: HivePaaSLoggingStatus }) {
    return (
        <>
            <SectionHeader>Status</SectionHeader>
            <div className="flex flex-col gap-6 px-3">
                <InfoBlock
                    titleWidth={220}
                    title={
                        <LabelWithInfo
                            label="Collector running"
                            content="Whether the log collector daemon is running across the cluster."
                        />
                    }
                >
                    {loggingStatus.collectorReady ? "Yes" : "No"}
                </InfoBlock>
                <InfoBlock
                    titleWidth={220}
                    title={
                        <LabelWithInfo
                            label="Backend running"
                            content="Whether the log store's service exists in the cluster."
                        />
                    }
                >
                    {loggingStatus.backendReady ? "Yes" : "No"}
                </InfoBlock>
            </div>
        </>
    );
}
