import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldErrors, useController, useForm, useFormState } from "react-hook-form";
import { useUpdateEffect } from "react-use";
import { InheritedSettingReadonlyNotice } from "~/settings/module-shared/components/inherited-setting-readonly-notice.com";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";
import { ESettingStatus } from "@application/shared/enums";

import { Button, Checkbox, Field, FieldError, FieldGroup, Tabs, TabsList, TabsTrigger } from "@/components/ui";
import { DateTimePicker } from "@/components/ui/date-time-picker";

import {
    type UpdateEmailAccountStatusFormInput,
    type UpdateEmailAccountStatusFormOutput,
    UpdateEmailAccountStatusFormSchema,
} from "../schemas";

const statusMap = {
    [ESettingStatus.Active]: "Active",
    [ESettingStatus.Disabled]: "Disabled",
} as const;

export function UpdateEmailAccountStatusForm({
    isPending,
    onSubmit,
    initialValues,
    onHasChanges,
    showAvailableInProjects,
    readOnlyInherited = false,
    onClose,
}: Props) {
    const {
        handleSubmit,
        control,
        formState: { errors },
    } = useForm<UpdateEmailAccountStatusFormInput, unknown, UpdateEmailAccountStatusFormOutput>({
        defaultValues: {
            status: initialValues?.status === ESettingStatus.Disabled ? ESettingStatus.Disabled : ESettingStatus.Active,
            expireAt: initialValues?.expireAt ?? undefined,
            availableInProjects: initialValues?.availableInProjects ?? false,
            default: initialValues?.default ?? false,
        },
        resolver: zodResolver(UpdateEmailAccountStatusFormSchema),
        mode: "onSubmit",
    });

    const { isDirty } = useFormState({ control });

    useUpdateEffect(() => {
        onHasChanges?.(readOnlyInherited ? false : isDirty);
    }, [isDirty]);

    const { field: status } = useController({ name: "status", control });
    const {
        field: expireAt,
        fieldState: { invalid: isExpireAtInvalid },
    } = useController({ name: "expireAt", control });
    const { field: availableInProjects } = useController({ name: "availableInProjects", control });
    const { field: defaultField } = useController({ name: "default", control });

    function onValid(values: UpdateEmailAccountStatusFormOutput) {
        if (readOnlyInherited) {
            return;
        }

        onSubmit(values);
    }

    function onInvalid(_errors: FieldErrors<UpdateEmailAccountStatusFormOutput>) {
        console.error(_errors);
    }

    return (
        <form
            onSubmit={event => {
                event.preventDefault();
                void handleSubmit(onValid, onInvalid)(event);
            }}
        >
            {readOnlyInherited && <InheritedSettingReadonlyNotice />}
            <fieldset
                disabled={readOnlyInherited}
                className="border-0 p-0 m-0 min-w-0"
            >
                <FieldGroup>
                    <Field>
                        <InfoBlock
                            title="Status"
                            titleWidth={160}
                        >
                            <Tabs
                                value={status.value}
                                onValueChange={value => {
                                    status.onChange(value as ESettingStatus);
                                }}
                            >
                                <TabsList>
                                    {Object.entries(statusMap).map(([value, label]) => (
                                        <TabsTrigger
                                            key={value}
                                            value={value}
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
                            title="Access expiration"
                            titleWidth={160}
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
                            />
                            <FieldError errors={[errors.expireAt]} />
                        </InfoBlock>
                    </Field>

                    {showAvailableInProjects && (
                        <Field>
                            <InfoBlock
                                title={<LabelWithInfo label="Available in Projects" />}
                                titleWidth={160}
                            >
                                <Checkbox
                                    checked={availableInProjects.value}
                                    onCheckedChange={checked => {
                                        availableInProjects.onChange(Boolean(checked));
                                    }}
                                />
                            </InfoBlock>
                        </Field>
                    )}

                    <Field>
                        <InfoBlock
                            title={<LabelWithInfo label="Default" />}
                            titleWidth={160}
                        >
                            <Checkbox
                                checked={defaultField.value}
                                onCheckedChange={checked => {
                                    defaultField.onChange(Boolean(checked));
                                }}
                            />
                        </InfoBlock>
                    </Field>

                    {!readOnlyInherited && (
                        <Field>
                            <div className="flex justify-end">
                                <Button
                                    type="submit"
                                    isLoading={isPending}
                                >
                                    Save
                                </Button>
                            </div>
                        </Field>
                    )}
                </FieldGroup>
            </fieldset>
            {readOnlyInherited && (
                <Field>
                    <div className="flex justify-end">
                        <Button
                            type="button"
                            onClick={onClose}
                        >
                            Close
                        </Button>
                    </div>
                </Field>
            )}
        </form>
    );
}

interface Props {
    isPending: boolean;
    onSubmit: (values: UpdateEmailAccountStatusFormOutput) => void;
    initialValues?: Partial<UpdateEmailAccountStatusFormInput>;
    onHasChanges?: (dirty: boolean) => void;
    showAvailableInProjects: boolean;
    readOnlyInherited?: boolean;
    onClose?: () => void;
}
