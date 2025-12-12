import React from "react";

import { EyeIcon } from "lucide-react";

import { AppLink } from "@application/shared/components/navigation";

function View({ id }: Props) {
    void id;
    return (
        <div className="flex items-center justify-center gap-4">
            <AppLink.Modules
                className="flex items-center justify-center text-link hover:opacity-50 transition-opacity duration-200"
                to="#"
            >
                <EyeIcon className="size-5" />
            </AppLink.Modules>
        </div>
    );
}

interface Props {
    id: string;
}

export const ActionsCell = React.memo(View);
