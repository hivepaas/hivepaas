import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldErrors, useController, useForm, useFormState } from "react-hook-form";
import { useUpdateEffect } from "react-use";
import { PermissionReadonlyNotice } from "~/settings/module-shared/components";

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
    type UpdateOAuthStatusFormInput,
    type UpdateOAuthStatusFormOutput,
    UpdateOAuthStatusFormSchema,
} from "../schemas";

const statusMap = {
    [ESettingStatus.Active]: "Active",
    [ESettingStatus.Disabled]: "Disabled",
} as const;

export function UpdateOAuthStatusForm({
    isPending,
    onSubmit,
    initialValues,
    onHasChanges,
    readOnly = false,
    onClose,
}: Props) {
    const isReadOnly = readOnly;

    const {
        handleSubmit,
        control,
        formState: { errors },
    } = useForm<UpdateOAuthStatusFormInput, unknown, UpdateOAuthStatusFormOutput>({
        defaultValues: {
            status: initialValues?.status === ESettingStatus.Disabled ? ESettingStatus.Disabled : ESettingStatus.Active,
            expireAt: initialValues?.expireAt ?? undefined,
            default: initialValues?.default ?? false,
        },
        resolver: zodResolver(UpdateOAuthStatusFormSchema),
        mode: "onSubmit",
    });

    const { isDirty } = useFormState({ control });

    useUpdateEffect(() => {
        onHasChanges?.(isReadOnly ? false : isDirty);
    }, [isDirty, isReadOnly]);

    const { field: status } = useController({ name: "status", control });
    const {
        field: expireAt,
        fieldState: { invalid: isExpireAtInvalid },
    } = useController({ name: "expireAt", control });
    const { field: defaultField } = useController({ name: "default", control });

    function onValid(values: UpdateOAuthStatusFormOutput) {
        if (isReadOnly) {
            return;
        }

        onSubmit(values);
    }

    function onInvalid(_errors: FieldErrors<UpdateOAuthStatusFormOutput>) {
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
                {isReadOnly && <PermissionReadonlyNotice />}
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
                            className="min-w-[100px]"
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
    onSubmit: (values: UpdateOAuthStatusFormOutput) => void;
    initialValues?: Partial<UpdateOAuthStatusFormInput>;
    onHasChanges?: (dirty: boolean) => void;
    readOnly?: boolean;
    onClose?: () => void;
}
