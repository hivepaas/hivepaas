import { cn } from "@/lib/utils";
import { AlertTriangle } from "lucide-react";

interface AvailableInAppsWarningProps {
    className?: string;
    message?: string;
}

export function AvailableInAppsWarning({
    className,
    message = "Warning: Apps will not be able to access this configuration.",
}: AvailableInAppsWarningProps) {
    return (
        <div
            role="status"
            className={cn(
                "flex items-center gap-1.5 text-xs text-amber-600 dark:text-amber-500 font-normal",
                className,
            )}
        >
            <AlertTriangle className="size-3.5 shrink-0 text-amber-600 dark:text-amber-500" />
            <span>{message}</span>
        </div>
    );
}
