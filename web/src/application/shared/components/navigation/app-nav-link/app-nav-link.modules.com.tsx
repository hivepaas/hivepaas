import { NavLink, type NavLinkProps, useLocation } from "react-router";

import { useAppLink } from "@application/shared/hooks/router";

export function AppNavLinkModules({ to, children, ignorePrevPath = false, ...props }: Props) {
    const { link } = useAppLink();

    const location = useLocation();

    console.log(link.modules(to));

    return (
        <NavLink
            {...props}
            to={link.modules(to)}
            state={ignorePrevPath ? (location.state as unknown) : { from: location.pathname + location.search }}
        >
            {children}
        </NavLink>
    );
}

type Props = NavLinkProps &
    React.RefAttributes<HTMLAnchorElement> & {
        ignorePrevPath?: boolean;
    };
