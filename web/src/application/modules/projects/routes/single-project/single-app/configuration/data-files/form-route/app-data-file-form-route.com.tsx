import { toast } from "sonner";
import { AppDataFilesCommands } from "~/projects/data/commands";

import { RouteFormHeader } from "@application/shared/components";
import { MODULE_IDS, ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";
import { useConditionalModule } from "@application/shared/permissions";

import { UploadAppDataFileForm } from "../form";
import type { UploadAppDataFileFormValues } from "../schemas";

export function AppDataFileFormRoute({ projectId, appId }: Props) {
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.Project });
    const { navigate } = useAppNavigate();

    function navigateToList() {
        navigate.modules(ROUTE.projects.single.apps.single.configuration.dataFiles.$route(projectId, appId), {
            ignorePrevPath: true,
        });
    }

    const { mutate: uploadLocal, isPending: isUploading } = AppDataFilesCommands.useUploadLocal({
        onSuccess: () => {
            toast.success("Data file(s) uploaded successfully");
            navigateToList();
        },
    });

    const { mutate: createOne, isPending: isCreating } = AppDataFilesCommands.useCreateOne({
        onSuccess: () => {
            toast.success("Cloud data file registered successfully");
            navigateToList();
        },
    });

    function onSubmit(values: UploadAppDataFileFormValues) {
        if (!canWrite) {
            return;
        }

        if (values["uploadMethod"] === "local") {
            uploadLocal({
                projectID: projectId,
                appID: appId,
                fileKind: values["fileKind"],
                files: values["files"],
            });
        } else {
            const { storage } = values;
            if (!storage) {
                return;
            }
            createOne({
                projectID: projectId,
                appID: appId,
                fileKind: values["fileKind"],
                filePath: values["filePath"],
                storageID: storage["id"],
                bucket: values["bucket"],
            });
        }
    }

    function handleClose() {
        navigateToList();
    }

    const isPending = isUploading || isCreating;

    return (
        <div className="flex w-full flex-col">
            <RouteFormHeader title="Upload Data File(s)" />

            <UploadAppDataFileForm
                isPending={isPending}
                projectId={projectId}
                onSubmit={onSubmit}
                onClose={handleClose}
            />
        </div>
    );
}

interface Props {
    projectId: string;
    appId: string;
}
