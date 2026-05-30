import React from "react";

import { Tabs, TabsList, TabsTrigger } from "@components/ui/tabs";
import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldErrors, useController, useForm, useFormState } from "react-hook-form";
import { useUpdateEffect } from "react-use";

import { InfoBlock } from "@application/shared/components";
import { ESettingStatus } from "@application/shared/enums";

import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { DateTimePicker } from "@/components/ui/date-time-picker";
import { Field, FieldError, FieldGroup } from "@/components/ui/field";

import type { UpdateAppHealthCheckStatusFormInput, UpdateAppHealthCheckStatusFormOutput } from "../schemas";
import { UpdateAppHealthCheckStatusFormSchema } from "../schemas";

const statusMap = {
    [ESettingStatus.Active]: "Active",
    [ESettingStatus.Disabled]: "Disabled",
} as const;

export function UpdateAppHealthCheckStatusForm({
    isPending,
    onSubmit,
    initialValues,
    onHasChanges,
    readOnly = false,
}: Props) {
    const {
        handleSubmit,
        control,
        formState: { errors },
    } = useForm<UpdateAppHealthCheckStatusFormInput, unknown, UpdateAppHealthCheckStatusFormOutput>({
        defaultValues: {
            status: initialValues?.status === ESettingStatus.Disabled ? ESettingStatus.Disabled : ESettingStatus.Active,
            expireAt: initialValues?.expireAt ?? null,
            default: initialValues?.default ?? false,
        },
        resolver: zodResolver(UpdateAppHealthCheckStatusFormSchema),
        mode: "onSubmit",
    });

    const { isDirty } = useFormState({ control });

    useUpdateEffect(() => {
        onHasChanges?.(readOnly ? false : isDirty);
    }, [isDirty, readOnly]);

    const { field: status } = useController({
        name: "status",
        control,
    });
    const {
        field: expireAt,
        fieldState: { invalid: isExpireAtInvalid },
    } = useController({
        name: "expireAt",
        control,
    });
    const { field: defaultField } = useController({
        name: "default",
        control,
    });

    function onValid(values: UpdateAppHealthCheckStatusFormOutput) {
        if (readOnly) {
            return;
        }

        void onSubmit(values);
    }

    function onInvalid(_errors: FieldErrors<UpdateAppHealthCheckStatusFormOutput>) {
        console.error(_errors);
    }

    return (
        <form
            onSubmit={event => {
                event.preventDefault();
                if (readOnly) {
                    return;
                }

                void handleSubmit(onValid, onInvalid)(event);
            }}
        >
            <fieldset
                disabled={readOnly}
                className="contents"
            >
                <FieldGroup>
                    <Field>
                        <InfoBlock
                            title="Status"
                            titleWidth={170}
                        >
                            <Tabs
                                value={status.value}
                                onValueChange={value => {
                                    status.onChange(value);
                                }}
                            >
                                <TabsList>
                                    {Object.entries(statusMap).map(([value, label]) => (
                                        <TabsTrigger
                                            key={value}
                                            value={value}
                                            className="flex-1"
                                        >
                                            {label}
                                        </TabsTrigger>
                                    ))}
                                </TabsList>
                            </Tabs>
                        </InfoBlock>
                    </Field>

                    <Field>
                        <InfoBlock
                            title="Access Expiration"
                            titleWidth={170}
                        >
                            <DateTimePicker
                                value={expireAt.value ?? undefined}
                                onChange={date => {
                                    expireAt.onChange(date ?? null);
                                }}
                                displayFormat={{ hour24: "yyyy-MM-dd HH:mm:ss" }}
                                granularity="second"
                                showClearButton
                                aria-invalid={isExpireAtInvalid}
                                disabled={readOnly}
                            />
                            <FieldError errors={[errors.expireAt]} />
                        </InfoBlock>
                    </Field>

                    <Field>
                        <InfoBlock
                            title="Default"
                            titleWidth={170}
                        >
                            <Checkbox
                                checked={defaultField.value}
                                onCheckedChange={checked => {
                                    defaultField.onChange(checked === true);
                                }}
                            />
                        </InfoBlock>
                    </Field>

                    <Field>
                        <div className="flex justify-end">
                            <Button
                                type="submit"
                                isLoading={isPending}
                                disabled={readOnly}
                                className="min-w-[140px]"
                            >
                                Save
                            </Button>
                        </div>
                    </Field>
                </FieldGroup>
            </fieldset>
        </form>
    );
}

interface Props {
    isPending: boolean;
    onSubmit: (values: UpdateAppHealthCheckStatusFormOutput) => Promise<void> | void;
    initialValues?: UpdateAppHealthCheckStatusFormInput;
    onHasChanges?: (dirty: boolean) => void;
    readOnly?: boolean;
}
