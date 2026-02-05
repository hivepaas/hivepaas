import { useEffect, useRef } from "react";

import { Button } from "@components/ui";
import { listBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { toast } from "sonner";
import invariant from "tiny-invariant";

import { useProfileContext } from "@application/shared/context";
import { ProfileCommands } from "@application/shared/data/commands";

import { ProfileForm } from "../form";
import { type ProfileFormSchemaOutput } from "../schemas";
import { type ProfileFormRef } from "../types";

export function ProfileRoute() {
    const { profile, setProfile } = useProfileContext();

    invariant(profile, "profile must be defined");

    const formRef = useRef<ProfileFormRef>(null);

    const { mutate: updateProfile, isPending } = ProfileCommands.useUpdate({
        onSuccess: newProfile => {
            toast.success("Profile updated");
            setProfile(newProfile);
        },
    });

    function handleSubmit(values: ProfileFormSchemaOutput) {
        updateProfile({
            profile: {
                ...values,
                photo:
                    values.photo && values.photoUpload
                        ? { fileName: values.photoUpload.fileName, dataBase64: values.photoUpload.dataBase64 }
                        : { delete: true },
            },
        });
    }

    useEffect(() => {
        formRef.current?.setValues({
            ...profile,
            fullName: profile.fullName ?? "",
            email: profile.email ?? "",
            photoUpload: null,
        });
    }, [profile]);

    return (
        <div className={cn(listBox)}>
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
