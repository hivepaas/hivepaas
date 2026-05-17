import { memo } from "react";

import { Button } from "@components/ui/button";
import { EyeIcon } from "lucide-react";
import { useCreateOrEditNotificationTargetDialog } from "~/settings/dialogs/create-or-edit-notification-target";

import type { NotificationTargetTableScope } from "../notification-target-table.types";

function View({ scope, id }: Props) {
    const createOrEditDialog = useCreateOrEditNotificationTargetDialog();

    return (
        <Button
            variant="ghost"
            size="icon"
            className="h-8 w-8 text-link hover:opacity-50"
            onClick={() => {
                createOrEditDialog.actions.openEdit(scope, id);
            }}
        >
            <EyeIcon className="size-5" />
            <span className="sr-only">Edit notification target</span>
        </Button>
    );
}

interface Props {
    scope: NotificationTargetTableScope;
    id: string;
}

export const NotificationTargetEditCell = memo(View);
