import { type PropsWithChildren } from "react";

export function AuthenticationLayout({ children }: PropsWithChildren) {
    return (
        <div className="flex min-h-svh w-full items-center justify-center p-6 md:p-10">
            <div className="w-full max-w-sm">{children}</div>
        </div>
    );
}
