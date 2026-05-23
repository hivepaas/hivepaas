import { memo } from "react";

import { Button } from "@components/ui/button";
import { EyeIcon } from "lucide-react";
import { useCreateOrEditRegistryAuthDialog } from "~/settings/dialogs/create-or-edit-registry-auth";
import { isInheritedProjectSetting, useInheritedSettingAlert } from "~/settings/module-shared/hooks";

import type { RegistryAuthTableScope } from "../../registry-auth-table.types";

function View({ scope, id, inherited }: Props) {
    const createOrEditDialog = useCreateOrEditRegistryAuthDialog();
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
            <span className="sr-only">Edit registry auth</span>
        </Button>
    );
}

interface Props {
    scope: RegistryAuthTableScope;
    id: string;
    inherited?: boolean;
}

export const RegistryAuthEditCell = memo(View);
