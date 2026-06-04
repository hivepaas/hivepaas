import React from "react";

import { Button } from "@components/ui";
import { EyeIcon } from "lucide-react";
import type { ClusterNetwork } from "~/cluster/domain";
import type { NetworkManagementScope } from "~/cluster/module-shared/types";

import { useViewNetworkDialog } from "@application/modules/cluster/dialogs";

function View({ network, scope }: Props) {
    const dialog = useViewNetworkDialog();

    return (
        <div className="flex items-center justify-center gap-4">
            <Button
                type="button"
                variant="ghost"
                size="icon"
                className="text-link hover:opacity-50 transition-opacity duration-200"
                onClick={() => {
                    dialog.actions.open(scope, network);
                }}
            >
                <EyeIcon className="size-5" />
                <span className="sr-only">View network</span>
            </Button>
        </div>
    );
}

interface Props {
    network: ClusterNetwork;
    scope: NetworkManagementScope;
}

export const ActionsCell = React.memo(View);
