import { useEffect } from "react";

import { PasswordInput } from "@components/ui/input-password";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@components/ui/select";
import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldErrors, useController, useForm } from "react-hook-form";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";
import { EAccessTokenKind } from "@application/shared/enums";

import { Button, Checkbox, Field, FieldError, FieldGroup, Input } from "@/components/ui";
import { DateTimePicker } from "@/components/ui/date-time-picker";

import type { CreateOrEditAccessTokenFormInput, CreateOrEditAccessTokenFormOutput } from "../schemas";
import { CreateOrEditAccessTokenFormSchema } from "../schemas";

const kindOptions = Object.values(EAccessTokenKind);

export function CreateOrEditAccessTokenForm({
    isPending,
    isTesting,
    testStatus,
    onSubmit,
    onTestConnection,
    onHasChanges,
    initialValues,
    showAvailableInProjects,
}: Props) {
    const {
        handleSubmit,
        control,
        formState: { errors, isDirty },
    } = useForm<CreateOrEditAccessTokenFormInput, unknown, CreateOrEditAccessTokenFormOutput>({
        defaultValues: {
            name: initialValues?.name ?? "",
            kind: initialValues?.kind ?? EAccessTokenKind.Github,
            user: initialValues?.user ?? "",
            token: initialValues?.token ?? "",
            baseURL: initialValues?.baseURL ?? "",
            expireAt: initialValues?.expireAt ?? null,
            availableInProjects: initialValues?.availableInProjects ?? false,
            default: initialValues?.default ?? false,
        },
        resolver: zodResolver(CreateOrEditAccessTokenFormSchema),
        mode: "onSubmit",
    });

    useEffect(() => {
        onHasChanges?.(isDirty);
    }, [isDirty, onHasChanges]);

    const {
        field: name,
        fieldState: { invalid: isNameInvalid },
    } = useController({ name: "name", control });
    const {
        field: kind,
        fieldState: { invalid: isKindInvalid },
    } = useController({ name: "kind", control });
    const {
        field: user,
        fieldState: { invalid: isUserInvalid },
    } = useController({ name: "user", control });
    const {
        field: token,
        fieldState: { invalid: isTokenInvalid },
    } = useController({ name: "token", control });
    const {
        field: baseURL,
        fieldState: { invalid: isBaseURLInvalid },
    } = useController({ name: "baseURL", control });
    const {
        field: expireAt,
        fieldState: { invalid: isExpireAtInvalid },
    } = useController({ name: "expireAt", control });
    const { field: availableInProjects } = useController({ name: "availableInProjects", control });
    const { field: defaultField } = useController({ name: "default", control });

    function onValid(values: CreateOrEditAccessTokenFormOutput) {
        onSubmit(values);
    }

    function onTestValid(values: CreateOrEditAccessTokenFormOutput) {
        onTestConnection(values);
    }

    function onInvalid(_errors: FieldErrors<CreateOrEditAccessTokenFormOutput>) {
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
            <InfoBlock
                titleWidth={220}
                title={<LabelWithInfo label="Name" />}
            >
                <FieldGroup>
                    <Field>
                        <Input
                            {...name}
                            aria-invalid={isNameInvalid}
                        />
                        <FieldError errors={[errors.name]} />
                    </Field>
                </FieldGroup>
            </InfoBlock>

            <InfoBlock
                titleWidth={220}
                title={
                    <LabelWithInfo
                        label="Type"
                        isRequired
                    />
                }
            >
                <FieldGroup>
                    <Field>
                        <Select
                            value={kind.value}
                            onValueChange={kind.onChange}
                        >
                            <SelectTrigger aria-invalid={isKindInvalid}>
                                <SelectValue placeholder="Select type" />
                            </SelectTrigger>
                            <SelectContent>
                                {kindOptions.map(option => (
                                    <SelectItem
                                        key={option}
                                        value={option}
                                    >
                                        {option}
                                    </SelectItem>
                                ))}
                            </SelectContent>
                        </Select>
                        <FieldError errors={[errors.kind]} />
                    </Field>
                </FieldGroup>
            </InfoBlock>

            <InfoBlock
                titleWidth={220}
                title={
                    <LabelWithInfo
                        label="User"
                        isRequired
                    />
                }
            >
                <FieldGroup>
                    <Field>
                        <Input
                            {...user}
                            aria-invalid={isUserInvalid}
                        />
                        <FieldError errors={[errors.user]} />
                    </Field>
                </FieldGroup>
            </InfoBlock>

            <InfoBlock
                titleWidth={220}
                title={
                    <LabelWithInfo
                        label="Token"
                        isRequired
                    />
                }
            >
                <FieldGroup>
                    <Field>
                        <PasswordInput
                            value={token.value}
                            onChange={token.onChange}
                            aria-invalid={isTokenInvalid}
                        />
                        <FieldError errors={[errors.token]} />
                    </Field>
                </FieldGroup>
            </InfoBlock>

            <InfoBlock
                titleWidth={220}
                title={
                    <LabelWithInfo
                        label="Base URL"
                        isRequired
                    />
                }
            >
                <FieldGroup>
                    <Field>
                        <Input
                            {...baseURL}
                            aria-invalid={isBaseURLInvalid}
                        />
                        <FieldError errors={[errors.baseURL]} />
                    </Field>
                </FieldGroup>
            </InfoBlock>

            <InfoBlock
                titleWidth={220}
                title={<LabelWithInfo label="Access expiration" />}
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
                            Test Access
                        </Button>
                        {testStatus === "succeeded" && <span className="text-sm text-green-600">Succeeded</span>}
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
        </form>
    );
}

interface Props {
    isPending: boolean;
    isTesting: boolean;
    testStatus: "idle" | "succeeded" | "failed";
    onSubmit: (values: CreateOrEditAccessTokenFormOutput) => void;
    onTestConnection: (values: CreateOrEditAccessTokenFormOutput) => void;
    onHasChanges?: (dirty: boolean) => void;
    initialValues?: Partial<CreateOrEditAccessTokenFormInput>;
    showAvailableInProjects: boolean;
}
