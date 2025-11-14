import React from "react";

import { EyeIcon } from "lucide-react";

import { AppLink } from "@application/shared/components/navigation";
import { ROUTE } from "@application/shared/constants";

function View({ id }: Props) {
    return (
        <AppLink.Modules
            className="flex items-center justify-center text-link hover:opacity-50 transition-opacity duration-200"
            to={ROUTE.modules.userManagement.users.details.$route(id)}
        >
            <EyeIcon className="size-5" />
        </AppLink.Modules>
    );
}

interface Props {
    id: string;
}

export const ActionsCell = React.memo(View);
