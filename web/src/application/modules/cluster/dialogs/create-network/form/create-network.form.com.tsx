import { type PropsWithChildren } from "react";

import { Checkbox, FieldError, Input } from "@components/ui";
import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldErrors, FormProvider, useController, useForm } from "react-hook-form";

import { InfoBlock } from "@application/shared/components";
import { KeyValueList } from "@application/shared/form";

import { type CreateNetworkFormInput, type CreateNetworkFormOutput, CreateNetworkFormSchema } from "../schemas";

export function CreateNetworkForm({
    readOnly = false,
    isPending = false,
    showAvailableInProjects = true,
    onSubmit,
    children,
}: Props) {
    const methods = useForm<CreateNetworkFormInput, unknown, CreateNetworkFormOutput>({
        defaultValues: {
            name: "",
            enableIPv4: true,
            enableIPv6: false,
            internal: false,
            attachable: false,
            ingress: false,
            labels: [],
            options: [],
            availableInProjects: false,
        },
        resolver: zodResolver(CreateNetworkFormSchema),
        mode: "onSubmit",
    });

    const {
        control,
        handleSubmit,
        formState: { errors },
    } = methods;

    const { field: name } = useController({ control, name: "name" });
    const { field: enableIPv4 } = useController({ control, name: "enableIPv4" });
    const { field: enableIPv6 } = useController({ control, name: "enableIPv6" });
    const { field: internal } = useController({ control, name: "internal" });
    const { field: attachable } = useController({ control, name: "attachable" });
    const { field: availableInProjects } = useController({ control, name: "availableInProjects" });

    function onValid(values: CreateNetworkFormOutput) {
        if (readOnly) {
            return;
        }

        onSubmit(values);
    }

    function onInvalid(_errors: FieldErrors<CreateNetworkFormInput>) {
        console.log("Invalid", _errors);
    }

    return (
        <FormProvider {...methods}>
            <form
                onSubmit={event => {
                    event.preventDefault();
                    void handleSubmit(onValid, onInvalid)(event);
                }}
                className="flex flex-col gap-6"
            >
                <fieldset
                    disabled={readOnly || isPending}
                    className="flex flex-col gap-6 border-0 p-0 m-0 min-w-0"
                >
                    <InfoBlock
                        title="Name"
                        titleWidth={190}
                    >
                        <Input
                            {...name}
                            value={name.value}
                            placeholder="project_a_net_1"
                            className="max-w-[520px]"
                            aria-invalid={Boolean(errors.name)}
                        />
                        <FieldError errors={[errors.name]} />
                    </InfoBlock>
                    <CheckboxField
                        title="Enable IPv4"
                        checked={enableIPv4.value}
                        onCheckedChange={value => {
                            enableIPv4.onChange(value);
                        }}
                    />
                    <CheckboxField
                        title="Enable IPv6"
                        checked={enableIPv6.value}
                        onCheckedChange={value => {
                            enableIPv6.onChange(value);
                        }}
                    />
                    <CheckboxField
                        title="Internal"
                        checked={internal.value}
                        onCheckedChange={value => {
                            internal.onChange(value);
                        }}
                    />
                    <CheckboxField
                        title="Attachable"
                        checked={attachable.value}
                        onCheckedChange={value => {
                            attachable.onChange(value);
                        }}
                    />
                    <InfoBlock
                        title="Labels"
                        titleWidth={190}
                    >
                        <KeyValueList<CreateNetworkFormInput>
                            name="labels"
                            checkDuplicates
                            keyPlaceholder="name"
                            valuePlaceholder="value"
                            className="max-w-[620px]"
                            disabled={readOnly || isPending}
                        />
                        <FieldError errors={[errors.labels]} />
                    </InfoBlock>
                    <InfoBlock
                        title="Options"
                        titleWidth={190}
                    >
                        <KeyValueList<CreateNetworkFormInput>
                            name="options"
                            checkDuplicates
                            keyPlaceholder="name"
                            valuePlaceholder="value"
                            className="max-w-[620px]"
                            disabled={readOnly || isPending}
                        />
                        <FieldError errors={[errors.options]} />
                    </InfoBlock>
                    {showAvailableInProjects ? (
                        <CheckboxField
                            title="Available in Projects"
                            checked={availableInProjects.value}
                            onCheckedChange={value => {
                                availableInProjects.onChange(value);
                            }}
                        />
                    ) : null}
                </fieldset>
                {children}
            </form>
        </FormProvider>
    );
}

function CheckboxField({ title, checked, onCheckedChange }: CheckboxFieldProps) {
    return (
        <InfoBlock
            title={title}
            titleWidth={190}
        >
            <Checkbox
                checked={checked}
                onCheckedChange={value => {
                    onCheckedChange(Boolean(value));
                }}
            />
        </InfoBlock>
    );
}

interface CheckboxFieldProps {
    title: string;
    checked: boolean;
    onCheckedChange: (checked: boolean) => void;
}

interface Props extends PropsWithChildren {
    readOnly?: boolean;
    isPending?: boolean;
    showAvailableInProjects?: boolean;
    onSubmit: (values: CreateNetworkFormOutput) => void;
}
