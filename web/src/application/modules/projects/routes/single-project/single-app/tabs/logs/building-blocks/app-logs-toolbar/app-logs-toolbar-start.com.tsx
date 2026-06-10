import { cn } from "@/lib/utils";
import { LoaderCircle } from "lucide-react";

import { Button } from "@/components/ui";

export function AppLogsToolbarStart({
    isConnectionActive,
    isStreaming,
    isRefreshPending,
    onStream,
    onRefresh,
    onStop,
}: AppLogsToolbarStartProps) {
    if (isConnectionActive) {
        return (
            <div className="flex min-w-0 items-center gap-3">
                <span className="flex items-center gap-2 text-sm text-rose-500">
                    <LoaderCircle className={cn("size-4", isStreaming && "animate-spin")} />
                    streaming
                </span>
                <Button
                    type="button"
                    variant="link"
                    className="h-auto p-0 text-sm text-primary"
                    onClick={onStop}
                >
                    Stop
                </Button>
            </div>
        );
    }

    return (
        <div className="flex min-w-0 items-center gap-3">
            <Button
                type="button"
                variant="link"
                className="h-auto p-0 text-sm text-primary"
                onClick={onStream}
            >
                Stream
            </Button>
            <Button
                type="button"
                variant="link"
                className="h-auto p-0 text-sm text-primary"
                isLoading={isRefreshPending}
                onClick={onRefresh}
            >
                Refresh
            </Button>
        </div>
    );
}

interface AppLogsToolbarStartProps {
    isConnectionActive: boolean;
    isStreaming: boolean;
    isRefreshPending: boolean;
    onStream: () => void;
    onRefresh: () => void;
    onStop: () => void;
}
