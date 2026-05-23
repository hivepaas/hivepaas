import { memo } from "react";

import { Button } from "@components/ui/button";
import { EyeIcon } from "lucide-react";
import { useCreateOrEditSslCertDialog } from "~/settings/dialogs/create-or-edit-ssl-cert";
import { isInheritedProjectSetting, useInheritedSettingAlert } from "~/settings/module-shared/hooks";

import type { SslCertTableScope } from "../../ssl-cert-table.types";

function View({ scope, id, inherited }: Props) {
    const createOrEditDialog = useCreateOrEditSslCertDialog();
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
            <span className="sr-only">Edit SSL certificate</span>
        </Button>
    );
}

interface Props {
    scope: SslCertTableScope;
    id: string;
    inherited?: boolean;
}

export const SslCertEditCell = memo(View);
