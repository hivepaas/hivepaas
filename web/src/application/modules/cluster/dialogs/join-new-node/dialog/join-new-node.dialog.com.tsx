import React, { useState } from "react";

import { Button } from "@components/ui";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { EJoinNodeMethod } from "~/cluster/module-shared/enums";

import { JoinNewNodeForm } from "../form";
import { useJoinNewNodeDialogState } from "../hooks";
import { type JoinNewNodeFormOutput } from "../schemas";

export function JoinNewNodeDialog() {
    const [method, setMethod] = useState<EJoinNodeMethod>(EJoinNodeMethod.RunCommandManually);
    const { state, props, ...actions } = useJoinNewNodeDialogState();

    function onSubmit(values: JoinNewNodeFormOutput) {
        // TODO: Implement SSH join node command when API is available
        console.log("Join node via SSH:", values);
        actions.close();
    }

    function handleMethodChange(newMethod: EJoinNodeMethod) {
        setMethod(newMethod);
    }

    const open = state.mode !== "closed";

    return (
        <Dialog
            open={open}
            onOpenChange={actions.close}
        >
            <DialogContent className="min-w-[500px] max-w-[600px]">
                <DialogHeader>
                    <DialogTitle>Join new node to the swarm cluster</DialogTitle>
                </DialogHeader>
                <JoinNewNodeForm
                    onSubmit={onSubmit}
                    onMethodChange={handleMethodChange}
                >
                    {method === EJoinNodeMethod.RunCommandViaSSH && (
                        <div className="flex justify-end mt-4">
                            <Button type="submit">Join Node</Button>
                        </div>
                    )}
                </JoinNewNodeForm>
            </DialogContent>
        </Dialog>
    );
}
