import { FieldError } from "@components/ui";
import { useController, useFormContext } from "react-hook-form";
import { PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS } from "~/projects/module-shared/constants";

import { EditableCombobox, InfoBlock, LabelWithInfo } from "@application/shared/components";
import { KeyValueList } from "@application/shared/form";

import {
    type AppConfigContainerSettingsFormSchemaInput,
    type AppConfigContainerSettingsFormSchemaOutput,
} from "../schemas";

const LOG_DRIVER_OPTIONS = [
    "awslogs",
    "etwlogs",
    "fluentd",
    "gcplogs",
    "gelf",
    "journald",
    "json-file",
    "local",
    "splunk",
    "syslog",
];

export function LogDriverFields() {
    const { control } = useFormContext<
        AppConfigContainerSettingsFormSchemaInput,
        unknown,
        AppConfigContainerSettingsFormSchemaOutput
    >();

    const {
        field: driver,
        fieldState: { error: driverError, invalid: isDriverInvalid },
    } = useController({ control, name: "logDriver.driver" });

    return (
        <div className="flex flex-col gap-6">
            <InfoBlock
                title={
                    <LabelWithInfo
                        label="Driver"
                        content="Docker logging driver. Select a common driver or type a custom one."
                    />
                }
            >
                <div className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}>
                    <EditableCombobox
                        options={LOG_DRIVER_OPTIONS}
                        value={driver.value}
                        onChange={driver.onChange}
                        placeholder="json-file"
                        aria-invalid={isDriverInvalid}
                    />
                    <FieldError errors={[driverError]} />
                </div>
            </InfoBlock>

            <InfoBlock
                title={
                    <LabelWithInfo
                        label="Options"
                        content="Key-value logging options passed to the selected driver."
                    />
                }
            >
                <KeyValueList<AppConfigContainerSettingsFormSchemaInput>
                    name="logDriver.options"
                    className="max-w-[600px]"
                    checkDuplicates
                    enableValueEditing
                />
            </InfoBlock>
        </div>
    );
}
