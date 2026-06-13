import { useMemo } from "react";

import { Plus } from "lucide-react";
import { useParams } from "react-router";
import invariant from "tiny-invariant";
import { APP_CONFIGURATION_QUERY_OPTIONS } from "~/projects/data/constants";
import { ProjectAppSecretsQueries } from "~/projects/data/queries";
import { useCreateOrEditAppSecretDialog } from "~/projects/dialogs/create-or-edit-app-secret/hooks";
import { AppSecretsTableDefs } from "~/projects/module-shared/definitions/tables/app-secrets";

import { TableActions } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA, MODULE_IDS } from "@application/shared/constants";
import { useTableState } from "@application/shared/hooks/table";
import { PermissionTooltipAction } from "@application/shared/permissions";

import { Button, DataTable } from "@/components/ui";
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from "@/components/ui/accordion";

export function AppConfigSecretsRoute() {
    const { id: projectId, appId } = useParams<{ id: string; appId: string }>();

    invariant(projectId, "projectId must be defined");
    invariant(appId, "appId must be defined");

    const { actions: secretDialogActions } = useCreateOrEditAppSecretDialog();
    const { pagination, setPagination, sorting, setSorting, search, setSearch } = useTableState();

    const { data: { data: secrets, meta } = DEFAULT_PAGINATED_DATA, isFetching } =
        ProjectAppSecretsQueries.useFindManyPaginated(
            {
                projectID: projectId,
                appID: appId,
                pagination,
                sorting,
                search,
            },
            APP_CONFIGURATION_QUERY_OPTIONS,
        );

    const ownSecrets = useMemo(() => secrets.filter(s => !s.inherited), [secrets]);
    const inheritedSecrets = useMemo(() => secrets.filter(s => s.inherited), [secrets]);

    const columns = useMemo(() => AppSecretsTableDefs.columns(projectId, appId), [projectId, appId]);

    return (
        <div className="flex flex-col gap-6">
            <TableActions
                search={{ value: search, onChange: setSearch }}
                renderActions={
                    <PermissionTooltipAction
                        id={MODULE_IDS.Project}
                        action="write"
                    >
                        {({ isDenied }) => (
                            <Button
                                onClick={() => {
                                    secretDialogActions.open(projectId, appId);
                                }}
                                disabled={isDenied}
                            >
                                <Plus className="size-4" /> New Secret
                            </Button>
                        )}
                    </PermissionTooltipAction>
                }
            />

            <Accordion
                type="multiple"
                defaultValue={["secrets", "inherited-secrets"]}
                className="w-full"
            >
                <AccordionItem
                    value="inherited-secrets"
                    className="border-none"
                >
                    <AccordionTrigger className="px-3 py-2 [&>svg]:rotate-90 [&[data-state=open]>svg]:rotate-0 data-[state=closed]:mb-6 bg-accent">
                        Inherited Secrets
                    </AccordionTrigger>
                    <AccordionContent className="p-0 pt-4 mb-4">
                        <DataTable
                            columns={columns}
                            data={inheritedSecrets}
                            isLoading={isFetching}
                            enablePagination={false}
                        />
                    </AccordionContent>
                </AccordionItem>
                <AccordionItem
                    value="secrets"
                    className="border-none"
                >
                    <AccordionTrigger className="px-3 py-2 [&>svg]:rotate-90 [&[data-state=open]>svg]:rotate-0 bg-accent">
                        Secrets
                    </AccordionTrigger>
                    <AccordionContent className="p-0 pt-4 mb-4">
                        <DataTable
                            columns={columns}
                            data={ownSecrets}
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
