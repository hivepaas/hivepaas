import { useEffect, useRef } from "react";

import { Button } from "@components/ui";
import { useParams } from "react-router";
import { toast } from "sonner";
import invariant from "tiny-invariant";

import { AppLoader } from "@application/shared/components";

import { UsersCommands } from "@application/modules/user-management/data/commands";
import { UsersQueries } from "@application/modules/user-management/data/queries";

import { isValidationException } from "@infrastructure/api/utils/api.exceptions.utils";

import { HttpException } from "@infrastructure/exceptions/http";
import { ValidationException } from "@infrastructure/exceptions/validation";

import { SingleUserForm } from "../form";
import { type SingleUserFormSchemaOutput } from "../schemas";
import { type SingleUserFormRef } from "../types";

export function SingleUserRoute() {
    const { id: userId } = useParams<{ id: string }>();

    const formRef = useRef<SingleUserFormRef>(null);

    invariant(userId, "userId must be defined");

    const { data, isLoading, error } = UsersQueries.useFindOneById({ id: userId });

    const { mutate: update, isPending } = UsersCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("User base information updated");
        },
        onError: err => {
            if (isValidationException(err)) {
                formRef.current?.onError(ValidationException.fromHttp(err));
            }
        },
    });

    function handleSubmit(values: SingleUserFormSchemaOutput) {
        invariant(userId, "userId must be defined");

        update({
            user: {
                ...values,
                id: userId,
                status: user.status,
            },
        });
    }

    useEffect(() => {
        if (!data) {
            return;
        }

        formRef.current?.setValues(data.data);
    }, [data]);

    if (isLoading) {
        return <AppLoader />;
    }

    if (error) {
        return <div className="text-red-500">Error: {error.message}</div>;
    }

    invariant(data, "data must be defined");

    const { data: user } = data;
    return (
        <div className="bg-background rounded-lg p-4 max-w-7xl w-full mx-auto">
            <SingleUserForm
                ref={formRef}
                defaultValues={user}
                onSubmit={handleSubmit}
            >
                <div className="flex justify-end">
                    <Button
                        type="submit"
                        className="min-w-[100px]"
                        disabled={isPending}
                        isLoading={isPending}
                    >
                        Save
                    </Button>
                </div>
            </SingleUserForm>
        </div>
    );
}
