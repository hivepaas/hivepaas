import { useRef } from "react";

import { Button } from "@components/ui";
import { dashedBorderBox, formBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { toast } from "sonner";
import invariant from "tiny-invariant";
import { AppPlacementSettingsCommands, AppPlacementSettingsQueries } from "~/settings/data";

import { AppLoader, FormActionBar } from "@application/shared/components";
import { MODULE_IDS } from "@application/shared/constants";
import { PageError } from "@application/shared/pages";
import { PermissionTooltipAction, useConditionalModule } from "@application/shared/permissions";

import { isValidationException } from "@infrastructure/api";

import { ValidationException } from "@infrastructure/exceptions/validation";

import { SettingsAppPlacementForm } from "../form";
import type { SettingsAppPlacementFormSchemaOutput } from "../schemas";
import type { SettingsAppPlacementFormRef } from "../types";

function NoteBox({ children }: { children: React.ReactNode }) {
    return (
        <div className={cn(dashedBorderBox, "text-sm leading-6")}>
            <span className="font-semibold text-orange-500">Note: </span>
            {children}
        </div>
    );
}

export function SettingsAppPlacementRoute() {
    const formRef = useRef<SettingsAppPlacementFormRef>(null);
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.Settings });

    const settingsQuery = AppPlacementSettingsQueries.useFindOne();

    const { mutate: update, isPending: isUpdating } = AppPlacementSettingsCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("App placement settings updated");
        },
        onError: err => {
            if (isValidationException(err)) {
                formRef.current?.onError(ValidationException.fromHttp(err));
            }
        },
    });

    function handleSubmit(values: SettingsAppPlacementFormSchemaOutput) {
        if (!canWrite) {
            return;
        }

        const settings = settingsQuery.data?.data;
        invariant(settings, "app placement settings must be defined");

        update({
            payload: {
                updateVer: settings.updateVer,
                inheritable: settings.inheritable ?? true,
                default: settings.default ?? true,
                excludeManagerNodes: values.excludeManagerNodes,
                excludeBuildNodes: values.excludeBuildNodes,
            },
        });
    }

    if (settingsQuery.isLoading) {
        return <AppLoader />;
    }

    if (settingsQuery.error) {
        return (
            <PageError
                error={settingsQuery.error}
                onRetry={settingsQuery.refetch}
            />
        );
    }

    invariant(settingsQuery.data, "app placement settings data must be defined");

    return (
        <div className={cn(formBox)}>
            <SettingsAppPlacementForm
                ref={formRef}
                defaultValues={settingsQuery.data.data}
                onSubmit={handleSubmit}
                readOnly={!canWrite}
                note={
                    <NoteBox>
                        If the cluster has only a single node, exclusion rules will be automatically bypassed to ensure
                        applications can still be deployed.
                    </NoteBox>
                }
                footer={
                    <FormActionBar>
                        <PermissionTooltipAction
                            id={MODULE_IDS.Settings}
                            action="write"
                        >
                            {({ isDenied }) => (
                                <Button
                                    type="submit"
                                    className="min-w-[100px]"
                                    disabled={isUpdating || isDenied}
                                    isLoading={isUpdating}
                                >
                                    Save
                                </Button>
                            )}
                        </PermissionTooltipAction>
                    </FormActionBar>
                }
            />
        </div>
    );
}
