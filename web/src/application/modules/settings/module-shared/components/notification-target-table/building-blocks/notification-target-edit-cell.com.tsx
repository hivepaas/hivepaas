import { memo } from "react";

import { Button } from "@components/ui/button";
import { EyeIcon } from "lucide-react";
import { useCreateOrEditNotificationTargetDialog } from "~/settings/dialogs/create-or-edit-notification-target";
import { isInheritedProjectSetting, useInheritedSettingAlert } from "~/settings/module-shared/hooks";

import type { NotificationTargetTableScope } from "../notification-target-table.types";

function View({ scope, id, inherited }: Props) {
    const createOrEditDialog = useCreateOrEditNotificationTargetDialog();
    const inheritedSettingAlert = useInheritedSettingAlert();

    return (
        <Button
            variant="ghost"
            size="icon"
            className="h-8 w-8 text-link hover:opacity-50"
            onClick={() => {
                if (isInheritedProjectSetting(scope, inherited)) {
                    inheritedSettingAlert.open();
                    return;
                }

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
    inherited?: boolean;
}

export const NotificationTargetEditCell = memo(View);
