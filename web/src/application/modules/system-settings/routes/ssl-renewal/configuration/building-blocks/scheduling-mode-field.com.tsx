import { useController, useFormContext } from "react-hook-form";

import { InfoBlock } from "@application/shared/components";

import { Tabs, TabsList, TabsTrigger } from "@/components/ui";

import {
    type SystemSslRenewalConfigurationFormInput,
    type SystemSslRenewalConfigurationFormOutput,
    SystemSslRenewalScheduleMode,
} from "../schemas";

type SchemaInput = SystemSslRenewalConfigurationFormInput;
type SchemaOutput = SystemSslRenewalConfigurationFormOutput;

export function SchedulingModeField() {
    const { control } = useFormContext<SchemaInput, unknown, SchemaOutput>();
    const { field: scheduleMode } = useController({ control, name: "scheduleMode" });

    return (
        <InfoBlock title="Scheduling Mode">
            <Tabs
                value={scheduleMode.value}
                onValueChange={scheduleMode.onChange}
            >
                <TabsList>
                    <TabsTrigger value={SystemSslRenewalScheduleMode.Interval}>Interval-based</TabsTrigger>
                    <TabsTrigger value={SystemSslRenewalScheduleMode.Cron}>Time-based</TabsTrigger>
                </TabsList>
            </Tabs>
        </InfoBlock>
    );
}
