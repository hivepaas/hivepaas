import React, { useState } from "react";

import { Field, FieldError, FieldGroup } from "@components/ui";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { InputNumber } from "@components/ui/input-number";
import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { useController, useFormContext } from "react-hook-form";
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

function View({ domainIndex }: ContainerPortProps) {
    const { id: projectId, appId } = useParams<{ id: string; appId: string }>();
    invariant(projectId, "projectId must be defined");
    invariant(appId, "appId must be defined");

    const { control } = useFormContext<
        AppConfigHttpSettingsFormSchemaInput,
        unknown,
        AppConfigHttpSettingsFormSchemaOutput
    >();

    const {
        field: containerPort,
        fieldState: { error: containerPortError, invalid: isContainerPortInvalid },
    } = useController({ control, name: `domains.${domainIndex}.containerPort` });

    const [modalOpen, setModalOpen] = useState(false);
    const [result, setResult] = useState<CheckPortResult | null>(null);

    const { mutate: checkPort, isPending } = AppContainerSettingsCommands.useCheckPort({
        onSuccess: data => {
            setResult({ port: containerPort.value, open: data.data.open });
            setModalOpen(true);
        },
    });

    function handleCheckPort() {
        const port = containerPort.value;
        if (!port || !projectId || !appId) return;

        checkPort({
            projectID: projectId,
            appID: appId,
            payload: { port, timeout: CHECK_PORT_TIMEOUT },
        });
    }

    return (
        <>
            <InfoBlock
                title={
                    <LabelWithInfo
                        label="Container Port"
                        content="The port on which the container will listen for incoming traffic."
                    />
                }
            >
                <FieldGroup>
                    <Field>
                        <div className="flex items-center gap-2">
                            <InputNumber
                                name={containerPort.name}
                                ref={containerPort.ref}
                                onBlur={containerPort.onBlur}
                                disabled={containerPort.disabled}
                                value={containerPort.value}
                                onValueChange={v => {
                                    containerPort.onChange(v ?? 0);
                                }}
                                useGrouping={false}
                                placeholder="80"
                                className="max-w-[100px]"
                                aria-invalid={isContainerPortInvalid}
                            />
                            <button
                                type="button"
                                className="text-blue-500 cursor-pointer hover:underline select-none disabled:opacity-50 disabled:cursor-not-allowed"
                                onClick={handleCheckPort}
                                disabled={isPending}
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
                <DialogContent className="min-w-[390px] w-[410px]">
                    <DialogHeader>
                        <DialogTitle>Port checking result</DialogTitle>
                    </DialogHeader>

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
                                <div className={cn(dashedBorderBox, "text-center")}>
                                    <span className="text-orange-500">Important:</span> You might need to save your
                                    settings before performing this action
                                </div>
                            )}
                        </div>
                    )}
                </DialogContent>
            </Dialog>
        </>
    );
}

interface ContainerPortProps {
    domainIndex: number;
}

export const ContainerPort = React.memo(View);
