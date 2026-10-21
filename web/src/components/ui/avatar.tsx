import * as React from "react";

import { cn } from "@/lib/utils";

import { avatarAbbreviation } from "./avatar-abbreviation";

function getColorFromName(name: string) {
    let hash = 0;
    for (let i = 0; i < name.length; i++) {
        hash = name.charCodeAt(i) + ((hash << 5) - hash);
    }
    const hue = Math.abs(hash) % 360;
    // Use HSL to avoid external deps; visually similar to chroma.hsl(hue, 0.65, 0.55)
    return `hsl(${hue} 65% 55%)`;
}

export interface AvatarProps extends React.HTMLAttributes<HTMLDivElement> {
    name: string | null;
    src?: string | null;
    borderless?: boolean;
}

export function Avatar({ name, borderless = false, className, style, src, ...props }: AvatarProps) {
    const [imageError, setImageError] = React.useState(false);
    const safeSource = src === "" ? null : src;
    const showInitials = !safeSource || imageError;

    return (
        <div
            data-slot="avatar"
            className={cn(
                "relative flex size-8 shrink-0 overflow-hidden rounded-full items-center justify-center",
                className,
            )}
            style={{
                backgroundColor: getColorFromName(name ?? ""),
                border: borderless ? "none" : undefined,
                ...style,
            }}
            {...props}
        >
            {!showInitials ? (
                <img
                    data-slot="avatar-image"
                    src={safeSource ?? undefined}
                    alt={name ?? ""}
                    onError={() => setImageError(true)}
                    className={cn("aspect-square size-full object-cover")}
                />
            ) : name ? (
                <span className="select-none font-medium text-white">{avatarAbbreviation(name ?? "")}</span>
            ) : null}
        </div>
    );
}
