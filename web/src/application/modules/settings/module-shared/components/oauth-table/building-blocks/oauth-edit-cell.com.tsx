import { memo } from "react";

import { Button } from "@components/ui/button";
import { EyeIcon } from "lucide-react";

import { ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";

function View({ id }: Props) {
    const { navigate } = useAppNavigate();

    return (
        <Button
            variant="ghost"
            size="icon"
            className="h-8 w-8 text-link hover:opacity-50"
            onClick={() => {
                navigate.modules(ROUTE.settings.oauth.edit.$route(id));
            }}
        >
            <EyeIcon className="size-5" />
            <span className="sr-only">Edit OAuth</span>
        </Button>
    );
}

interface Props {
    id: string;
}

export const OAuthEditCell = memo(View);
