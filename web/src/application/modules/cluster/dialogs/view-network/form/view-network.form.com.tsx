import { Checkbox, Input } from "@components/ui";
import { Tabs, TabsList, TabsTrigger } from "@components/ui/tabs";
import { FormProvider, useForm } from "react-hook-form";
import type { ClusterNetwork } from "~/cluster/domain";
import {
    CLUSTER_NETWORK_FORM_COMPOUND_CONTROL_MAX_WIDTH_CLASS,
    CLUSTER_NETWORK_FORM_CONTROL_MAX_WIDTH_CLASS,
} from "~/cluster/module-shared/constants/network-form-layout.constants";
import { EClusterNetworkDriver } from "~/cluster/module-shared/enums";

import { InfoBlock } from "@application/shared/components";
import { KeyValueList } from "@application/shared/form";

interface ViewNetworkFormValues {
    labels: { key: string; value: string }[];
    options: { key: string; value: string }[];
}

function toKeyValueList(record: Record<string, string>) {
    return Object.entries(record).map(([key, value]) => ({ key, value }));
}

export function ViewNetworkForm({ network }: Props) {
    const methods = useForm<ViewNetworkFormValues>({
        defaultValues: {
            labels: toKeyValueList(network.labels),
            options: toKeyValueList(network.options),
        },
    });

    const isKnownDriver =
        network.driver === EClusterNetworkDriver.Overlay || network.driver === EClusterNetworkDriver.Bridge;

    return (
        <FormProvider {...methods}>
            <div className="flex flex-col gap-6">
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
                <ReadonlyCheckbox
                    title="Available in Projects"
                    checked={network.availableInProjects}
                />
            </div>
        </FormProvider>
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

interface ReadonlyCheckboxProps {
    title: string;
    checked: boolean;
}

interface Props {
    network: ClusterNetwork;
}
