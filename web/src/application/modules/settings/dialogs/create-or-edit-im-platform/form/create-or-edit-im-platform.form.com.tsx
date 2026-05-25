import { useEffect } from "react";

import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldErrors, useController, useForm, useWatch } from "react-hook-form";
import { InheritedSettingReadonlyNotice } from "~/settings/module-shared/components/inherited-setting-readonly-notice.com";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";
import { EImServiceKind } from "@application/shared/enums";

import {
    Button,
    Checkbox,
    Field,
    FieldError,
    FieldGroup,
    Input,
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui";

import type { CreateOrEditImPlatformFormInput, CreateOrEditImPlatformFormOutput } from "../schemas";
import { CreateOrEditImPlatformFormSchema } from "../schemas";

export function CreateOrEditImPlatformForm({
    isPending,
    isTesting,
    testStatus,
    onSubmit,
    onTestSendMsg,
    onHasChanges,
    initialValues,
    showAvailableInProjects,
    readOnlyInherited = false,
    onClose,
}: Props) {
    const {
        handleSubmit,
        control,
        formState: { errors, isDirty },
    } = useForm<CreateOrEditImPlatformFormInput, unknown, CreateOrEditImPlatformFormOutput>({
        defaultValues: {
            name: initialValues?.name ?? "",
            kind: initialValues?.kind ?? EImServiceKind.Slack,
            webhook: initialValues?.webhook ?? "",
            availableInProjects: initialValues?.availableInProjects ?? false,
            default: initialValues?.default ?? false,
        },
        resolver: zodResolver(CreateOrEditImPlatformFormSchema),
        mode: "onSubmit",
    });

    useEffect(() => {
        onHasChanges?.(readOnlyInherited ? false : isDirty);
    }, [isDirty, onHasChanges, readOnlyInherited]);

    const kindValue = useWatch({ control, name: "kind" });

    const {
        field: name,
        fieldState: { invalid: isNameInvalid },
    } = useController({ name: "name", control });
    const {
        field: kind,
        fieldState: { invalid: isKindInvalid },
    } = useController({ name: "kind", control });
    const {
        field: webhook,
        fieldState: { invalid: isWebhookInvalid },
    } = useController({ name: "webhook", control });
    const { field: availableInProjects } = useController({ name: "availableInProjects", control });
    const { field: defaultField } = useController({ name: "default", control });

    function onValid(values: CreateOrEditImPlatformFormOutput) {
        if (readOnlyInherited) {
            return;
        }

        onSubmit(values);
    }

    function onTestValid(values: CreateOrEditImPlatformFormOutput) {
        onTestSendMsg(values);
    }

    function onInvalid(_errors: FieldErrors<CreateOrEditImPlatformFormOutput>) {
        console.error(_errors);
    }

    return (
        <form
            onSubmit={event => {
                event.preventDefault();
                void handleSubmit(onValid, onInvalid)(event);
            }}
            className="flex flex-col gap-6"
        >
            {readOnlyInherited && <InheritedSettingReadonlyNotice />}
            <fieldset
                disabled={readOnlyInherited}
                className="flex flex-col gap-6 border-0 p-0 m-0 min-w-0"
            >
                <FieldGroup>
                    <InfoBlock
                        titleWidth={220}
                        title={<LabelWithInfo label="Name" />}
                    >
                        <Field>
                            <Input
                                {...name}
                                aria-invalid={isNameInvalid}
                            />
                            <FieldError errors={[errors.name]} />
                        </Field>
                    </InfoBlock>

                    <InfoBlock
                        titleWidth={220}
                        title={<LabelWithInfo label="Type" />}
                    >
                        <Field>
                            <Select
                                value={kind.value}
                                onValueChange={value => {
                                    kind.onChange(value);
                                }}
                            >
                                <SelectTrigger aria-invalid={isKindInvalid}>
                                    <SelectValue placeholder="Select type" />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value={EImServiceKind.Slack}>Slack Webhook</SelectItem>
                                    <SelectItem value={EImServiceKind.Discord}>Discord Webhook</SelectItem>
                                </SelectContent>
                            </Select>
                            <FieldError errors={[errors.kind]} />
                        </Field>
                    </InfoBlock>

                    <InfoBlock
                        titleWidth={220}
                        title={
                            <LabelWithInfo
                                label="Webhook URL"
                                isRequired
                            />
                        }
                    >
                        <Field>
                            <Input
                                {...webhook}
                                aria-invalid={isWebhookInvalid}
                                placeholder={
                                    kindValue === EImServiceKind.Slack ? "Slack webhook URL" : "Discord webhook URL"
                                }
                            />
                            <FieldError errors={[errors.webhook]} />
                        </Field>
                    </InfoBlock>

                    {showAvailableInProjects && (
                        <InfoBlock
                            titleWidth={220}
                            title={<LabelWithInfo label="Available in Projects" />}
                        >
                            <Checkbox
                                checked={availableInProjects.value}
                                onCheckedChange={checked => {
                                    availableInProjects.onChange(Boolean(checked));
                                }}
                            />
                        </InfoBlock>
                    )}

                    <InfoBlock
                        titleWidth={220}
                        title={<LabelWithInfo label="Default" />}
                    >
                        <Checkbox
                            checked={defaultField.value}
                            onCheckedChange={checked => {
                                defaultField.onChange(Boolean(checked));
                            }}
                        />
                    </InfoBlock>
                </FieldGroup>

                {!readOnlyInherited && (
                    <Field>
                        <div className="flex items-center justify-between gap-4">
                            <div className="flex items-center gap-3">
                                <Button
                                    type="button"
                                    variant="secondary"
                                    isLoading={isTesting}
                                    onClick={() => {
                                        void handleSubmit(onTestValid, onInvalid)();
                                    }}
                                >
                                    Test Send Msg
                                </Button>
                                {testStatus === "succeeded" && (
                                    <span className="text-sm text-green-600">Succeeded</span>
                                )}
                                {testStatus === "failed" && <span className="text-sm text-destructive">Failed</span>}
                            </div>
                            <Button
                                type="submit"
                                isLoading={isPending}
                            >
                                Save
                            </Button>
                        </div>
                    </Field>
                )}
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
    isTesting: boolean;
    testStatus: "idle" | "succeeded" | "failed";
    onSubmit: (values: CreateOrEditImPlatformFormOutput) => void;
    onTestSendMsg: (values: CreateOrEditImPlatformFormOutput) => void;
    onHasChanges?: (dirty: boolean) => void;
    initialValues?: Partial<CreateOrEditImPlatformFormInput>;
    showAvailableInProjects: boolean;
    readOnlyInherited?: boolean;
    onClose?: () => void;
}
