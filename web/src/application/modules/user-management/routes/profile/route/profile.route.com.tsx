import { useEffect, useRef } from "react";

import { Button } from "@components/ui";
import { listBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { toast } from "sonner";
import invariant from "tiny-invariant";

import { type ProfilePhotoPayload } from "@application/shared/api/services";
import { FormActionBar } from "@application/shared/components";
import { useProfileContext } from "@application/shared/context";
import { ProfileCommands } from "@application/shared/data/commands";

import { ProfileForm } from "../form";
import { type ProfileFormSchemaOutput } from "../schemas";
import { type ProfileFormRef } from "../types";

export function ProfileRoute() {
    const { profile, setProfile } = useProfileContext();

    invariant(profile, "profile must be defined");

    const formRef = useRef<ProfileFormRef>(null);
    const photoActionRef = useRef<"update" | "delete">("update");

    const { mutate: updateProfile, isPending: isUpdatingProfile } = ProfileCommands.useUpdate({
        onSuccess: newProfile => {
            toast.success("Profile updated");
            setProfile(newProfile);
        },
    });

    const { mutate: updateProfilePhoto, isPending: isUpdatingProfilePhoto } = ProfileCommands.useUpdate({
        onSuccess: newProfile => {
            toast.success(photoActionRef.current === "delete" ? "Profile photo removed" : "Profile photo updated");
            setProfile(newProfile);
        },
    });

    function handleSubmit(values: ProfileFormSchemaOutput) {
        updateProfile({
            profile: {
                fullName: values.fullName,
                email: values.email,
                username: values.username,
                position: values.position,
                notes: values.notes,
            },
        });
    }

    function handlePhotoSubmit(photo: ProfilePhotoPayload) {
        photoActionRef.current = "delete" in photo ? "delete" : "update";

        updateProfilePhoto({
            profile: {
                photo,
            },
        });
    }

    useEffect(() => {
        formRef.current?.setValues({
            ...profile,
            fullName: profile.fullName ?? "",
            email: profile.email ?? "",
        });
    }, [profile]);

    return (
        <div className={cn(listBox)}>
            <ProfileForm
                ref={formRef}
                defaultValues={profile}
                onSubmit={handleSubmit}
                onPhotoSubmit={handlePhotoSubmit}
                isPhotoPending={isUpdatingProfilePhoto}
            >
                <FormActionBar>
                    <Button
                        type="submit"
                        className="min-w-[100px]"
                        disabled={isUpdatingProfile || isUpdatingProfilePhoto}
                        isLoading={isUpdatingProfile}
                    >
                        Save
                    </Button>
                </FormActionBar>
            </ProfileForm>
        </div>
    );
}
