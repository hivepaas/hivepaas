import { useCallback } from "react";

import { type NavigateOptions, type To, useNavigate } from "react-router-dom";

import { useAppLink } from "@application/shared/hooks/router";

function createHook() {
    return function useAppNavigate() {
        const { linkTo } = useAppLink();

        const navigate = useNavigate();

        const navigateLocalized = useCallback(
            (to: To, options?: NavigateOptions) => {
                navigate(linkTo(to), options);
            },
            [linkTo, navigate],
        );

        return {
            navigate: navigateLocalized,
        };
    };
}

export const useAppNavigate = createHook();
