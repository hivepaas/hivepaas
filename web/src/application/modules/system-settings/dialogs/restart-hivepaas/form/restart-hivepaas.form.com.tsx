import { Button } from "@components/ui/button";
import { Checkbox } from "@components/ui/checkbox";
import { DialogActionFooter, DialogBody } from "@components/ui/dialog";
import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldErrors, useController, useForm } from "react-hook-form";

import {
    RestartHivePaaSFormSchema,
    type RestartHivePaaSFormSchemaInput,
    type RestartHivePaaSFormSchemaOutput,
} from "../schemas";

interface Props {
    isPending: boolean;
    onSubmit: (values: RestartHivePaaSFormSchemaOutput) => void;
}

export function RestartHivePaaSForm({ isPending, onSubmit }: Props) {
    const {
        handleSubmit,
        control,
        watch,
        formState: { errors },
    } = useForm<RestartHivePaaSFormSchemaInput, unknown, RestartHivePaaSFormSchemaOutput>({
        defaultValues: {
            restartMainApp: false,
            restartDbApp: false,
            restartCacheApp: false,
            restartWorkers: false,
            restartAgents: false,
        },
        resolver: zodResolver(RestartHivePaaSFormSchema),
        mode: "onChange",
    });

    const { field: restartMainApp } = useController({ name: "restartMainApp", control });
    const { field: restartDbApp } = useController({ name: "restartDbApp", control });
    const { field: restartCacheApp } = useController({ name: "restartCacheApp", control });
    const { field: restartWorkers } = useController({ name: "restartWorkers", control });
    const { field: restartAgents } = useController({ name: "restartAgents", control });

    const watched = watch();
    const isAnySelected =
        watched.restartMainApp ||
        watched.restartDbApp ||
        watched.restartCacheApp ||
        watched.restartWorkers ||
        watched.restartAgents;

    function onValid(values: RestartHivePaaSFormSchemaOutput) {
        onSubmit(values);
    }

    function onInvalid(_errors: FieldErrors<RestartHivePaaSFormSchemaOutput>) {
        console.log(_errors);
    }

    return (
        <form
            onSubmit={event => {
                event.preventDefault();
                void handleSubmit(onValid, onInvalid)(event);
            }}
            className="min-h-0 flex flex-1 flex-col"
        >
            <DialogBody className="flex flex-col gap-4">
                <p className="text-sm font-medium text-foreground">Please select the services you want to restart.</p>

                <div className="flex flex-col gap-3 rounded-lg border bg-background/50 p-3">
                    <div className="flex items-center space-x-3">
                        <Checkbox
                            id="restartMainApp"
                            checked={restartMainApp.value}
                            onCheckedChange={restartMainApp.onChange}
                        />
                        <label
                            htmlFor="restartMainApp"
                            className="text-sm font-medium leading-none cursor-pointer peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
                        >
                            Main service
                        </label>
                    </div>

                    <div className="flex items-center space-x-3">
                        <Checkbox
                            id="restartDbApp"
                            checked={restartDbApp.value}
                            onCheckedChange={restartDbApp.onChange}
                        />
                        <label
                            htmlFor="restartDbApp"
                            className="text-sm font-medium leading-none cursor-pointer peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
                        >
                            DB service
                        </label>
                    </div>

                    <div className="flex items-center space-x-3">
                        <Checkbox
                            id="restartCacheApp"
                            checked={restartCacheApp.value}
                            onCheckedChange={restartCacheApp.onChange}
                        />
                        <label
                            htmlFor="restartCacheApp"
                            className="text-sm font-medium leading-none cursor-pointer peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
                        >
                            Cache service
                        </label>
                    </div>

                    <div className="flex items-center space-x-3">
                        <Checkbox
                            id="restartWorkers"
                            checked={restartWorkers.value}
                            onCheckedChange={restartWorkers.onChange}
                        />
                        <label
                            htmlFor="restartWorkers"
                            className="text-sm font-medium leading-none cursor-pointer peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
                        >
                            Worker service(s)
                        </label>
                    </div>

                    <div className="flex items-center space-x-3">
                        <Checkbox
                            id="restartAgents"
                            checked={restartAgents.value}
                            onCheckedChange={restartAgents.onChange}
                        />
                        <label
                            htmlFor="restartAgents"
                            className="text-sm font-medium leading-none cursor-pointer peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
                        >
                            Agent service(s)
                        </label>
                    </div>
                </div>

                {errors.root?.message && <p className="text-xs text-destructive">{errors.root.message}</p>}
            </DialogBody>

            <DialogActionFooter>
                <Button
                    type="submit"
                    variant="destructive"
                    isLoading={isPending}
                    disabled={!isAnySelected || isPending}
                    className="min-w-[100px]"
                >
                    Restart
                </Button>
            </DialogActionFooter>
        </form>
    );
}
