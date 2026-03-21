import { Checkbox } from "@components/ui";
import { useController, useFormContext } from "react-hook-form";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";

import { NotificationSelect } from "../form-components";
import {
    type AppConfigDeploymentSettingsFormSchemaInput,
    type AppConfigDeploymentSettingsFormSchemaOutput,
} from "../schemas";

export function NotificationFields() {
    const { control } = useFormContext<
        AppConfigDeploymentSettingsFormSchemaInput,
        unknown,
        AppConfigDeploymentSettingsFormSchemaOutput
    >();

    const { field: useDefaultOnSuccess } = useController({
        control,
        name: "notification.successUseDefault",
        defaultValue: false,
    });

    const { field: useDefaultOnFailure } = useController({
        control,
        name: "notification.failureUseDefault",
        defaultValue: false,
    });

    return (
        <>
            <InfoBlock
                title={
                    <LabelWithInfo
                        label="On Success Use Default"
                        content="Use the default notification settings on success"
                    />
                }
            >
                <Checkbox
                    checked={useDefaultOnSuccess.value}
                    onCheckedChange={useDefaultOnSuccess.onChange}
                />
            </InfoBlock>

            <NotificationSelect
                name="notification.success"
                title="On Success"
            />

            <InfoBlock
                title={
                    <LabelWithInfo
                        label="On Failure Use Default"
                        content="Use the default notification settings on failure"
                    />
                }
            >
                <Checkbox
                    checked={useDefaultOnFailure.value}
                    onCheckedChange={useDefaultOnFailure.onChange}
                />
            </InfoBlock>

            <NotificationSelect
                name="notification.failure"
                title="On Failure"
            />
        </>
    );
}
