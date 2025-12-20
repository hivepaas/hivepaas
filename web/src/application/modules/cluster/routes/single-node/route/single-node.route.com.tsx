import { useRef } from "react";

import { Button } from "@components/ui";
import { useParams } from "react-router";
import { toast } from "sonner";
import invariant from "tiny-invariant";
import { NodesCommands, NodesQueries } from "~/cluster/data";

import { AppLoader } from "@application/shared/components";

import { isValidationException } from "@infrastructure/api";

import { ValidationException } from "@infrastructure/exceptions/validation";

import { SingleNodeForm } from "../form";
import { type SingleNodeFormSchemaOutput } from "../schemas";
import { type SingleNodeFormRef } from "../types";

export function SingleNodeRoute() {
    const { id: nodeId } = useParams<{ id: string }>();
    const formRef = useRef<SingleNodeFormRef>(null);

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
        <div className="bg-background rounded-lg p-4 max-w-7xl w-full mx-auto">
            <SingleNodeForm
                ref={formRef}
                defaultValues={node}
                onSubmit={handleSubmit}
            >
                <div className="flex justify-end mt-4">
                    <Button
                        type="submit"
                        className="min-w-[100px]"
                        disabled={isPending}
                        isLoading={isPending}
                    >
                        Save
                    </Button>
                </div>
            </SingleNodeForm>
        </div>
    );
}
