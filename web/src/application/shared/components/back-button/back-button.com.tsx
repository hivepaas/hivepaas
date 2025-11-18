import React from "react";

import { cn } from "@/lib/utils";
import { Button } from "@components/ui";
import { CircleArrowLeft } from "lucide-react";
import { matchPath, useLocation } from "react-router";

import { useAppNavigate } from "@application/shared/hooks/router";

function useFrom() {
    const location = useLocation();

    const from = (location.state as LocationState | null)?.from ?? "";

    const matchAuth = matchPath(
        {
            path: "auth/*",
            caseSensitive: false,
            end: false,
        },
        from,
    );

    if (matchAuth !== null) {
        return null;
    }

    return from.trim() || null;
}

interface LocationState {
    from?: string;
}

export function BackButton() {
    const { navigate } = useAppNavigate();

    const from = useFrom();

    const disabled = from === null;

    return (
        <Button
            variant="ghost"
            size="icon"
            className={cn("h-8 w-8 rounded-full p-0", disabled && "opacity-50 cursor-not-allowed")}
            onClick={event => {
                event.preventDefault();

                if (from !== null) {
                    void navigate.basic(from, {
                        ignorePrevPath: true,
                    });
                }
            }}
            disabled={disabled}
        >
            <CircleArrowLeft className="size-4" />
        </Button>
    );
}
