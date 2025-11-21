import { useEffect, useRef } from "react";

import { Button } from "@components/ui";
import { toast } from "sonner";
import invariant from "tiny-invariant";
import { UsersCommands } from "~/user-management/data/commands";
import { UsersQueries } from "~/user-management/data/queries";

import { AppLoader } from "@application/shared/components";
import { useProfileContext } from "@application/shared/context";

import { ProfileForm } from "../form";
import { type ProfileFormSchemaOutput } from "../schemas";
import { type ProfileFormRef } from "../types";

export function ProfileRoute() {
    const { profile } = useProfileContext();

    invariant(profile, "profile must be defined");

    const formRef = useRef<ProfileFormRef>(null);

    const { data, isLoading, error } = UsersQueries.useFindOneById({ id: profile.id });

    const { mutate: update, isPending } = UsersCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("User base information updated");
        },
    });

    function handleSubmit(values: ProfileFormSchemaOutput) {
        console.log(values);
        // update({
        //     user: {
        //         ...values,
        //         id: profile.id,
        //         status: user.status,
        //     },
        // });
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
        <div className="bg-background rounded-lg p-4 container mx-auto">
            <ProfileForm
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
            </ProfileForm>
        </div>
    );
}
