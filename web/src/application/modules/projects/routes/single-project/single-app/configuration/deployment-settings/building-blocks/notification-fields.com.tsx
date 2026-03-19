import { Checkbox, FieldError } from "@components/ui";
import { useController, useFormContext } from "react-hook-form";

import { Combobox, InfoBlock } from "@application/shared/components";

import {
    type AppConfigDeploymentSettingsFormSchemaInput,
    type AppConfigDeploymentSettingsFormSchemaOutput,
} from "../schemas";

const MOCK_NOTIFICATIONS = [
    { value: { id: "notif-1", name: "Email Notification" }, label: "Email Notification" },
    { value: { id: "notif-2", name: "Slack Notification" }, label: "Slack Notification" },
];

export function NotificationFields() {
    const { control } = useFormContext<
        AppConfigDeploymentSettingsFormSchemaInput,
        unknown,
        AppConfigDeploymentSettingsFormSchemaOutput
    >();

    const { field: useDefaultOnSuccess } = useController({
        control,
        name: "notification.useDefaultOnSuccess",
        defaultValue: false,
    });

    const {
        field: success,
        fieldState: { invalid: isSuccessInvalid, error: successError },
    } = useController({
        control,
        name: "notification.success",
    });

    const { field: useDefaultOnFailure } = useController({
        control,
        name: "notification.useDefaultOnFailure",
        defaultValue: false,
    });

    const {
        field: failure,
        fieldState: { invalid: isFailureInvalid, error: failureError },
    } = useController({
        control,
        name: "notification.failure",
    });

    return (
        <>
            <InfoBlock
                title="On Success Use Default"
                description="Use the default notification settings on success"
            >
                <Checkbox
                    checked={useDefaultOnSuccess.value}
                    onCheckedChange={useDefaultOnSuccess.onChange}
                />
            </InfoBlock>

            <InfoBlock title="On Success">
                <Combobox
                    options={MOCK_NOTIFICATIONS}
                    value={success.value?.id ?? null}
                    onChange={(_, option) => {
                        success.onChange(option ?? undefined);
                    }}
                    placeholder="None"
                    searchable={false}
                    closeOnSelect
                    emptyText="No notifications available"
                    className="max-w-[400px]"
                    valueKey="id"
                    aria-invalid={isSuccessInvalid}
                />
                <FieldError errors={[successError]} />
            </InfoBlock>

            <InfoBlock
                title="On Failure Use Default"
                description="Use the default notification settings on failure"
            >
                <Checkbox
                    checked={useDefaultOnFailure.value}
                    onCheckedChange={useDefaultOnFailure.onChange}
                />
            </InfoBlock>

            <InfoBlock title="On Failure">
                <Combobox
                    options={MOCK_NOTIFICATIONS}
                    value={failure.value?.id ?? null}
                    onChange={(_, option) => {
                        failure.onChange(option ?? undefined);
                    }}
                    placeholder="None"
                    searchable={false}
                    closeOnSelect
                    emptyText="No notifications available"
                    className="max-w-[400px]"
                    valueKey="id"
                    aria-invalid={isFailureInvalid}
                />
                <FieldError errors={[failureError]} />
            </InfoBlock>
        </>
    );
}
