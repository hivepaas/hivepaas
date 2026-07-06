import { type PropsWithChildren, type ReactNode } from "react";

import { Checkbox, Input } from "@components/ui";
import { Tabs, TabsList, TabsTrigger } from "@components/ui/tabs";
import { type FieldErrors, FormProvider, useController, useForm } from "react-hook-form";
import type { ClusterNetwork } from "~/cluster/domain";
import {
    CLUSTER_NETWORK_FORM_COMPOUND_CONTROL_MAX_WIDTH_CLASS,
    CLUSTER_NETWORK_FORM_CONTROL_MAX_WIDTH_CLASS,
} from "~/cluster/module-shared/constants/network-form-layout.constants";
import { EClusterNetworkDriver } from "~/cluster/module-shared/enums";
import { InheritedSettingReadonlyNotice, PermissionReadonlyNotice } from "~/settings/module-shared/components";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";
import { KeyValueList } from "@application/shared/form";

interface ViewNetworkFormValues {
    labels: { key: string; value: string }[];
    options: { key: string; value: string }[];
    availableInProjects: boolean;
    default: boolean;
}

export type ViewNetworkFormOutput = ViewNetworkFormValues;

function toKeyValueList(record: Record<string, string>) {
    return Object.entries(record).map(([key, value]) => ({ key, value }));
}

export function ViewNetworkForm({
    network,
    readOnlyAvailableInProjects = true,
    readOnlyDefault = true,
    readOnlyInherited = false,
    readOnlyPermission = false,
    isPending = false,
    showAvailableInProjects = true,
    onSubmit,
    children,
}: Props) {
    const methods = useForm<ViewNetworkFormValues>({
        defaultValues: {
            labels: toKeyValueList(network.labels),
            options: toKeyValueList(network.options),
            availableInProjects: network.availableInProjects,
            default: network.default,
        },
    });
    const { control, handleSubmit } = methods;
    const { field: availableInProjects } = useController({ control, name: "availableInProjects" });
    const { field: defaultField } = useController({ control, name: "default" });

    const isKnownDriver =
        network.driver === EClusterNetworkDriver.Overlay || network.driver === EClusterNetworkDriver.Bridge;
    const coreDisabled = true;
    const availableInProjectsDisabled =
        readOnlyAvailableInProjects || readOnlyInherited || readOnlyPermission || isPending;
    const defaultDisabled = readOnlyDefault || readOnlyInherited || readOnlyPermission || isPending;

    function onValid(values: ViewNetworkFormOutput) {
        if (readOnlyInherited || readOnlyPermission || isPending) {
            return;
        }

        onSubmit(values);
    }

    function onInvalid(_errors: FieldErrors<ViewNetworkFormOutput>) {
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
                            titleWidth={190}
                        >
                            <Input
                                value={network.name}
                                readOnly
                                className={CLUSTER_NETWORK_FORM_CONTROL_MAX_WIDTH_CLASS}
                            />
                        </InfoBlock>
                        <InfoBlock
                            title="Driver"
                            titleWidth={190}
                        >
                            {isKnownDriver ? (
                                <Tabs
                                    value={network.driver}
                                    className="w-fit"
                                >
                                    <TabsList className="bg-zinc-100/80 p-1 rounded-lg">
                                        <TabsTrigger
                                            value={EClusterNetworkDriver.Overlay}
                                            disabled
                                        >
                                            Overlay
                                        </TabsTrigger>
                                        <TabsTrigger
                                            value={EClusterNetworkDriver.Bridge}
                                            disabled
                                        >
                                            Bridge
                                        </TabsTrigger>
                                    </TabsList>
                                </Tabs>
                            ) : (
                                <Input
                                    value={network.driver}
                                    readOnly
                                    className={CLUSTER_NETWORK_FORM_CONTROL_MAX_WIDTH_CLASS}
                                />
                            )}
                        </InfoBlock>
                        <ReadonlyCheckbox
                            title="Enable IPv4"
                            checked={network.enableIPv4}
                        />
                        <ReadonlyCheckbox
                            title="Enable IPv6"
                            checked={network.enableIPv6}
                        />
                        <ReadonlyCheckbox
                            title="Internal"
                            checked={network.internal}
                        />
                        <ReadonlyCheckbox
                            title="Ingress"
                            checked={network.ingress}
                        />
                        <ReadonlyCheckbox
                            title="Attachable"
                            checked={network.attachable}
                        />
                        <InfoBlock
                            title="Labels"
                            titleWidth={190}
                        >
                            <KeyValueList<ViewNetworkFormValues>
                                name="labels"
                                keyPlaceholder="name"
                                valuePlaceholder="value"
                                className={CLUSTER_NETWORK_FORM_COMPOUND_CONTROL_MAX_WIDTH_CLASS}
                                disabled
                            />
                        </InfoBlock>
                        <InfoBlock
                            title="Options"
                            titleWidth={190}
                        >
                            <KeyValueList<ViewNetworkFormValues>
                                name="options"
                                keyPlaceholder="name"
                                valuePlaceholder="value"
                                className={CLUSTER_NETWORK_FORM_COMPOUND_CONTROL_MAX_WIDTH_CLASS}
                                disabled
                            />
                        </InfoBlock>
                    </fieldset>

                    {showAvailableInProjects ? (
                        <fieldset
                            disabled={availableInProjectsDisabled}
                            className="border-0 p-0 m-0 min-w-0"
                        >
                            <CheckboxField
                                title={<LabelWithInfo label="Available in Projects" />}
                                checked={availableInProjects.value}
                                onCheckedChange={checked => {
                                    availableInProjects.onChange(checked);
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

function ReadonlyCheckbox({ title, checked }: ReadonlyCheckboxProps) {
    return (
        <InfoBlock
            title={title}
            titleWidth={190}
        >
            <Checkbox
                checked={checked}
                disabled
            />
        </InfoBlock>
    );
}

interface CheckboxFieldProps {
    title: string | ReactNode;
    checked: boolean;
    onCheckedChange: (checked: boolean) => void;
}

interface ReadonlyCheckboxProps {
    title: string;
    checked: boolean;
}

interface Props extends PropsWithChildren {
    network: ClusterNetwork;
    readOnlyAvailableInProjects?: boolean;
    readOnlyDefault?: boolean;
    readOnlyInherited?: boolean;
    readOnlyPermission?: boolean;
    isPending?: boolean;
    showAvailableInProjects?: boolean;
    onSubmit: (values: ViewNetworkFormOutput) => void;
}
