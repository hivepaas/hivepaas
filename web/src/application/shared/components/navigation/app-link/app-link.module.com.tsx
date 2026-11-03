import React from "react";

import { Link, type LinkProps, useLocation } from "react-router";

import { useAppLink } from "@application/shared/hooks/router";

export function AppLinkModules({ to, children, ignorePrevPath = false, ...props }: Props) {
    const { link } = useAppLink();

    const location = useLocation();

    return (
        <Link
            {...props}
            to={link.modules(to)}
            state={ignorePrevPath ? (location.state as unknown) : { from: location.pathname + location.search }}
        >
            {children}
        </Link>
    );
}

type Props = LinkProps &
    React.RefAttributes<HTMLAnchorElement> & {
        ignorePrevPath?: boolean;
    };
