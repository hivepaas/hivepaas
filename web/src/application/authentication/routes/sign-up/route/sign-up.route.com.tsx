import { useState } from "react";

import { useSearchParams } from "react-router";
import { useMount } from "react-use";
import { toast } from "sonner";

import { AppLoader } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";
import { Page404NotFound, PageError } from "@application/shared/pages";

import { AuthCommands } from "@application/authentication/data/commands";
import type { SignUp } from "@application/authentication/domain";
import type { Candidate } from "@application/authentication/domain/user";
import { AuthenticationLayout } from "@application/authentication/layouts";

import { isInviteTokenInvalidException } from "@infrastructure/api";

import { SignUpForm } from "../form";

function View({ method }: ViewProps) {
    const { navigate } = useAppNavigate();

    const { mutate: signUp, isPending } = AuthCommands.useSignUp({
        onSuccess: () => {
            toast.success("Sign Up complete", {
                description: "You can now sign in to your account",
            });

            void navigate.basic(ROUTE.auth.signIn.$route);
        },
    });

    function handleSubmit(data: SignUp) {
        signUp({
            data,
        });
    }

    return (
        <AuthenticationLayout>
            <SignUpForm
                method={method}
                isPending={isPending}
                onSubmit={data => {
                    handleSubmit({
                        inviteToken: method.inviteToken,
                        data,
                    });
                }}
            />
        </AuthenticationLayout>
    );
}

interface ViewProps {
    method: { inviteToken: string; candidate: Candidate };
}

type State =
    | {
          type: "pending";
      }
    | {
          type: "invalid";
      }
    | {
          type: "success";
          method: { inviteToken: string; candidate: Candidate };
      }
    | {
          type: "error";
          error: Error;
      };

export function SignUpRoute() {
    const [state, setState] = useState<State>({ type: "pending" });
    console.log("SignUpRoute");

    const { mutate: validate } = AuthCommands.useValidateInviteToken({
        onSuccess: ({ data }, request) => {
            setState({
                type: "success",
                method: {
                    inviteToken: request.inviteToken,
                    candidate: data.candidate,
                },
            });
        },
        onError: error => {
            if (isInviteTokenInvalidException(error)) {
                setState({
                    type: "invalid",
                });

                return;
            }

            setState({
                type: "error",
                error,
            });
        },
    });

    const [params] = useSearchParams();

    const token = params.get("token");

    useMount(() => {
        switch (true) {
            case token === null: {
                setState({
                    type: "invalid",
                });

                break;
            }

            case token !== null: {
                validate({
                    inviteToken: token,
                });

                break;
            }
        }
    });

    if (state.type === "pending") {
        return <AppLoader />;
    }

    if (state.type === "invalid") {
        return <Page404NotFound />;
    }

    if (state.type === "error") {
        return (
            <PageError
                error={state.error}
                onRetry={() => {
                    if (token !== null) {
                        validate({
                            inviteToken: token,
                        });
                    }
                }}
            />
        );
    }

    return <View method={state.method} />;
}
