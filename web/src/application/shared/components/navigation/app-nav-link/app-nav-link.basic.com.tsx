import { NavLink, type NavLinkProps, useLocation } from "react-router";

export function AppNavLinkBasic({ children, ignorePrevPath = false, ...props }: Props) {
    const location = useLocation();

    return (
        <NavLink
            {...props}
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
