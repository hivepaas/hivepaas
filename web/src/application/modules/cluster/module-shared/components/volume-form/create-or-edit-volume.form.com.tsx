import { type PropsWithChildren, type ReactNode } from "react";

import { zodResolver } from "@hookform/resolvers/zod";
import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { type FieldErrors, FormProvider, useController, useForm, useWatch } from "react-hook-form";
import { NodesQueries } from "~/cluster/data/queries";
import {
    EClusterVolumeDriverMode,
    EClusterVolumeLocalType,
    EClusterVolumePropagation,
} from "~/cluster/module-shared/enums";
import { InheritedSettingReadonlyNotice, PermissionReadonlyNotice } from "~/settings/module-shared/components";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";
import { KeyValueList } from "@application/shared/form";

import {
    Checkbox,
    FieldError,
    FieldGroup,
    Input,
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
    Tabs,
    TabsList,
    TabsTrigger,
} from "@/components/ui";

import {
    type CreateOrEditVolumeFormInput,
    type CreateOrEditVolumeFormOutput,
    CreateOrEditVolumeFormSchema,
} from "./create-or-edit-volume.form.schema";
import { DEFAULT_VOLUME_FORM_VALUES } from "./volume-form.constants";

const LOCAL_VOLUME_NODE_NOTE =
    "A local bind volume is physically stored on a specific node in your cluster. Selecting a target Node or specifying a Node Label ensures the volume directory is created on that node, and any service mounting this volume will be automatically pinned to run on that specific node.";

const propagationOptions = [
    { value: EClusterVolumePropagation.Default, label: "default" },
    { value: EClusterVolumePropagation.RPrivate, label: "rprivate" },
    { value: EClusterVolumePropagation.Private, label: "private" },
    { value: EClusterVolumePropagation.RShared, label: "rshared" },
    { value: EClusterVolumePropagation.Shared, label: "shared" },
    { value: EClusterVolumePropagation.RSlave, label: "rslave" },
    { value: EClusterVolumePropagation.Slave, label: "slave" },
] as const;

export function CreateOrEditVolumeForm({
    initialValues,
    readOnlyCore = false,
    readOnlyAvailableInProjects = false,
    readOnlyDefault = false,
    readOnlyInherited = false,
    readOnlyPermission = false,
    isPending = false,
    showAvailableInProjects = true,
    onSubmit,
    children,
}: Props) {
    const methods = useForm<CreateOrEditVolumeFormInput, unknown, CreateOrEditVolumeFormOutput>({
        defaultValues: {
            ...DEFAULT_VOLUME_FORM_VALUES,
            ...initialValues,
            bindOptions: {
                ...DEFAULT_VOLUME_FORM_VALUES.bindOptions,
                ...initialValues?.bindOptions,
            },
            nfsOptions: {
                ...DEFAULT_VOLUME_FORM_VALUES.nfsOptions,
                ...initialValues?.nfsOptions,
            },
            tmpfsOptions: {
                ...DEFAULT_VOLUME_FORM_VALUES.tmpfsOptions,
                ...initialValues?.tmpfsOptions,
            },
        },
        resolver: zodResolver(CreateOrEditVolumeFormSchema),
        mode: "onSubmit",
    });

    const {
        control,
        handleSubmit,
        formState: { errors },
    } = methods;

    const { field: name } = useController({ control, name: "name" });
    const { field: driverMode } = useController({ control, name: "driverMode" });
    const { field: customDriverName } = useController({ control, name: "customDriverName" });
    const { field: localType } = useController({ control, name: "localType" });
    const { field: bindDirectory } = useController({ control, name: "bindOptions.directory" });
    const { field: bindNodeId } = useController({ control, name: "bindOptions.nodeId" });
    const { field: bindNodeLabel } = useController({ control, name: "bindOptions.nodeLabel" });
    const { field: bindPropagation } = useController({ control, name: "bindOptions.propagation" });
    const { field: bindReadonly } = useController({ control, name: "bindOptions.readonly" });
    const { field: bindExtraOptions } = useController({ control, name: "bindOptions.extraOptions" });
    const { field: nfsAddr } = useController({ control, name: "nfsOptions.addr" });
    const { field: nfsDevice } = useController({ control, name: "nfsOptions.device" });
    const { field: nfsVersion } = useController({ control, name: "nfsOptions.version" });
    const { field: nfsReadonly } = useController({ control, name: "nfsOptions.readonly" });
    const { field: nfsExtraOptions } = useController({ control, name: "nfsOptions.extraOptions" });
    const { field: tmpfsDevice } = useController({ control, name: "tmpfsOptions.device" });
    const { field: tmpfsSize } = useController({ control, name: "tmpfsOptions.size" });
    const { field: tmpfsUid } = useController({ control, name: "tmpfsOptions.uid" });
    const { field: tmpfsExtraOptions } = useController({ control, name: "tmpfsOptions.extraOptions" });
    const { field: inheritable } = useController({ control, name: "inheritable" });
    const { field: defaultField } = useController({ control, name: "default" });

    const nodesQuery = NodesQueries.useFindManyPaginated({
        pagination: {
            page: 1,
            size: 100,
        },
    });
    const nodes = nodesQuery.data?.data ?? [];

    const currentDriverMode = useWatch({ control, name: "driverMode" });
    const currentLocalType = useWatch({ control, name: "localType" });
    const coreDisabled = readOnlyCore || readOnlyInherited || readOnlyPermission || isPending;
    const inheritableDisabled = readOnlyAvailableInProjects || readOnlyInherited || readOnlyPermission || isPending;
    const defaultDisabled = readOnlyDefault || readOnlyInherited || readOnlyPermission || isPending;

    function onValid(values: CreateOrEditVolumeFormOutput) {
        if (readOnlyInherited || readOnlyPermission || isPending) {
            return;
        }

        onSubmit(values);
    }

    function onInvalid(_errors: FieldErrors<CreateOrEditVolumeFormInput>) {
        console.error("Invalid", _errors);
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
                <div className="flex flex-col gap-6">
                    {readOnlyInherited && <InheritedSettingReadonlyNotice />}
                    {readOnlyPermission && !readOnlyInherited && <PermissionReadonlyNotice />}

                    <fieldset
                        disabled={coreDisabled}
                        className="flex flex-col gap-6 border-0 p-0 m-0 min-w-0"
                    >
                        <InfoBlock
                            title="Name"
                            titleWidth={220}
                        >
                            <Input
                                {...name}
                                value={name.value}
                                placeholder="my-volume"
                                className="max-w-[600px]"
                                aria-invalid={Boolean(errors.name)}
                            />
                            <FieldError errors={[errors.name]} />
                        </InfoBlock>

                        <InfoBlock
                            title="Driver"
                            titleWidth={220}
                        >
                            <Tabs
                                value={driverMode.value}
                                onValueChange={value => {
                                    driverMode.onChange(value as EClusterVolumeDriverMode);
                                }}
                            >
                                <TabsList>
                                    <TabsTrigger value={EClusterVolumeDriverMode.Local}>local</TabsTrigger>
                                    <TabsTrigger value={EClusterVolumeDriverMode.Custom}>custom</TabsTrigger>
                                </TabsList>
                            </Tabs>
                        </InfoBlock>

                        {currentDriverMode === EClusterVolumeDriverMode.Custom && (
                            <InfoBlock
                                title="Driver Name"
                                titleWidth={220}
                            >
                                <Input
                                    {...customDriverName}
                                    value={customDriverName.value}
                                    placeholder="driver"
                                    className="max-w-[600px]"
                                    aria-invalid={Boolean(errors.customDriverName)}
                                />
                                <FieldError errors={[errors.customDriverName]} />
                            </InfoBlock>
                        )}

                        {currentDriverMode === EClusterVolumeDriverMode.Local && (
                            <>
                                <InfoBlock
                                    title="Volume Type"
                                    titleWidth={220}
                                >
                                    <Tabs
                                        value={localType.value}
                                        onValueChange={value => {
                                            localType.onChange(value as EClusterVolumeLocalType);
                                        }}
                                    >
                                        <TabsList>
                                            <TabsTrigger value={EClusterVolumeLocalType.Bind}>bind</TabsTrigger>
                                            <TabsTrigger value={EClusterVolumeLocalType.Nfs}>nfs</TabsTrigger>
                                            <TabsTrigger value={EClusterVolumeLocalType.Tmpfs}>tmpfs</TabsTrigger>
                                        </TabsList>
                                    </Tabs>
                                </InfoBlock>

                                {currentLocalType === EClusterVolumeLocalType.Bind && (
                                    <FieldGroup>
                                        <InfoBlock
                                            title="Directory"
                                            titleWidth={220}
                                        >
                                            <Input
                                                {...bindDirectory}
                                                value={bindDirectory.value}
                                                placeholder="auto"
                                                className="max-w-[600px]"
                                            />
                                        </InfoBlock>

                                        <div className={cn(dashedBorderBox, "text-sm leading-6")}>
                                            <span className="font-semibold text-orange-500">Note:</span>{" "}
                                            {LOCAL_VOLUME_NODE_NOTE}
                                        </div>

                                        <InfoBlock
                                            title="Node"
                                            titleWidth={220}
                                        >
                                            <Select
                                                value={bindNodeId.value || "__none__"}
                                                onValueChange={value => {
                                                    bindNodeId.onChange(value === "__none__" ? "" : value);
                                                }}
                                            >
                                                <SelectTrigger className="max-w-[600px]">
                                                    <SelectValue placeholder="Select node (optional)" />
                                                </SelectTrigger>
                                                <SelectContent>
                                                    <SelectItem value="__none__">
                                                        None (Current Manager Node)
                                                    </SelectItem>
                                                    {nodes.map(node => {
                                                        const nodeName = node.name || node.hostname || node.refId;
                                                        return (
                                                            <SelectItem
                                                                key={node.refId || node.id}
                                                                value={node.refId || node.id}
                                                            >
                                                                {nodeName !== node.refId
                                                                    ? `${nodeName} (${node.refId})`
                                                                    : node.refId}
                                                            </SelectItem>
                                                        );
                                                    })}
                                                </SelectContent>
                                            </Select>
                                        </InfoBlock>
                                        <InfoBlock
                                            title="Node Label"
                                            titleWidth={220}
                                        >
                                            <Input
                                                {...bindNodeLabel}
                                                value={bindNodeLabel.value}
                                                placeholder="key=value or key"
                                                className="max-w-[600px]"
                                            />
                                        </InfoBlock>
                                        <InfoBlock
                                            title="Propagation"
                                            titleWidth={220}
                                        >
                                            <Select
                                                value={bindPropagation.value || "__default__"}
                                                onValueChange={value => {
                                                    bindPropagation.onChange(
                                                        value === "__default__"
                                                            ? EClusterVolumePropagation.Default
                                                            : value,
                                                    );
                                                }}
                                            >
                                                <SelectTrigger className="max-w-[600px]">
                                                    <SelectValue placeholder="select propagation" />
                                                </SelectTrigger>
                                                <SelectContent>
                                                    {propagationOptions.map(option => (
                                                        <SelectItem
                                                            key={option.label}
                                                            value={option.value || "__default__"}
                                                        >
                                                            {option.label}
                                                        </SelectItem>
                                                    ))}
                                                </SelectContent>
                                            </Select>
                                        </InfoBlock>
                                        <CheckboxField
                                            title="Read-only"
                                            checked={bindReadonly.value}
                                            onCheckedChange={checked => {
                                                bindReadonly.onChange(checked);
                                            }}
                                        />
                                        <InfoBlock
                                            title="Extra Options"
                                            titleWidth={220}
                                        >
                                            <Input
                                                {...bindExtraOptions}
                                                value={bindExtraOptions.value}
                                                placeholder="noexec,nosuid,nodev"
                                                className="max-w-[600px]"
                                            />
                                        </InfoBlock>
                                    </FieldGroup>
                                )}

                                {currentLocalType === EClusterVolumeLocalType.Nfs && (
                                    <FieldGroup>
                                        <InfoBlock
                                            title="Address"
                                            titleWidth={220}
                                        >
                                            <Input
                                                {...nfsAddr}
                                                value={nfsAddr.value}
                                                placeholder="1.2.3.4"
                                                className="max-w-[600px]"
                                                aria-invalid={Boolean(errors.nfsOptions?.addr)}
                                            />
                                            <FieldError errors={[errors.nfsOptions?.addr]} />
                                        </InfoBlock>
                                        <InfoBlock
                                            title="Device"
                                            titleWidth={220}
                                        >
                                            <Input
                                                {...nfsDevice}
                                                value={nfsDevice.value}
                                                placeholder=":/mnt/shared/data"
                                                className="max-w-[600px]"
                                                aria-invalid={Boolean(errors.nfsOptions?.device)}
                                            />
                                            <FieldError errors={[errors.nfsOptions?.device]} />
                                        </InfoBlock>
                                        <InfoBlock
                                            title="Version"
                                            titleWidth={220}
                                        >
                                            <Input
                                                {...nfsVersion}
                                                value={nfsVersion.value}
                                                placeholder="4"
                                                className="max-w-[600px]"
                                            />
                                        </InfoBlock>
                                        <CheckboxField
                                            title="Read Only"
                                            checked={nfsReadonly.value}
                                            onCheckedChange={checked => {
                                                nfsReadonly.onChange(checked);
                                            }}
                                        />
                                        <InfoBlock
                                            title="Extra Options"
                                            titleWidth={220}
                                        >
                                            <Input
                                                {...nfsExtraOptions}
                                                value={nfsExtraOptions.value}
                                                placeholder="nolock,soft"
                                                className="max-w-[600px]"
                                            />
                                        </InfoBlock>
                                    </FieldGroup>
                                )}

                                {currentLocalType === EClusterVolumeLocalType.Tmpfs && (
                                    <FieldGroup>
                                        <InfoBlock
                                            title="Device"
                                            titleWidth={220}
                                        >
                                            <Input
                                                {...tmpfsDevice}
                                                value={tmpfsDevice.value}
                                                placeholder="device"
                                                className="max-w-[600px]"
                                            />
                                        </InfoBlock>
                                        <InfoBlock
                                            title="Size"
                                            titleWidth={220}
                                        >
                                            <Input
                                                {...tmpfsSize}
                                                value={tmpfsSize.value}
                                                placeholder="1gb"
                                                className="max-w-[600px]"
                                                aria-invalid={Boolean(errors.tmpfsOptions?.size)}
                                            />
                                            <FieldError errors={[errors.tmpfsOptions?.size]} />
                                        </InfoBlock>
                                        <InfoBlock
                                            title="UID"
                                            titleWidth={220}
                                        >
                                            <Input
                                                {...tmpfsUid}
                                                value={tmpfsUid.value}
                                                placeholder="0"
                                                className="max-w-[600px]"
                                            />
                                        </InfoBlock>
                                        <InfoBlock
                                            title="Extra Options"
                                            titleWidth={220}
                                        >
                                            <Input
                                                {...tmpfsExtraOptions}
                                                value={tmpfsExtraOptions.value}
                                                placeholder="noexec,nosuid,nodev"
                                                className="max-w-[600px]"
                                            />
                                        </InfoBlock>
                                    </FieldGroup>
                                )}
                            </>
                        )}

                        <InfoBlock
                            title="Labels"
                            titleWidth={220}
                        >
                            <KeyValueList<CreateOrEditVolumeFormInput>
                                name="labels"
                                checkDuplicates
                                keyPlaceholder="name"
                                valuePlaceholder="value"
                                className="max-w-[800px]"
                                disabled={coreDisabled}
                                enableValueEditing
                            />
                            <FieldError errors={[errors.labels]} />
                        </InfoBlock>
                        <InfoBlock
                            title="Options"
                            titleWidth={220}
                        >
                            <KeyValueList<CreateOrEditVolumeFormInput>
                                name="options"
                                checkDuplicates
                                keyPlaceholder="name"
                                valuePlaceholder="value"
                                className="max-w-[800px]"
                                disabled={coreDisabled}
                                enableValueEditing
                            />
                            <FieldError errors={[errors.options]} />
                        </InfoBlock>
                    </fieldset>

                    {showAvailableInProjects ? (
                        <fieldset
                            disabled={inheritableDisabled}
                            className="border-0 p-0 m-0 min-w-0"
                        >
                            <CheckboxField
                                title={<LabelWithInfo label="Available in Projects" />}
                                checked={inheritable.value}
                                onCheckedChange={checked => {
                                    inheritable.onChange(checked);
                                }}
                            />
                        </fieldset>
                    ) : null}

                    <fieldset
                        disabled={defaultDisabled}
                        className="border-0 p-0 m-0 min-w-0"
                    >
                        <CheckboxField
                            title={<LabelWithInfo label="Default" />}
                            checked={defaultField.value}
                            onCheckedChange={checked => {
                                defaultField.onChange(checked);
                            }}
                        />
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
            titleWidth={220}
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
    title: string | ReactNode;
    checked: boolean;
    onCheckedChange: (checked: boolean) => void;
}

interface Props extends PropsWithChildren {
    initialValues?: Partial<CreateOrEditVolumeFormInput>;
    readOnlyCore?: boolean;
    readOnlyAvailableInProjects?: boolean;
    readOnlyDefault?: boolean;
    readOnlyInherited?: boolean;
    readOnlyPermission?: boolean;
    isPending?: boolean;
    showAvailableInProjects?: boolean;
    onSubmit: (values: CreateOrEditVolumeFormOutput) => void;
}
