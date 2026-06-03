import { useRef } from "react";

import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { useParams } from "react-router";
import { toast } from "sonner";
import invariant from "tiny-invariant";
import { ProjectDomainSettingsCommands, ProjectDomainSettingsQueries } from "~/projects/data";
import { ProjectPermissionSubmitButton } from "~/projects/module-shared/components";

import { AppLoader } from "@application/shared/components";
import { MODULE_IDS } from "@application/shared/constants";
import { PageError } from "@application/shared/pages";
import { useConditionalModule } from "@application/shared/permissions";

import { isValidationException } from "@infrastructure/api";

import { ValidationException } from "@infrastructure/exceptions/validation";

import { ProjectDomainSettingsForm } from "../form";
import { type ProjectDomainSettingsFormSchemaOutput, ProjectDomainSettingsKeyTypeUnspecified } from "../schemas";
import type { ProjectDomainSettingsFormRef } from "../types";

function NoteBox({ children }: { children: React.ReactNode }) {
    return (
        <div className={cn(dashedBorderBox, "text-center text-sm leading-6")}>
            <span className="text-orange-500">Note: </span>
            {children}
        </div>
    );
}

export function ProjectDomainSettingsRoute() {
    const { id: projectId } = useParams<{ id: string }>();
    const formRef = useRef<ProjectDomainSettingsFormRef>(null);
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.Project });

    invariant(projectId, "projectId must be defined");
    const resolvedProjectId = projectId;

    const settingsQuery = ProjectDomainSettingsQueries.useFindOne({ projectID: resolvedProjectId });

    const { mutate: update, isPending: isUpdating } = ProjectDomainSettingsCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Domain settings updated");
        },
        onError: err => {
            if (isValidationException(err)) {
                formRef.current?.onError(ValidationException.fromHttp(err));
            }
        },
    });

    function handleSubmit(values: ProjectDomainSettingsFormSchemaOutput) {
        if (!canWrite) {
            return;
        }

        const settings = settingsQuery.data?.data;
        invariant(settings, "domain settings must be defined");

        update({
            projectID: resolvedProjectId,
            payload: {
                updateVer: settings.updateVer,
                rootDomain: settings.rootDomain,
                allowedDomains: values.allowedDomains.map(item => item.value).filter(Boolean),
                certSettings: {
                    certType: values.certSettings.certType,
                    keyType:
                        values.certSettings.keyType === ProjectDomainSettingsKeyTypeUnspecified
                            ? undefined
                            : values.certSettings.keyType,
                    email: values.certSettings.email,
                    validPeriod: settings.certSettings?.validPeriod,
                    autoRenew: settings.certSettings?.autoRenew,
                },
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

    invariant(settingsQuery.data, "domain settings data must be defined");

    return (
        <ProjectDomainSettingsForm
            ref={formRef}
            defaultValues={settingsQuery.data.data}
            onSubmit={handleSubmit}
            readOnly={!canWrite}
            note={
                <NoteBox>
                    If you configure Allowed Domains, all apps in the project will only be allowed to use these domains.
                </NoteBox>
            }
        >
            <div className="flex justify-end mt-4">
                <ProjectPermissionSubmitButton isPending={isUpdating} />
            </div>
        </ProjectDomainSettingsForm>
    );
}
