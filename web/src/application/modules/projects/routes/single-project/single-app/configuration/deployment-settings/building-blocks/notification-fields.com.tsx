import { useEffect } from "react";

import { Checkbox } from "@components/ui";
import { useController, useFormContext } from "react-hook-form";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";

import { NotificationSelect } from "../form-components";
import {
    type AppConfigDeploymentSettingsFormSchemaInput,
    type AppConfigDeploymentSettingsFormSchemaOutput,
} from "../schemas";

export function NotificationFields({ readOnly = false }: Props) {
    const { control } = useFormContext<
        AppConfigDeploymentSettingsFormSchemaInput,
        unknown,
        AppConfigDeploymentSettingsFormSchemaOutput
    >();

    const { field: useDefaultOnSuccess } = useController({
        control,
        name: "notification.successUseDefault",
        defaultValue: true,
    });

    const { field: success } = useController({
        control,
        name: "notification.success",
    });

    const { field: useDefaultOnFailure } = useController({
        control,
        name: "notification.failureUseDefault",
        defaultValue: true,
    });

    const { field: failure } = useController({
        control,
        name: "notification.failure",
    });

    useEffect(() => {
        if (!readOnly && useDefaultOnSuccess.value) {
            success.onChange(undefined);
        }
    }, [readOnly, useDefaultOnSuccess.value, success]);

    useEffect(() => {
        if (!readOnly && useDefaultOnFailure.value) {
            failure.onChange(undefined);
        }
    }, [readOnly, useDefaultOnFailure.value, failure]);

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
                    onCheckedChange={value => {
                        if (readOnly) {
                            return;
                        }

                        useDefaultOnSuccess.onChange(value);
                    }}
                    disabled={readOnly}
                />
            </InfoBlock>

            <NotificationSelect
                name="notification.success"
                title="On Success"
                disabled={readOnly || useDefaultOnSuccess.value}
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
                    onCheckedChange={value => {
                        if (readOnly) {
                            return;
                        }

                        useDefaultOnFailure.onChange(value);
                    }}
                    disabled={readOnly}
                />
            </InfoBlock>

            <NotificationSelect
                name="notification.failure"
                title="On Failure"
                disabled={readOnly || useDefaultOnFailure.value}
            />
        </>
    );
}

type Props = {
    readOnly?: boolean;
};
