import React from "react";

import { Button } from "@components/ui/button";

import { NoAccessImage } from "@assets/images";

import { ROUTE } from "@application/shared/constants";
import { useProfileContext } from "@application/shared/context";
import { useAppNavigate } from "@application/shared/hooks/router";

function View({ withBackButton = true }: Props) {
    const { profile } = useProfileContext();

    const { navigate } = useAppNavigate();

    return (
        <div className="min-h-svh w-full bg-muted/20 flex items-center justify-center px-6">
            <div className="flex flex-col items-center text-center gap-6">
                <div className="w-full max-w-[640px]">
                    <NoAccessImage className="w-full h-auto" />
                </div>

                <h1 className="text-2xl md:text-3xl font-semibold tracking-tight text-foreground">
                    You don&apos;t have the permission to view this page
                </h1>

                <p className="text-sm text-muted-foreground whitespace-pre-line">
                    Ask for access, or switch to an account with access
                </p>

                {withBackButton && (
                    <div className="pt-2">
                        <Button
                            variant="outline"
                            onClick={() => {
                                if (profile === null) {
                                    navigate(ROUTE.auth.signIn.$route, {
                                        replace: true,
                                    });

                                    return;
                                }

                                navigate("/", {
                                    replace: true,
                                });
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

export const PageNoAccess = React.memo(View);
