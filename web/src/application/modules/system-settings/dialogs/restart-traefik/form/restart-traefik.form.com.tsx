import { Button } from "@components/ui/button";
import { Checkbox } from "@components/ui/checkbox";
import { DialogActionFooter, DialogBody } from "@components/ui/dialog";
import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldErrors, useController, useForm } from "react-hook-form";

import {
    RestartTraefikFormSchema,
    type RestartTraefikFormSchemaInput,
    type RestartTraefikFormSchemaOutput,
} from "../schemas";

interface Props {
    isPending: boolean;
    onSubmit: (values: RestartTraefikFormSchemaOutput) => void;
}

export function RestartTraefikForm({ isPending, onSubmit }: Props) {
    const {
        handleSubmit,
        control,
        watch,
        formState: { errors },
    } = useForm<RestartTraefikFormSchemaInput, unknown, RestartTraefikFormSchemaOutput>({
        defaultValues: {
            restartTraefikService: true,
        },
        resolver: zodResolver(RestartTraefikFormSchema),
        mode: "onChange",
    });

    const { field: restartTraefikService } = useController({ name: "restartTraefikService", control });

    const watched = watch();
    const isAnySelected = watched.restartTraefikService;

    function onValid(values: RestartTraefikFormSchemaOutput) {
        onSubmit(values);
    }

    function onInvalid(_errors: FieldErrors<RestartTraefikFormSchemaOutput>) {
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
                            id="restartTraefikService"
                            checked={restartTraefikService.value}
                            onCheckedChange={restartTraefikService.onChange}
                        />
                        <label
                            htmlFor="restartTraefikService"
                            className="text-sm font-medium leading-none cursor-pointer peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
                        >
                            Traefik Service
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
