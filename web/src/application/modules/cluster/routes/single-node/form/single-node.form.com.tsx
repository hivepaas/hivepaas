import * as React from "react";
import { type PropsWithChildren, useImperativeHandle, useState } from "react";

import { Button, FieldError, Input } from "@components/ui";
import { Tabs, TabsList, TabsTrigger } from "@components/ui/tabs";
import { zodResolver } from "@hookform/resolvers/zod";
import { Plus, Trash2 } from "lucide-react";
import { FormProvider, useController, useFieldArray, useForm } from "react-hook-form";
import { toast } from "sonner";
import { type NodeDetails } from "~/cluster/domain";
import { ENodeAvailability, ENodeRole } from "~/cluster/module-shared/enums";

import { InfoBlock, InputWithAddOn, LabelWithInfo } from "@application/shared/components";

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

    const { fields, append, remove } = useFieldArray({
        control,
        name: "labels",
    });

    const [newLabelKey, setNewLabelKey] = useState("");
    const [newLabelValue, setNewLabelValue] = useState("");

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

    const handleAddLabel = () => {
        if (newLabelKey && newLabelValue) {
            const exists = fields.some(field => field.key === newLabelKey);

            if (exists) {
                toast.error(`Label key "${newLabelKey}" already exists`);
                return;
            }

            append({ key: newLabelKey, value: newLabelValue });
            setNewLabelKey("");
            setNewLabelValue("");
        }
    };

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
                        <div className="flex flex-col gap-4 max-w-[660px]">
                            <div className="flex gap-4">
                                <InputWithAddOn
                                    addonLeft="Label"
                                    value={newLabelKey}
                                    onChange={e => {
                                        setNewLabelKey(e.target.value);
                                    }}
                                // className="bg-white border-zinc-200"
                                />
                                <InputWithAddOn
                                    addonLeft="Value"
                                    value={newLabelValue}
                                    onChange={e => {
                                        setNewLabelValue(e.target.value);
                                    }}
                                // className="bg-white border-zinc-200"
                                />
                                <Button
                                    type="button"
                                    variant="outline"
                                    onClick={handleAddLabel}
                                    disabled={newLabelKey === "" || newLabelValue === ""}
                                    className=" border-zinc-20 h-9 px-4"
                                >
                                    <Plus className="size-4 mr-2" /> Add
                                </Button>
                            </div>

                            <div className="mt-2 divide-y divide-zinc-200">
                                {fields.map((field, index) => (
                                    <div
                                        key={field.id}
                                        className="flex items-center group gap-4 py-2"
                                    >
                                        <div className="grid grid-cols-2 flex-1 gap-4">
                                            <div className="text-sm break-words">{field.key}</div>
                                            <div className="text-sm break-words">{field.value}</div>
                                        </div>
                                        <div className="w-[84px]">

                                            <Button
                                                type="button"
                                                variant="ghost"
                                                size="icon"
                                                onClick={() => {
                                                    remove(index);
                                                }}
                                                className="h-8 w-8 text-zinc-400 hover:text-red-500 hover:bg-red-50 rounded-md"
                                            >
                                                <Trash2 className="size-4" />
                                            </Button>
                                        </div>
                                    </div>
                                ))}
                            </div>
                        </div>
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
