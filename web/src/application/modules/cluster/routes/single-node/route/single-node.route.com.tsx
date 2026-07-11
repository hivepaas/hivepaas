import { useRef } from "react";

import { Button } from "@components/ui";
import { listBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { useParams } from "react-router";
import { toast } from "sonner";
import invariant from "tiny-invariant";
import { NodesCommands, NodesQueries } from "~/cluster/data";

import { AppLoader, FormActionBar } from "@application/shared/components";
import { MODULE_IDS } from "@application/shared/constants";
import { PermissionTooltipAction, useConditionalModule } from "@application/shared/permissions";

import { isValidationException } from "@infrastructure/api";

import { ValidationException } from "@infrastructure/exceptions/validation";

import { SingleNodeForm } from "../form";
import { type SingleNodeFormSchemaOutput } from "../schemas";
import { type SingleNodeFormRef } from "../types";

export function SingleNodeRoute() {
    const { id: nodeId } = useParams<{ id: string }>();
    const formRef = useRef<SingleNodeFormRef>(null);
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.Cluster });

    invariant(nodeId, "nodeId must be defined");

    const { data, isLoading, error } = NodesQueries.useFindOneById({ id: nodeId });

    const { mutate: update, isPending } = NodesCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Node information updated");
        },
        onError: err => {
            if (isValidationException(err)) {
                formRef.current?.onError(ValidationException.fromHttp(err));
            }
        },
    });

    function handleSubmit(values: SingleNodeFormSchemaOutput) {
        if (!canWrite) {
            return;
        }

        invariant(nodeId, "nodeId must be defined");
        invariant(data, "data must be defined");

        update({
            id: nodeId,
            ...values,
            updateVer: data.data.updateVer,
        });
    }

    if (isLoading) {
        return <AppLoader />;
    }

    if (error) {
        return <div className="text-red-500">Error: {error.message}</div>;
    }

    invariant(data, "data must be defined");

    const { data: node } = data;

    return (
        <div className={cn(listBox)}>
            <SingleNodeForm
                ref={formRef}
                defaultValues={node}
                onSubmit={handleSubmit}
                readOnly={!canWrite}
            >
                <FormActionBar>
                    <PermissionTooltipAction
                        id={MODULE_IDS.Cluster}
                        action="write"
                    >
                        {({ isDenied }) => (
                            <Button
                                type="submit"
                                className="min-w-[100px]"
                                disabled={isPending || isDenied}
                                isLoading={isPending}
                            >
                                Save
                            </Button>
                        )}
                    </PermissionTooltipAction>
                </FormActionBar>
            </SingleNodeForm>
        </div>
    );
}
