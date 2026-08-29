import React from "react";

import { Separator } from "@components/ui";
import { Dialog, DialogFixedContent, DialogHeader, DialogTitle } from "@components/ui/dialog";

import { MODULE_IDS } from "@application/shared/constants";
import { useConditionalModule } from "@application/shared/permissions";

import { JoinNewNodeForm } from "../form";
import { useJoinNewNodeDialogState } from "../hooks";

export function JoinNewNodeDialog() {
    const { state, ...actions } = useJoinNewNodeDialogState();
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.Cluster });

    const open = state.mode !== "closed";

    return (
        <Dialog
            open={open}
            onOpenChange={actions.close}
        >
            <DialogFixedContent className="sm:max-w-[650px]">
                <DialogHeader>
                    <DialogTitle>Join new node to the swarm cluster</DialogTitle>
                </DialogHeader>
                <div className="px-4">
                    <Separator className="opacity-50" />
                </div>
                <JoinNewNodeForm readOnly={!canWrite} />
            </DialogFixedContent>
        </Dialog>
    );
}
