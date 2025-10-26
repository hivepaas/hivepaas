import React from "react";

import { Button } from "@components/ui/button";
import { useNavigate } from "react-router";

import { NoAccessImage } from "@assets/images";

import { ROUTE } from "@application/shared/constants";
import { useProfileContext } from "@application/shared/context";

function View({ withBackButton = true }: Props) {
    const { profile } = useProfileContext();

    const navigate = useNavigate();

    return (
        <div className="page-no-access">
            <div className="body">
                <div className="image-wrapper">
                    <NoAccessImage />
                </div>
                <h1 className="title">You don’t have the permission to view this page</h1>
                <div className="description">Ask for access, or switch to an account with access</div>
            </div>

            {withBackButton && (
                <Button
                    onClick={() => {
                        if (profile === null) {
                            void navigate(ROUTE.auth.signIn.$route, {
                                replace: true,
                            });

                            return;
                        }

                        void navigate("/", {
                            replace: true,
                        });
                    }}
                >
                    Back to application
                </Button>
            )}
        </div>
    );
}

interface Props {
    withBackButton?: boolean;
}

export const PageNoAccess = React.memo(View);
