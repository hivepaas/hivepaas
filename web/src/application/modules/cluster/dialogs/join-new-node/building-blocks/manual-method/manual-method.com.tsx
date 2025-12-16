import React from "react";

import { Copy } from "lucide-react";
import { useFormContext, useWatch } from "react-hook-form";
import { toast } from "sonner";
import { NodesCommands } from "~/cluster/data/commands/nodes";

import { Button } from "@/components/ui/button";
import { Field } from "@/components/ui/field";

import { type JoinNewNodeFormInput, type JoinNewNodeFormOutput } from "../../schemas";

function View({ onComplete }: Props) {
    const { control } = useFormContext<JoinNewNodeFormInput, unknown, JoinNewNodeFormOutput>();
    const joinAsManager = useWatch({ control, name: "joinAsManager" });

    const {
        mutate: getJoinCommand,
        data: commandData,
        isPending: isGettingCommand,
    } = NodesCommands.useGetJoinNode({
        onSuccess: () => {
            // Command will be available in commandData
            onComplete();
        },
    });

    const command = commandData?.data.command ?? null;

    function handleGetCommand() {
        getJoinCommand({ joinAsManager });
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
        <Field>
            <div className="border border-dashed border-primary dark:border-primary rounded-lg p-4 bg-gray-50/50 dark:bg-gray-950/20">
                <div className="flex items-center justify-between gap-4">
                    <div className="text-sm text-foreground text-center flex-1">
                        {command ? (
                            <p className="break-all font-mono text-left">{command}</p>
                        ) : (
                            <>
                                Press the below button to get the join command and run it in the server you want to join
                                it to the cluster
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
            <div className="flex justify-end mt-2">
                <Button
                    type="button"
                    onClick={handleGetCommand}
                    isLoading={isGettingCommand}
                    disabled={isGettingCommand}
                >
                    Get Join Command
                </Button>
            </div>
        </Field>
    );
}

interface Props {
    onComplete: () => void;
}

export const ManualMethod = React.memo(View);
