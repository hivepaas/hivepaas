import { useState } from "react";

import { useSearchParams } from "react-router";
import { useMount } from "react-use";
import { toast } from "sonner";
import { z } from "zod";

import { AppLoader } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";
import { Page404NotFound } from "@application/shared/pages";

import { AuthCommands } from "@application/authentication/data/commands";
import type { ResetPassword } from "@application/authentication/domain";
import { AuthenticationLayout } from "@application/authentication/layouts";

import { ResetPasswordForm } from "../form/reset-password.form.com";

function View({ userId, token }: ViewProps) {
    const { navigate } = useAppNavigate();

    const { mutate: resetPassword, isPending } = AuthCommands.useResetPassword({
        onSuccess: () => {
            toast.success("Password updated successfully");
            void navigate.basic(ROUTE.auth.signIn.$route, { replace: true });
        },
    });

    function handleSubmit(values: Pick<ResetPassword, "newPassword">) {
        resetPassword({
            userId,
            token,
            newPassword: values.newPassword,
        });
    }

    return (
        <AuthenticationLayout>
            <div className="w-full max-w-sm mx-auto">
                <ResetPasswordForm
                    isPending={isPending}
                    onSubmit={handleSubmit}
                />
            </div>
        </AuthenticationLayout>
    );
}

interface ViewProps {
    userId: string;
    token: string;
}

const Schema = z.object({
    userId: z.string().trim().nonempty(),
    token: z.string().trim().nonempty(),
});

type State =
    | {
          type: "pending";
      }
    | {
          type: "invalid";
      }
    | {
          type: "success";
          userId: string;
          token: string;
      };

export function ResetPasswordRoute() {
    const [state, setState] = useState<State>({ type: "pending" });

    const [params] = useSearchParams();

    function validate() {
        const parsed = Schema.safeParse({
            userId: params.get("userID") ?? params.get("userId"),
            token: params.get("token"),
        });

        if (!parsed.success) {
            setState({
                type: "invalid",
            });

            return;
        }

        setState({
            type: "success",
            userId: parsed.data.userId,
            token: parsed.data.token,
        });
    }

    useMount(validate);

    if (state.type === "pending") {
        return <AppLoader />;
    }

    if (state.type === "invalid") {
        return <Page404NotFound />;
    }

    return (
        <View
            userId={state.userId}
            token={state.token}
        />
    );
}
