import { memo } from "react";

import { Button } from "@components/ui";

import { NotFound404Image } from "@assets/images";

import { useAppNavigate } from "@application/shared/hooks/router";

function View({ withBackButton = true }: Props) {
    const { navigate } = useAppNavigate();

    return (
        <div className="min-h-svh w-full bg-muted/20 flex items-center justify-center px-6">
            <div className="flex flex-col items-center text-center gap-6">
                <div className="w-full max-w-[640px]">
                    <NotFound404Image className="w-full h-auto" />
                </div>

                <h1 className="text-2xl md:text-3xl font-semibold tracking-tight text-foreground">
                    Sorry, the page not found
                </h1>

                <p className="text-sm text-muted-foreground whitespace-pre-line">
                    {"The link you followed probably broken\nor the page has been removed"}
                </p>

                {withBackButton && (
                    <div className="pt-2">
                        <Button
                            variant="outline"
                            onClick={() => {
                                void navigate.basic("/", { replace: true });
                            }}
                        >
                            Back to application
                        </Button>
                    </div>
                )}
            </div>
        </div>
    );
}

interface Props {
    withBackButton?: boolean;
}
export const Page404NotFound = memo(View);
