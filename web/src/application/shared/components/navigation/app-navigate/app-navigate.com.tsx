import { Navigate, type NavigateProps, useLocation } from "react-router";

export function AppNavigate({ ignorePrevPath = false, ...props }: Props) {
    const location = useLocation();

    return (
        <Navigate
            {...props}
            state={ignorePrevPath ? (location.state as unknown) : { from: location.pathname + location.search }}
        />
    );
}

type Props = NavigateProps & {
    ignorePrevPath?: boolean;
};
