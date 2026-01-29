import React from "react";

import { Tabs, TabsList, TabsTrigger } from "@components/ui/tabs";
import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldErrors, useController, useForm, useFormState } from "react-hook-form";
import { useUpdateEffect } from "react-use";

import { InfoBlock } from "@application/shared/components";
import { EProfileApiKeyStatus } from "@application/shared/enums";

import { Button } from "@/components/ui/button";
import { DateTimePicker } from "@/components/ui/date-time-picker";
import { Field, FieldError, FieldGroup } from "@/components/ui/field";

import {
    type UpdateApiKeyStatusFormInput,
    type UpdateApiKeyStatusFormOutput,
    UpdateApiKeyStatusFormSchema,
} from "../schemas";

const statusMap = {
    [EProfileApiKeyStatus.Active]: "Active",
    [EProfileApiKeyStatus.Disabled]: "Disabled",
} as const;

export function UpdateApiKeyStatusForm({ isPending, onSubmit, initialValues, onHasChanges }: Props) {
    const {
        handleSubmit,
        control,
        formState: { errors },
    } = useForm<UpdateApiKeyStatusFormInput, unknown, UpdateApiKeyStatusFormOutput>({
        defaultValues: {
            status:
                initialValues?.status === EProfileApiKeyStatus.Disabled
                    ? EProfileApiKeyStatus.Disabled
                    : EProfileApiKeyStatus.Active,
            expireAt: initialValues?.expireAt ?? undefined,
        },
        resolver: zodResolver(UpdateApiKeyStatusFormSchema),
        mode: "onSubmit",
    });

    const { isDirty } = useFormState({ control });

    useUpdateEffect(() => {
        onHasChanges?.(isDirty);
    }, [isDirty]);

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

    function onValid(values: UpdateApiKeyStatusFormOutput) {
        void onSubmit(values);
    }

    function onInvalid(_errors: FieldErrors<UpdateApiKeyStatusFormOutput>) {
        // Optional: log errors or show notification
    }

    return (
        <div className="flex flex-col gap-6">
            <form
                onSubmit={event => {
                    event.preventDefault();
                    void handleSubmit(onValid, onInvalid)(event);
                }}
            >
                <FieldGroup>
                    <Field>
                        {/* <FieldLabel htmlFor="status">Status</FieldLabel>
                        <div className="flex gap-2">
                            <Button
                                type="button"
                                variant={status.value === EProfileApiKeyStatus.Active ? "default" : "outline"}
                                onClick={() => {
                                    status.onChange(EProfileApiKeyStatus.Active);
                                }}
                                className="flex-1"
                            >
                                Active
                            </Button>
                            <Button
                                type="button"
                                variant={status.value === EProfileApiKeyStatus.Disabled ? "default" : "outline"}
                                onClick={() => {
                                    status.onChange(EProfileApiKeyStatus.Disabled);
                                }}
                                className="flex-1"
                            >
                                Disabled
                            </Button>
                        </div>
                        <FieldError errors={[errors.status]} /> */}
                        <InfoBlock
                            title="Status"
                            titleWidth={120}
                        >
                            <Tabs
                                value={status.value}
                                onValueChange={value => {
                                    status.onChange(value as EProfileApiKeyStatus);
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
                            titleWidth={120}
                        >
                            <DateTimePicker
                                value={expireAt.value}
                                onChange={expireAt.onChange}
                                displayFormat={{ hour24: "yyyy-MM-dd HH:mm:ss" }}
                                granularity="second"
                                showClearButton
                                fromDate={new Date()}
                                aria-invalid={isExpireAtInvalid}
                            />
                            <FieldError errors={[errors.expireAt]} />
                        </InfoBlock>
                    </Field>

                    <Field>
                        <Button
                            type="submit"
                            isLoading={isPending}
                            className="w-full"
                        >
                            Save
                        </Button>
                    </Field>
                </FieldGroup>
            </form>
        </div>
    );
}

interface Props {
    isPending: boolean;
    onSubmit: (values: UpdateApiKeyStatusFormOutput) => Promise<void> | void;
    initialValues?: UpdateApiKeyStatusFormInput;
    onHasChanges?: (dirty: boolean) => void;
}
