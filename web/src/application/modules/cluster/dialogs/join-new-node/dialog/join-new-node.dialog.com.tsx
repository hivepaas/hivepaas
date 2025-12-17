import React, { useEffect, useState } from "react";

import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { toast } from "sonner";

import { NodesCommands } from "@application/modules/cluster/data";
import { EJoinNodeMethod } from "@application/modules/cluster/module-shared/enums";

import { JoinNewNodeForm } from "../form";
import { useJoinNewNodeDialogState } from "../hooks";
import { type JoinNewNodeFormOutput } from "../schemas";

export function JoinNewNodeDialog() {
    const { state, props, ...actions } = useJoinNewNodeDialogState();
    const [hasChanges, setHasChanges] = useState(false);

    const open = state.mode !== "closed";

    const { mutate: joinNode } = NodesCommands.useJoinNode({
        onSuccess: () => {
            toast.success("Node joined successfully");
            actions.close();
        },
    });

    // Reset hasChanges when dialog closes
    useEffect(() => {
        if (state.mode === "closed") {
            setHasChanges(false);
        }
    }, [state.mode]);

    function onSubmit(values: JoinNewNodeFormOutput) {
        if (values.method === EJoinNodeMethod.RunCommandViaSSH && values.sshKey) {
            joinNode({
                sshKey: {
                    id: values.sshKey.id,
                },
                host: values.host,
                port: values.port,
                user: values.user,
                joinAsManager: values.joinAsManager,
            });
        }
    }

    function handleClose(): void {
        if (hasChanges) {
            const userConfirmed: boolean = window.confirm("Are you sure you want to close without saving changes?");
            if (!userConfirmed) {
                return;
            }
        }

        setHasChanges(false);
        actions.close();
    }

    return (
        <Dialog
            open={open}
            onOpenChange={handleClose}
        >
            <DialogContent className="min-w-[500px] max-w-[600px]">
                <DialogHeader>
                    <DialogTitle>Join new node to the swarm cluster</DialogTitle>
                </DialogHeader>
                <JoinNewNodeForm
                    onSubmit={onSubmit}
                    onHasChanges={setHasChanges}
                />
            </DialogContent>
        </Dialog>
    );
}
