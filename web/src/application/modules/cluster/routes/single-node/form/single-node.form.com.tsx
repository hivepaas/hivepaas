import * as React from "react";
import { type PropsWithChildren, useImperativeHandle } from "react";

import { FieldError, Input } from "@components/ui";
import { Tabs, TabsList, TabsTrigger } from "@components/ui/tabs";
import { zodResolver } from "@hookform/resolvers/zod";
import { FormProvider, useController, useForm } from "react-hook-form";
import { type NodeDetails } from "~/cluster/domain";
import { ENodeAvailability, ENodeRole } from "~/cluster/module-shared/enums";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";
import { KeyValueList } from "@application/shared/form";

import { NodeStatusBadge } from "@application/modules/cluster/module-shared/components";

import { SingleNodeFormSchema, type SingleNodeFormSchemaInput, type SingleNodeFormSchemaOutput } from "../schemas";
import { type SingleNodeFormRef } from "../types";

export function SingleNodeForm({ ref, defaultValues, onSubmit, children }: Props) {
    const methods = useForm<SingleNodeFormSchemaInput, unknown, SingleNodeFormSchemaOutput>({
        defaultValues: {
            name: defaultValues.name,
            role: defaultValues.role,
            availability: defaultValues.availability,
            labels: Object.entries(defaultValues.labels).map(([key, value]) => ({ key, value })),
        },
        resolver: zodResolver(SingleNodeFormSchema),
        mode: "onSubmit",
    });

    const {
        control,
        formState: { errors },
    } = methods;

    useImperativeHandle(
        ref,
        () => ({
            setValues: (values: Partial<SingleNodeFormSchemaInput>) => {
                methods.reset({
                    ...methods.getValues(),
                    ...values,
                });
            },
            onError: () => {
                // Implementation for error handling if needed
            },
        }),
        [methods],
    );

    const {
        field: name,
        fieldState: { invalid: isNameInvalid },
    } = useController({
        control,
        name: "name",
    });

    return (
        <div className="pt-2">
            <FormProvider {...methods}>
                <form
                    onSubmit={event => {
                        event.preventDefault();

                        void methods.handleSubmit(onSubmit)(event);
                    }}
                    className="flex flex-col gap-6"
                >
                    {/* Name */}
                    <InfoBlock title="Name">
                        <Input
                            {...name}
                            value={name.value}
                            onChange={name.onChange}
                            type="text"
                            className="max-w-[400px]"
                            placeholder=""
                            aria-invalid={isNameInvalid}
                        />
                        <FieldError errors={[errors.name]} />
                    </InfoBlock>

                    {/* Resources */}
                    <InfoBlock title="Resources">
                        <span className="text-sm font-normal">
                            {defaultValues.resources?.cpus} CPU, {defaultValues.resources?.memoryMB} MB RAM
                        </span>
                    </InfoBlock>

                    {/* Platform */}
                    <InfoBlock title="Platform">
                        <span className="text-sm font-normal">
                            {defaultValues.platform?.os} {defaultValues.platform?.architecture}
                        </span>
                    </InfoBlock>

                    {/* Docker Engine Version */}
                    <InfoBlock title="Docker Engine Version">
                        <span className="text-sm font-normal">28.0.0</span>
                    </InfoBlock>

                    {/* Node State */}
                    <InfoBlock title="Node State">
                        <NodeStatusBadge status={defaultValues.status} />
                    </InfoBlock>

                    {/* Role */}
                    <InfoBlock title="Role">
                        <Tabs
                            value={methods.watch("role")}
                            onValueChange={v => {
                                methods.setValue("role", v as ENodeRole);
                            }}
                            className="w-fit"
                        >
                            <TabsList className="bg-zinc-100/80 p-1 rounded-lg">
                                <TabsTrigger
                                    disabled={defaultValues.isLeader}
                                    value={ENodeRole.Manager}
                                >
                                    Manager
                                </TabsTrigger>
                                <TabsTrigger
                                    disabled={defaultValues.isLeader}
                                    value={ENodeRole.Worker}
                                >
                                    Worker
                                </TabsTrigger>
                            </TabsList>
                        </Tabs>
                    </InfoBlock>

                    <div className="h-px bg-zinc-100" />

                    {/* Availability */}
                    <InfoBlock title="Availability">
                        <Tabs
                            value={methods.watch("availability")}
                            onValueChange={v => {
                                methods.setValue("availability", v as ENodeAvailability);
                            }}
                            className="w-fit"
                        >
                            <TabsList className="bg-zinc-100/80 p-1 rounded-lg">
                                <TabsTrigger
                                    disabled={defaultValues.isLeader}
                                    value={ENodeAvailability.Active}
                                >
                                    Active
                                </TabsTrigger>
                                <TabsTrigger
                                    disabled={defaultValues.isLeader}
                                    value={ENodeAvailability.Pause}
                                >
                                    Pause
                                </TabsTrigger>
                                <TabsTrigger
                                    disabled={defaultValues.isLeader}
                                    value={ENodeAvailability.Drain}
                                >
                                    Drain
                                </TabsTrigger>
                            </TabsList>
                        </Tabs>
                    </InfoBlock>

                    <div className="h-px bg-zinc-100" />

                    {/* Labels */}
                    <InfoBlock
                        title={
                            <LabelWithInfo
                                label="Labels"
                                content="Labels description!!"
                            />
                        }
                    >
                        <KeyValueList<SingleNodeFormSchemaInput>
                            name="labels"
                            keyLabel="Label"
                            className="max-w-[660px]"
                            checkDuplicates
                        />
                    </InfoBlock>
                    {children}
                </form>
            </FormProvider>
        </div>
    );
}

type Props = PropsWithChildren<{
    ref?: React.Ref<SingleNodeFormRef>;
    defaultValues: NodeDetails;
    onSubmit: (values: SingleNodeFormSchemaOutput) => void;
}>;
