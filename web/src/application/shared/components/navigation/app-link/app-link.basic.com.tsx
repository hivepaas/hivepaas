import { Link, type LinkProps, useLocation } from "react-router";

export function AppLinkBasic({ children, ignorePrevPath = false, ...props }: Props) {
    const location = useLocation();

    return (
        <Link
            {...props}
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
