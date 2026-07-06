import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldErrors, useController, useForm, useFormState } from "react-hook-form";
import { useUpdateEffect } from "react-use";
import { InheritedSettingReadonlyNotice, PermissionReadonlyNotice } from "~/settings/module-shared/components";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";
import { ESettingStatus } from "@application/shared/enums";

import {
    Button,
    Checkbox,
    DialogActionFooter,
    DialogBody,
    Field,
    FieldError,
    FieldGroup,
    Tabs,
    TabsList,
    TabsTrigger,
} from "@/components/ui";
import { DateTimePicker } from "@/components/ui/date-time-picker";

import {
    type UpdateVolumeStatusFormInput,
    type UpdateVolumeStatusFormOutput,
    UpdateVolumeStatusFormSchema,
} from "../schemas";

const statusMap = {
    [ESettingStatus.Active]: "Active",
    [ESettingStatus.Disabled]: "Disabled",
} as const;

export function UpdateVolumeStatusForm({
    isPending,
    onSubmit,
    initialValues,
    onHasChanges,
    showAvailableInProjects,
    readOnlyInherited = false,
    readOnly = false,
    onClose,
}: Props) {
    const isReadOnly = readOnlyInherited || readOnly;

    const {
        handleSubmit,
        control,
        formState: { errors },
    } = useForm<UpdateVolumeStatusFormInput, unknown, UpdateVolumeStatusFormOutput>({
        defaultValues: {
            status: initialValues?.status === ESettingStatus.Disabled ? ESettingStatus.Disabled : ESettingStatus.Active,
            expireAt: initialValues?.expireAt ?? undefined,
            availableInProjects: initialValues?.availableInProjects ?? false,
            default: initialValues?.default ?? false,
        },
        resolver: zodResolver(UpdateVolumeStatusFormSchema),
        mode: "onSubmit",
    });

    const { isDirty } = useFormState({ control });

    useUpdateEffect(() => {
        onHasChanges?.(isReadOnly ? false : isDirty);
    }, [isDirty]);

    const { field: status } = useController({ name: "status", control });
    const {
        field: expireAt,
        fieldState: { invalid: isExpireAtInvalid },
    } = useController({ name: "expireAt", control });
    const { field: availableInProjects } = useController({ name: "availableInProjects", control });
    const { field: defaultField } = useController({ name: "default", control });

    function onValid(values: UpdateVolumeStatusFormOutput) {
        if (isReadOnly) {
            return;
        }

        onSubmit(values);
    }

    function onInvalid(_errors: FieldErrors<UpdateVolumeStatusFormOutput>) {
        console.error(_errors);
    }

    return (
        <form
            onSubmit={event => {
                event.preventDefault();
                void handleSubmit(onValid, onInvalid)(event);
            }}
            className="min-h-0 flex flex-1 flex-col"
        >
            <DialogBody className="flex flex-col gap-4">
                {readOnlyInherited && <InheritedSettingReadonlyNotice />}
                {readOnly && !readOnlyInherited && <PermissionReadonlyNotice />}
                <fieldset
                    disabled={isReadOnly}
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
                    </FieldGroup>
                </fieldset>
            </DialogBody>
            {!isReadOnly && (
                <DialogActionFooter>
                    <div className="flex justify-end">
                        <Button
                            type="submit"
                            isLoading={isPending}
                            className="min-w-[100px]"
                        >
                            Save
                        </Button>
                    </div>
                </DialogActionFooter>
            )}
            {isReadOnly && (
                <DialogActionFooter>
                    <div className="flex justify-end">
                        <Button
                            type="button"
                            onClick={onClose}
                        >
                            Close
                        </Button>
                    </div>
                </DialogActionFooter>
            )}
        </form>
    );
}

interface Props {
    isPending: boolean;
    onSubmit: (values: UpdateVolumeStatusFormOutput) => void;
    initialValues?: Partial<UpdateVolumeStatusFormInput>;
    onHasChanges?: (dirty: boolean) => void;
    showAvailableInProjects: boolean;
    readOnlyInherited?: boolean;
    readOnly?: boolean;
    onClose?: () => void;
}
