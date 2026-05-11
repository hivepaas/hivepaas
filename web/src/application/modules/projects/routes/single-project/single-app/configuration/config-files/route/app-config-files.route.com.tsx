import { useMemo } from "react";

import { Plus } from "lucide-react";
import { useParams } from "react-router";
import invariant from "tiny-invariant";
import { AppConfigFilesQueries } from "~/projects/data/queries";
import { useCreateOrEditAppConfigFileDialog } from "~/projects/dialogs/create-or-edit-app-config-file/hooks";
import { AppConfigFilesTableDefs } from "~/projects/module-shared/definitions/tables/app-config-files";

import { TableActions } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA } from "@application/shared/constants";
import { useTableState } from "@application/shared/hooks/table";

import { Button, DataTable } from "@/components/ui";
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from "@/components/ui/accordion";

export function AppConfigFilesRoute() {
    const { id: projectId, appId } = useParams<{ id: string; appId: string }>();

    invariant(projectId, "projectId must be defined");
    invariant(appId, "appId must be defined");

    const { actions: configFileDialogActions } = useCreateOrEditAppConfigFileDialog();
    const { pagination, setPagination, sorting, setSorting, search, setSearch } = useTableState();

    const { data: { data: configFiles, meta } = DEFAULT_PAGINATED_DATA, isFetching } =
        AppConfigFilesQueries.useFindManyPaginated({
            projectID: projectId,
            appID: appId,
            pagination,
            sorting,
            search,
        });

    const ownConfigFiles = useMemo(() => configFiles.filter(configFile => !configFile.inherited), [configFiles]);
    const inheritedConfigFiles = useMemo(() => configFiles.filter(configFile => configFile.inherited), [configFiles]);

    const columns = useMemo(() => AppConfigFilesTableDefs.columns(projectId, appId), [projectId, appId]);

    return (
        <div className="flex flex-col gap-6">
            <TableActions
                search={{ value: search, onChange: setSearch }}
                renderActions={
                    <Button
                        onClick={() => {
                            configFileDialogActions.open(projectId, appId);
                        }}
                    >
                        <Plus className="mr-2 h-4 w-4" /> New Config File
                    </Button>
                }
            />

            <Accordion
                type="multiple"
                defaultValue={["config-files", "inherited-config-files"]}
                className="w-full"
            >
                <AccordionItem
                    value="inherited-config-files"
                    className="border-none"
                >
                    <AccordionTrigger className="px-3 py-2 [&>svg]:rotate-90 [&[data-state=open]>svg]:rotate-0 bg-accent">
                        Inherited Config Files
                    </AccordionTrigger>
                    <AccordionContent className="p-0 pt-4 mb-4">
                        <DataTable
                            columns={columns}
                            data={inheritedConfigFiles}
                            isLoading={isFetching}
                            enablePagination={false}
                        />
                    </AccordionContent>
                </AccordionItem>
                <AccordionItem
                    value="config-files"
                    className="border-none"
                >
                    <AccordionTrigger className="px-3 py-2 [&>svg]:rotate-90 [&[data-state=open]>svg]:rotate-0 bg-accent">
                        Config Files
                    </AccordionTrigger>
                    <AccordionContent className="p-0 pt-4 mb-4">
                        <DataTable
                            columns={columns}
                            data={ownConfigFiles}
                            isLoading={isFetching}
                            pageSize={pagination.size}
                            manualPagination
                            totalCount={meta.page.total}
                            onPaginationChange={setPagination}
                            manualSorting
                            onSortingChange={setSorting}
                            enablePagination
                            showPageSizeSelector={false}
                        />
                    </AccordionContent>
                </AccordionItem>
            </Accordion>
        </div>
    );
}
