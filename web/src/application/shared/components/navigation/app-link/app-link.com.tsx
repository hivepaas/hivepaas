import React from "react";

import { Link, type LinkProps } from "react-router-dom";

import { useAppLink } from "@application/shared/hooks/router";

export function AppLink({
    to, //
    children,
    ...props
}: Props) {
    const { linkTo } = useAppLink();

    return (
        <Link
            to={linkTo(to)}
            {...props}
        >
            {children}
        </Link>
    );
}

type Props = LinkProps & React.RefAttributes<HTMLAnchorElement>;
