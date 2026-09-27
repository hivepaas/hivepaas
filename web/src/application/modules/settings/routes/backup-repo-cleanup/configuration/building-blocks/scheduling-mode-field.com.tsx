import { useController, useFormContext } from "react-hook-form";

import { InfoBlock } from "@application/shared/components";

import { Tabs, TabsList, TabsTrigger } from "@/components/ui";

import {
    type SystemBackupRepoCleanupConfigurationFormInput,
    type SystemBackupRepoCleanupConfigurationFormOutput,
    SystemBackupRepoCleanupScheduleMode,
} from "../schemas";

type SchemaInput = SystemBackupRepoCleanupConfigurationFormInput;
type SchemaOutput = SystemBackupRepoCleanupConfigurationFormOutput;

export function SchedulingModeField() {
    const { control } = useFormContext<SchemaInput, unknown, SchemaOutput>();
    const { field: scheduleMode } = useController({ control, name: "scheduleMode" });

    return (
        <InfoBlock
            titleWidth={220}
            title="Scheduling Mode"
        >
            <Tabs
                value={scheduleMode.value}
                onValueChange={scheduleMode.onChange}
            >
                <TabsList>
                    <TabsTrigger value={SystemBackupRepoCleanupScheduleMode.Interval}>Interval-based</TabsTrigger>
                    <TabsTrigger value={SystemBackupRepoCleanupScheduleMode.Cron}>Time-based</TabsTrigger>
                </TabsList>
            </Tabs>
        </InfoBlock>
    );
}
