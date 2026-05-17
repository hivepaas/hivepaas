import { useEffect } from "react";

import { PasswordInput } from "@components/ui/input-password";
import { Textarea } from "@components/ui/textarea";
import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldErrors, FormProvider, useController, useForm } from "react-hook-form";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";
import { SingleValueList } from "@application/shared/form";

import { Button, Checkbox, Field, FieldError, FieldGroup, Input } from "@/components/ui";

import type { CreateOrEditSSHKeyFormInput, CreateOrEditSSHKeyFormOutput } from "../schemas";
import { CreateOrEditSSHKeyFormSchema } from "../schemas";

export function CreateOrEditSSHKeyForm({
    isPending,
    onSubmit,
    onHasChanges,
    initialValues,
    showAvailableInProjects,
}: Props) {
    const form = useForm<CreateOrEditSSHKeyFormInput, unknown, CreateOrEditSSHKeyFormOutput>({
        defaultValues: {
            name: initialValues?.name ?? "",
            privateKey: initialValues?.privateKey ?? "",
            passphrase: initialValues?.passphrase ?? "",
            targets: initialValues?.targets ?? [],
            availableInProjects: initialValues?.availableInProjects ?? false,
            default: initialValues?.default ?? false,
        },
        resolver: zodResolver(CreateOrEditSSHKeyFormSchema),
        mode: "onSubmit",
    });

    const {
        handleSubmit,
        control,
        formState: { errors, isDirty },
    } = form;

    useEffect(() => {
        onHasChanges?.(isDirty);
    }, [isDirty, onHasChanges]);

    const {
        field: name,
        fieldState: { invalid: isNameInvalid },
    } = useController({ name: "name", control });
    const {
        field: privateKey,
        fieldState: { invalid: isPrivateKeyInvalid },
    } = useController({ name: "privateKey", control });
    const {
        field: passphrase,
        fieldState: { invalid: isPassphraseInvalid },
    } = useController({ name: "passphrase", control });
    const { field: availableInProjects } = useController({ name: "availableInProjects", control });
    const { field: defaultField } = useController({ name: "default", control });

    function onValid(values: CreateOrEditSSHKeyFormOutput) {
        onSubmit(values);
    }

    function onInvalid(_errors: FieldErrors<CreateOrEditSSHKeyFormOutput>) {
        console.error(_errors);
    }

    return (
        <FormProvider {...form}>
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
                            label="Private Key"
                            isRequired
                        />
                    }
                >
                    <FieldGroup>
                        <Field>
                            <Textarea
                                {...privateKey}
                                className="min-h-24"
                                aria-invalid={isPrivateKeyInvalid}
                            />
                            <FieldError errors={[errors.privateKey]} />
                        </Field>
                    </FieldGroup>
                </InfoBlock>

                <InfoBlock
                    titleWidth={220}
                    title={<LabelWithInfo label="Passphrase" />}
                >
                    <FieldGroup>
                        <Field>
                            <PasswordInput
                                value={passphrase.value}
                                onChange={passphrase.onChange}
                                aria-invalid={isPassphraseInvalid}
                            />
                            <FieldError errors={[errors.passphrase]} />
                        </Field>
                    </FieldGroup>
                </InfoBlock>

                <InfoBlock
                    titleWidth={220}
                    title={<LabelWithInfo label="Targets" />}
                >
                    <SingleValueList<CreateOrEditSSHKeyFormInput>
                        name="targets"
                        label="Name"
                        placeholder="name"
                    />
                    <FieldError errors={[errors.targets]} />
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
                    <div className="flex justify-end">
                        <Button
                            type="submit"
                            isLoading={isPending}
                        >
                            Save
                        </Button>
                    </div>
                </Field>
            </form>
        </FormProvider>
    );
}

interface Props {
    isPending: boolean;
    onSubmit: (values: CreateOrEditSSHKeyFormOutput) => void;
    onHasChanges?: (dirty: boolean) => void;
    initialValues?: Partial<CreateOrEditSSHKeyFormInput>;
    showAvailableInProjects: boolean;
}
