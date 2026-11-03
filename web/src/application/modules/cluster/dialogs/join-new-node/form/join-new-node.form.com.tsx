import React from "react";

import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { Copy } from "lucide-react";
import { toast } from "sonner";
import { NodesCommands } from "~/cluster/data/commands/nodes";

import { MODULE_IDS } from "@application/shared/constants";
import { PermissionTooltipAction, useConditionalModule } from "@application/shared/permissions";

import { Button } from "@/components/ui/button";
import { DialogActionFooter, DialogBody } from "@/components/ui/dialog";
import { Field } from "@/components/ui/field";

export function JoinNewNodeForm({ readOnly = false }: Props) {
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.Cluster });
    const isReadOnly = readOnly || !canWrite;

    const { mutate: getJoinCommand, data: commandData, isPending: isGettingCommand } = NodesCommands.useGetJoinNode({});

    const command = commandData?.data.command ?? null;

    function handleGetCommand() {
        if (isReadOnly) {
            return;
        }

        getJoinCommand({ joinAsManager: false });
    }

    function handleCopyCommand() {
        if (!command) {
            return;
        }

        void navigator.clipboard
            .writeText(command)
            .then(() => {
                toast.success("Command copied to clipboard");
            })
            .catch(() => {
                toast.error("Failed to copy command");
            });
    }

    return (
        <>
            <DialogBody>
                <Field>
                    <div className={cn(dashedBorderBox)}>
                        <div className="flex items-center justify-between gap-4">
                            <div className="text-sm text-foreground text-center flex-1">
                                {command ? (
                                    <p className="break-all font-mono text-left">{command}</p>
                                ) : (
                                    <>
                                        Click the button below to get the command, then run it on the server you want to
                                        join the system
                                    </>
                                )}
                            </div>
                            {command && (
                                <Button
                                    type="button"
                                    variant="link"
                                    size="icon"
                                    className="shrink-0"
                                    onClick={handleCopyCommand}
                                >
                                    <Copy className="h-4 w-4 text-muted-foreground" />
                                </Button>
                            )}
                        </div>
                    </div>
                </Field>
            </DialogBody>
            <DialogActionFooter className="flex justify-end">
                <PermissionTooltipAction
                    id={MODULE_IDS.Cluster}
                    action="write"
                >
                    {({ isDenied }) => (
                        <Button
                            type="button"
                            onClick={handleGetCommand}
                            isLoading={isGettingCommand}
                            disabled={isDenied || readOnly || isGettingCommand}
                        >
                            Get Join Command
                        </Button>
                    )}
                </PermissionTooltipAction>
            </DialogActionFooter>
        </>
    );
}

interface Props {
    readOnly?: boolean;
}
