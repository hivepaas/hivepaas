import React, { useState } from "react";

import { Field, FieldError, FieldGroup, FieldLabel } from "@components/ui";
import { Checkbox } from "@components/ui/checkbox";
import { Dialog, DialogBody, DialogFixedContent, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { InputNumber } from "@components/ui/input-number";
import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { useController, useFormContext, useWatch } from "react-hook-form";
import { useParams } from "react-router";
import invariant from "tiny-invariant";
import { AppContainerSettingsCommands } from "~/projects/data/commands";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";

import { type AppConfigHttpSettingsFormSchemaInput, type AppConfigHttpSettingsFormSchemaOutput } from "../../schemas";

const CHECK_PORT_TIMEOUT = "5s";

interface CheckPortResult {
    port: number;
    open: boolean;
}

function View({ domainIndex, readOnly = false }: ContainerPortProps) {
    const { id: projectId, env, appId } = useParams<{ id: string; env: string; appId: string }>();
    invariant(projectId, "projectId must be defined");
    invariant(env, "env must be defined");
    invariant(appId, "appId must be defined");

    const { control, getValues } = useFormContext<
        AppConfigHttpSettingsFormSchemaInput,
        unknown,
        AppConfigHttpSettingsFormSchemaOutput
    >();

    const outerPort = useWatch({ control, name: "port" });

    const {
        field: containerPort,
        fieldState: { error: containerPortError, invalid: isContainerPortInvalid },
    } = useController({ control, name: `domains.${domainIndex}.containerPort` });

    const { field: overridePort } = useController({ control, name: `domains.${domainIndex}.overridePort` });

    const isOverride = Boolean(overridePort.value);
    const effectivePort = isOverride ? containerPort.value : outerPort;

    const [modalOpen, setModalOpen] = useState(false);
    const [result, setResult] = useState<CheckPortResult | null>(null);

    const { mutate: checkPort, isPending } = AppContainerSettingsCommands.useCheckPort({
        onSuccess: data => {
            setResult({ port: effectivePort, open: data.data.open });
            setModalOpen(true);
        },
    });

    function handleCheckPort() {
        if (readOnly) {
            return;
        }

        const port = effectivePort;
        if (!port || !projectId || !env || !appId) return;

        checkPort({
            projectID: projectId,
            env,
            appID: appId,
            payload: { port, timeout: CHECK_PORT_TIMEOUT },
        });
    }

    return (
        <>
            <InfoBlock
                titleWidth={240}
                title={
                    <LabelWithInfo
                        label="Container Port"
                        content="The port on which the container will listen for incoming traffic."
                    />
                }
            >
                <FieldGroup>
                    <Field>
                        <div className="flex items-center gap-3">
                            <InputNumber
                                name={containerPort.name}
                                ref={containerPort.ref}
                                onBlur={containerPort.onBlur}
                                disabled={readOnly || !isOverride || containerPort.disabled}
                                value={effectivePort}
                                onValueChange={v => {
                                    if (readOnly || !isOverride) {
                                        return;
                                    }

                                    containerPort.onChange(v ?? 0);
                                }}
                                useGrouping={false}
                                placeholder="80"
                                className="max-w-[100px]"
                                aria-invalid={isContainerPortInvalid}
                            />

                            <div className="flex items-center gap-1.5">
                                <Checkbox
                                    id={`domains.${domainIndex}.overridePort`}
                                    checked={isOverride}
                                    disabled={readOnly}
                                    onCheckedChange={checked => {
                                        if (readOnly) return;
                                        const nextOverride = Boolean(checked);
                                        overridePort.onChange(nextOverride);
                                        if (!nextOverride) {
                                            const currentOuterPort = getValues("port");
                                            containerPort.onChange(currentOuterPort);
                                        }
                                    }}
                                />
                                <FieldLabel
                                    htmlFor={`domains.${domainIndex}.overridePort`}
                                    className="cursor-pointer font-normal text-muted-foreground select-none"
                                >
                                    Override port
                                </FieldLabel>
                            </div>

                            <button
                                type="button"
                                className="text-blue-500 cursor-pointer hover:underline select-none disabled:opacity-50 disabled:cursor-not-allowed"
                                onClick={handleCheckPort}
                                disabled={readOnly || isPending}
                                aria-label="Check port availability"
                            >
                                {isPending ? "Checking..." : "Check port"}
                            </button>
                        </div>
                        <FieldError errors={[containerPortError]} />
                    </Field>
                </FieldGroup>
            </InfoBlock>

            <Dialog
                open={modalOpen}
                onOpenChange={setModalOpen}
            >
                <DialogFixedContent className="sm:max-w-[410px]">
                    <DialogHeader>
                        <DialogTitle>Port checking result</DialogTitle>
                    </DialogHeader>

                    <DialogBody>
                        {result && (
                            <div className="flex flex-col gap-4 pt-2 ">
                                <div className="grid grid-cols-2 gap-y-3 text-sm">
                                    <span className="font-medium">Port</span>
                                    <span>{result.port}</span>

                                    <span className="font-medium">Status</span>
                                    <span>
                                        {result.open ? (
                                            <span className="text-green-600">Open</span>
                                        ) : (
                                            <span className="text-red-500">Closed</span>
                                        )}
                                    </span>
                                </div>

                                {!result.open && (
                                    <div className={cn(dashedBorderBox, "text-sm leading-6")}>
                                        <span className="font-semibold text-orange-500">Important:</span> You might need
                                        to save your settings before performing this action
                                    </div>
                                )}
                            </div>
                        )}
                    </DialogBody>
                </DialogFixedContent>
            </Dialog>
        </>
    );
}

interface ContainerPortProps {
    domainIndex: number;
    readOnly?: boolean;
}

export const ContainerPort = React.memo(View);
