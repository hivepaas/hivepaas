import { type PropsWithChildren } from "react";

import { Checkbox, FieldError, Input } from "@components/ui";
import { zodResolver } from "@hookform/resolvers/zod";
import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { type FieldErrors, FormProvider, useController, useForm } from "react-hook-form";
import {
    CLUSTER_NETWORK_FORM_COMPOUND_CONTROL_MAX_WIDTH_CLASS,
    CLUSTER_NETWORK_FORM_CONTROL_MAX_WIDTH_CLASS,
} from "~/cluster/module-shared/constants/network-form-layout.constants";

import { InfoBlock } from "@application/shared/components";
import { KeyValueList } from "@application/shared/form";

import { type CreateNetworkFormInput, type CreateNetworkFormOutput, CreateNetworkFormSchema } from "../schemas";

export function CreateNetworkForm({
    readOnly = false,
    isPending = false,
    showAvailableInProjects = true,
    showProjectNamePrefixNote = false,
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
                className="min-h-0 flex flex-1 flex-col"
            >
                <div>
                    {showProjectNamePrefixNote ? (
                        <div className={cn(dashedBorderBox, "mb-6 text-center text-sm leading-6")}>
                            <span className="text-orange-500">Note:</span> The name of the network in the project will
                            be prefixed with the name of the project.
                        </div>
                    ) : null}
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
                                placeholder="my network"
                                className={CLUSTER_NETWORK_FORM_CONTROL_MAX_WIDTH_CLASS}
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
                                className={CLUSTER_NETWORK_FORM_COMPOUND_CONTROL_MAX_WIDTH_CLASS}
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
                                className={CLUSTER_NETWORK_FORM_COMPOUND_CONTROL_MAX_WIDTH_CLASS}
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
                </div>
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
    showProjectNamePrefixNote?: boolean;
    onSubmit: (values: CreateNetworkFormOutput) => void;
}
