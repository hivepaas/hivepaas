import { Navigate, type NavigateProps, useLocation } from "react-router";

import { useAppLink } from "@application/shared/hooks/router";

export function AppNavigateModules({ to, ignorePrevPath = false, ...props }: Props) {
    const { link } = useAppLink();

    const location = useLocation();

    return (
        <Navigate
            {...props}
            to={link.modules(to)}
            state={ignorePrevPath ? (location.state as unknown) : { from: location.pathname + location.search }}
        />
    );
}

type Props = NavigateProps & {
    ignorePrevPath?: boolean;
};
