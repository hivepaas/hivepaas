import { useState } from "react";

import { useSearchParams } from "react-router";
import { useMount } from "react-use";
import { toast } from "sonner";
import { z } from "zod";

import { AppLoader } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";
import { Page404NotFound, PageError } from "@application/shared/pages";

import { AuthCommands } from "@application/authentication/data/commands";
import type { ResetPassword } from "@application/authentication/domain";
import { AuthenticationLayout } from "@application/authentication/layouts";

import { isPasswordResetTokenInvalidException } from "@infrastructure/api";

import { ResetPasswordForm } from "../form/reset-password.form.com";

function View({ email, token }: ViewProps) {
    const { navigate } = useAppNavigate();

    const { mutate: resetPassword, isPending } = AuthCommands.useResetPassword({
        onSuccess: () => {
            toast.success("Password updated successfully");
            void navigate.basic(ROUTE.auth.signIn.$route, { replace: true });
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

const Schema = z.object({
    email: z.string().trim().email(),
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
          email: string;
          token: string;
      }
    | {
          type: "error";
          error: Error;
      };

export function ResetPasswordRoute() {
    const [state, setState] = useState<State>({ type: "pending" });

    const [params] = useSearchParams();

    const { mutate, reset } = AuthCommands.useValidateResetToken({
        onSuccess: (_, request) => {
            setState({
                type: "success",
                email: request.email,
                token: request.token,
            });
        },
        onError: error => {
            setState({
                type: "error",
                error,
            });
        },
    });

    function validate() {
        const parsed = Schema.safeParse({
            email: params.get("email"),
            token: params.get("token"),
        });

        if (!parsed.success) {
            setState({
                type: "invalid",
            });

            return;
        }

        mutate({
            email: parsed.data.email,
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

    if (state.type === "error") {
        if (isPasswordResetTokenInvalidException(state.error)) {
            return <Page404NotFound />;
        }

        return (
            <PageError
                error={state.error}
                onRetry={() => {
                    reset();

                    validate();
                }}
            />
        );
    }
    return (
        <View
            email={state.email}
            token={state.token}
        />
    );
}
