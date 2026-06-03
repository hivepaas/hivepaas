import React, { useEffect } from "react";

import { Checkbox, Field, FieldError } from "@components/ui";
import { type FieldPath, type FieldValues, useController, useFormContext } from "react-hook-form";

import { AppLink, Combobox, InfoBlock, LabelWithInfo } from "@application/shared/components";

import type {
    NotificationSettingsManageLink,
    NotificationSettingsRef,
    NotificationSettingsSources,
} from "./notification-settings.types";

function NotificationSelect<TFieldValues extends FieldValues>({
    name,
    source,
    title,
    titleWidth,
    readOnly,
    manageLink,
}: NotificationSelectProps<TFieldValues>) {
    const { control } = useFormContext<TFieldValues>();

    const {
        field,
        fieldState: { error, invalid },
    } = useController({ control, name });

    const selectedValue = field.value as NotificationSettingsRef | undefined;

    return (
        <InfoBlock
            title={title}
            titleWidth={titleWidth}
        >
            <Field>
                <Combobox
                    options={source.options}
                    value={selectedValue?.id ?? null}
                    onChange={(_, option) => {
                        if (readOnly) {
                            return;
                        }

                        field.onChange(option ?? undefined);
                    }}
                    onSearch={source.onSearch}
                    placeholder="None"
                    searchable
                    closeOnSelect
                    emptyText="No notifications available"
                    className="max-w-[420px]"
                    valueKey="id"
                    aria-invalid={invalid}
                    loading={source.isFetching}
                    disabled={readOnly}
                    onRefresh={source.onRefresh}
                    isRefreshing={source.isRefreshing}
                    allowClear
                />
                <FieldError errors={[error]} />
                <AppLink.Basic
                    target="_blank"
                    rel="noreferrer"
                    to={manageLink.to}
                    className="text-sm text-link"
                    ignorePrevPath
                >
                    {manageLink.label}
                </AppLink.Basic>
            </Field>
        </InfoBlock>
    );
}

export function NotificationSettings<TFieldValues extends FieldValues>({
    names,
    sources,
    manageLink,
    readOnly = false,
    titleWidth = 270,
    className = "flex flex-col gap-6",
    children,
}: NotificationSettingsProps<TFieldValues>) {
    const { control } = useFormContext<TFieldValues>();

    const { field: successUseDefault } = useController({ control, name: names.successUseDefault });
    const { field: success } = useController({ control, name: names.success });
    const { field: failureUseDefault } = useController({ control, name: names.failureUseDefault });
    const { field: failure } = useController({ control, name: names.failure });

    const isSuccessUseDefault = successUseDefault.value === true;
    const isFailureUseDefault = failureUseDefault.value === true;
    const successValue = success.value as NotificationSettingsRef | undefined;
    const failureValue = failure.value as NotificationSettingsRef | undefined;

    useEffect(() => {
        if (!readOnly && isSuccessUseDefault && successValue) {
            success.onChange(undefined);
        }
    }, [isSuccessUseDefault, readOnly, success, successValue]);

    useEffect(() => {
        if (!readOnly && isFailureUseDefault && failureValue) {
            failure.onChange(undefined);
        }
    }, [failure, failureValue, isFailureUseDefault, readOnly]);

    return (
        <div className={className}>
            <InfoBlock
                titleWidth={titleWidth}
                title={
                    <LabelWithInfo
                        label="On Success Use Default"
                        content="Use the default notification settings on success"
                    />
                }
            >
                <Checkbox
                    checked={isSuccessUseDefault}
                    onCheckedChange={checked => {
                        if (readOnly) {
                            return;
                        }

                        successUseDefault.onChange(checked === true);
                    }}
                    disabled={readOnly}
                />
            </InfoBlock>

            {!isSuccessUseDefault && (
                <NotificationSelect
                    name={names.success}
                    source={sources.success}
                    title="On Success"
                    titleWidth={titleWidth}
                    readOnly={readOnly}
                    manageLink={manageLink}
                />
            )}

            <InfoBlock
                titleWidth={titleWidth}
                title={
                    <LabelWithInfo
                        label="On Failure Use Default"
                        content="Use the default notification settings on failure"
                    />
                }
            >
                <Checkbox
                    checked={isFailureUseDefault}
                    onCheckedChange={checked => {
                        if (readOnly) {
                            return;
                        }

                        failureUseDefault.onChange(checked === true);
                    }}
                    disabled={readOnly}
                />
            </InfoBlock>

            {!isFailureUseDefault && (
                <NotificationSelect
                    name={names.failure}
                    source={sources.failure}
                    title="On Failure"
                    titleWidth={titleWidth}
                    readOnly={readOnly}
                    manageLink={manageLink}
                />
            )}

            {children}
        </div>
    );
}

type NotificationFieldNames<TFieldValues extends FieldValues> = {
    successUseDefault: FieldPath<TFieldValues>;
    success: FieldPath<TFieldValues>;
    failureUseDefault: FieldPath<TFieldValues>;
    failure: FieldPath<TFieldValues>;
};

type NotificationSelectProps<TFieldValues extends FieldValues> = {
    name: FieldPath<TFieldValues>;
    source: NotificationSettingsSources["success"];
    title: string;
    titleWidth: number;
    readOnly: boolean;
    manageLink: NotificationSettingsManageLink;
};

type NotificationSettingsProps<TFieldValues extends FieldValues> = React.PropsWithChildren<{
    names: NotificationFieldNames<TFieldValues>;
    sources: NotificationSettingsSources;
    manageLink: NotificationSettingsManageLink;
    readOnly?: boolean;
    titleWidth?: number;
    className?: string;
}>;
