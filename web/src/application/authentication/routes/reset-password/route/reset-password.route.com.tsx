import { useSearchParams } from "react-router";
import { toast } from "sonner";

import { ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";

import { AuthCommands } from "@application/authentication/data/commands";
import type { ResetPassword } from "@application/authentication/domain";
import { AuthenticationLayout } from "@application/authentication/layouts";

import { ResetPasswordForm } from "../form/reset-password.form.com";

function View({ email, token }: ViewProps) {
    const { navigate } = useAppNavigate();

    const { mutate: resetPassword, isPending } = AuthCommands.useResetPassword({
        onSuccess: () => {
            toast.success("Password updated successfully");
            navigate(ROUTE.auth.signIn.$route, { replace: true });
        },
    });

    function handleSubmit(values: Pick<ResetPassword, "newPassword">) {
        resetPassword({
            email,
            token,
            newPassword: values.newPassword,
        });
    }

    return (
        <AuthenticationLayout>
            <ResetPasswordForm
                isPending={isPending}
                onSubmit={handleSubmit}
            />
        </AuthenticationLayout>
    );
}

interface ViewProps {
    email: string;
    token: string;
}

export const ResetPasswordRoute = () => {
    const [params] = useSearchParams();

    const email = params.get("email") ?? "";
    const token = params.get("token") ?? "";

    return (
        <View
            email={email}
            token={token}
        />
    );
};
