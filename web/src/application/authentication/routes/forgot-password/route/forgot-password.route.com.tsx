import { toast } from "sonner";

import { AuthCommands } from "@application/authentication/data/commands";
import type { ForgotPassword } from "@application/authentication/domain";
import { AuthenticationLayout } from "@application/authentication/layouts";

import { ForgotPasswordForm } from "../form/forgot-password.form.com";

export function ForgotPasswordRoute() {
    const { mutate: forgotPassword, isPending } = AuthCommands.useForgotPassword({
        onSuccess: () => {
            toast.success("An email has been sent to your email address. Please check your inbox.");
        },
    });

    function handleSubmit(values: ForgotPassword) {
        forgotPassword(values);
    }

    return (
        <AuthenticationLayout>
            <div className="w-full max-w-sm mx-auto">
                <ForgotPasswordForm
                    isPending={isPending}
                    onSubmit={handleSubmit}
                />
            </div>
        </AuthenticationLayout>
    );
}
