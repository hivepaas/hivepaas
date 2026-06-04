import { useMemo } from "react";

import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { toast } from "sonner";

import { useProfileContext } from "@application/shared/context";
import { SupportFeedbacksCommands } from "@application/shared/data/commands";

import { CreateFeedbackForm } from "../form";
import { useCreateFeedbackDialogState } from "../hooks";
import type { CreateFeedbackFormSchemaOutput } from "../schemas";

const fnPlaceholder = () => null;

export function CreateFeedbackDialog() {
    const { state, props: { onClose = fnPlaceholder } = {}, ...actions } = useCreateFeedbackDialogState();
    const profile = useProfileContext(profileState => profileState.profile);

    const defaultValues = useMemo(
        () => ({
            name: profile?.fullName ?? "",
            email: profile?.email ?? "",
        }),
        [profile?.email, profile?.fullName],
    );

    const { mutate: createFeedback, isPending } = SupportFeedbacksCommands.useCreateOne({
        onSuccess: () => {
            toast.success("Thank you for your feedback.");
            actions.close();
            onClose();
        },
    });

    function handleSubmit(values: CreateFeedbackFormSchemaOutput) {
        createFeedback({
            category: values.category,
            name: values.name.trim() || undefined,
            email: values.email.trim() || undefined,
            company: values.company.trim() || undefined,
            subject: values.subject,
            description: values.description,
        });
    }

    const open = state.mode !== "closed";

    return (
        <Dialog
            open={open}
            onOpenChange={nextOpen => {
                if (!nextOpen) {
                    actions.close();
                }
            }}
        >
            <DialogContent className="sm:max-w-[780px] max-h-[90vh] overflow-y-auto">
                <DialogHeader>
                    <DialogTitle className="text-2xl">Create your feedback</DialogTitle>
                </DialogHeader>
                <CreateFeedbackForm
                    defaultValues={defaultValues}
                    isPending={isPending}
                    onSubmit={handleSubmit}
                />
            </DialogContent>
        </Dialog>
    );
}
