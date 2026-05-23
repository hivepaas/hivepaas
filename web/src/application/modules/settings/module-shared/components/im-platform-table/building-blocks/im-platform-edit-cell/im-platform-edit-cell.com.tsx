import { memo } from "react";

import { Button } from "@components/ui/button";
import { EyeIcon } from "lucide-react";
import { useCreateOrEditImPlatformDialog } from "~/settings/dialogs/create-or-edit-im-platform";
import { isInheritedProjectSetting, useInheritedSettingAlert } from "~/settings/module-shared/hooks";

import type { ImPlatformTableScope } from "../../im-platform-table.types";

function View({ scope, id, inherited }: Props) {
    const createOrEditDialog = useCreateOrEditImPlatformDialog();
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
            <span className="sr-only">Edit IM platform</span>
        </Button>
    );
}

interface Props {
    scope: ImPlatformTableScope;
    id: string;
    inherited?: boolean;
}

export const ImPlatformEditCell = memo(View);
