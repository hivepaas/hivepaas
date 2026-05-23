import { memo } from "react";

import { Button } from "@components/ui/button";
import { EyeIcon } from "lucide-react";
import { useCreateOrEditEmailAccountDialog } from "~/settings/dialogs/create-or-edit-email-account";
import { isInheritedProjectSetting, useInheritedSettingAlert } from "~/settings/module-shared/hooks";

import type { EmailAccountTableScope } from "../../email-account-table.types";

function View({ scope, id, inherited }: Props) {
    const createOrEditDialog = useCreateOrEditEmailAccountDialog();
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
            <span className="sr-only">Edit email account</span>
        </Button>
    );
}

interface Props {
    scope: EmailAccountTableScope;
    id: string;
    inherited?: boolean;
}

export const EmailAccountEditCell = memo(View);
