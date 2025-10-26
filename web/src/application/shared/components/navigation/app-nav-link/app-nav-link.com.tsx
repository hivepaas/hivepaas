import React from "react";

import { NavLink, type NavLinkProps } from "react-router-dom";

import { useAppLink } from "@application/shared/hooks/router";

export function AppNavLink({
    to, //
    children,
    ...props
}: Props) {
    const { linkTo } = useAppLink();

    return (
        <NavLink
            to={linkTo(to)}
            {...props}
        >
            {children}
        </NavLink>
    );
}

type Props = NavLinkProps & React.RefAttributes<HTMLAnchorElement>;
