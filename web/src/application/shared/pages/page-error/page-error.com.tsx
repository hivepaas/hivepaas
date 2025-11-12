import { memo } from "react";

import { Button } from "@components/ui";

import { useAppNavigate } from "@application/shared/hooks/router";

function View({ error, onRetry }: Props) {
    const { navigate } = useAppNavigate();

    return (
        <div className="min-h-svh w-full bg-muted/20 flex items-center justify-center px-6">
            <div className="flex flex-col items-center text-center gap-6 max-w-[640px]">
                <div className="flex flex-col items-center gap-4">
                    <div className="text-6xl font-bold text-destructive">!</div>
                    <h1 className="text-2xl md:text-3xl font-semibold tracking-tight text-foreground">
                        Something went wrong
                    </h1>
                </div>

                <p className="text-sm text-muted-foreground">
                    {error.message || "An unexpected error occurred. Please try again."}
                </p>

                <div className="flex gap-3 pt-2">
                    <Button
                        variant="outline"
                        onClick={() => {
                            navigate("/", { replace: true });
                        }}
                    >
                        Back to application
                    </Button>
                    <Button onClick={onRetry}>Retry</Button>
                </div>
            </div>
        </div>
    );
}

export const PageError = memo(View);

interface Props {
    error: Error;
    onRetry: () => unknown;
}
