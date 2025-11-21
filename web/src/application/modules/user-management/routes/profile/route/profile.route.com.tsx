import { useEffect, useRef } from "react";

import { Button } from "@components/ui";
import { toast } from "sonner";
import invariant from "tiny-invariant";
import { UsersCommands } from "~/user-management/data/commands";

import { useProfileContext } from "@application/shared/context";

import { ProfileForm } from "../form";
import { type ProfileFormSchemaOutput } from "../schemas";
import { type ProfileFormRef } from "../types";

export function ProfileRoute() {
    const { profile } = useProfileContext();

    invariant(profile, "profile must be defined");

    const formRef = useRef<ProfileFormRef>(null);

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
        formRef.current?.setValues({
            ...profile,
            fullName: profile.fullName ?? "",
            email: profile.email ?? "",
        });
    }, [profile]);

    return (
        <div className="bg-background rounded-lg p-4 container mx-auto">
            <ProfileForm
                ref={formRef}
                defaultValues={profile}
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
