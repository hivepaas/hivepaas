import { useRef, useState } from "react";

import { EyeOffIcon } from "lucide-react";
import { useFormContext, useWatch } from "react-hook-form";
import { useParams } from "react-router";
import { toast } from "sonner";
import invariant from "tiny-invariant";
import { AppKindSettingsCommands, AppKindSettingsQueries } from "~/projects/data";
import { APP_CONFIGURATION_QUERY_OPTIONS } from "~/projects/data/constants";
import { type AppKindSettingsUpdatePayload } from "~/projects/domain";
import { ProjectPermissionSubmitButton } from "~/projects/module-shared/components";
import { ConfirmRevealSecretsDialog, RevealSecretsButton } from "~/settings/module-shared/components";

import { AppLoader, FormActionBar } from "@application/shared/components";
import { CAPABILITY_IDS, MODULE_IDS } from "@application/shared/constants";
import { useProfileContext } from "@application/shared/context";
import { EUserRole } from "@application/shared/enums";
import { useCapability, useConditionalModule } from "@application/shared/permissions";

import { isValidationException } from "@infrastructure/api";

import { ValidationException } from "@infrastructure/exceptions/validation";

import { Button } from "@/components/ui/button";

import { AppConfigKindSettingsForm } from "../form";
import { type AppConfigKindSettingsFormSchemaInput, type AppConfigKindSettingsFormSchemaOutput } from "../schemas";
import { type AppConfigKindSettingsFormRef } from "../types";

function mapFormValuesToPayload(
    values: AppConfigKindSettingsFormSchemaOutput,
    updateVer: number,
): AppKindSettingsUpdatePayload {
    return {
        category: values.category,
        engine: values.engine,
        port: values.port,
        version: values.version,
        updateVer,
        ...(values.category === "webapp" ? { webapp: {} } : {}),
        ...(values.category === "database"
            ? {
                  database: {
                      dbName: values.database?.dbName ?? "",
                      username: values.database?.username ?? "",
                      password: values.database?.password ?? "",
                      rootPassword: values.database?.rootPassword ?? "",
                      sslMode: values.database?.sslMode ?? "disable",
                      sslCert: values.database?.sslCert?.id ? { id: values.database.sslCert.id } : null,
                      tlsPassthrough: values.database?.tlsPassthrough ?? false,
                  },
              }
            : {}),
        ...(values.category === "cache"
            ? {
                  cache: {
                      password: values.cache?.password ?? "",
                      maxMemory: values.cache?.maxMemory ?? "",
                      evictionRule: values.cache?.evictionRule ?? "",
                      persistenceMode: values.cache?.persistenceMode ?? "",
                      sslCert: values.cache?.sslCert?.id ? { id: values.cache.sslCert.id } : null,
                  },
              }
            : {}),
        ...(values.category === "storage"
            ? {
                  storage: {
                      keyId: values.storage?.keyId ?? "",
                      secret: values.storage?.secret ?? "",
                      bucket: values.storage?.bucket ?? "",
                      region: values.storage?.region ?? "",
                  },
              }
            : {}),
    };
}

function KindSettingsRevealSecretsButton({
    canShow,
    isRevealed,
    onReveal,
    onHide,
    isLoading,
    disabled,
}: {
    canShow: boolean;
    isRevealed: boolean;
    onReveal: () => void;
    onHide: () => void;
    isLoading: boolean;
    disabled: boolean;
}) {
    const { control } = useFormContext<AppConfigKindSettingsFormSchemaInput>();
    const category = useWatch({ control, name: "category" });

    if (!canShow || category === "webapp") {
        return null;
    }

    if (isRevealed) {
        return (
            <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={onHide}
                disabled={disabled}
                className="gap-1.5 shrink-0 px-2.5 sm:px-3 text-xs sm:text-sm h-8 sm:h-9"
            >
                <EyeOffIcon className="size-3.5 sm:size-4" />
                <span>Hide Secrets</span>
            </Button>
        );
    }

    return (
        <RevealSecretsButton
            onClick={onReveal}
            isLoading={isLoading}
            disabled={disabled}
        />
    );
}

export function AppConfigKindSettingsRoute() {
    const { id: projectId, env, appId } = useParams<{ id: string; env: string; appId: string }>();
    const formRef = useRef<AppConfigKindSettingsFormRef>(null);
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.Project });
    const [isRevealed, setIsRevealed] = useState(false);
    const [isRevealDialogOpen, setIsRevealDialogOpen] = useState(false);

    const profile = useProfileContext(state => state.profile);
    const isAdmin = profile?.role === EUserRole.Admin;
    const { hasCapability: canRevealSecrets } = useCapability(CAPABILITY_IDS.SecretReveal);
    const canShowRevealButton = isAdmin || canRevealSecrets;

    invariant(projectId, "projectId must be defined");
    invariant(env, "env must be defined");
    invariant(appId, "appId must be defined");

    const { data, isLoading, isFetching } = AppKindSettingsQueries.useFindOne(
        {
            projectID: projectId,
            env,
            appID: appId,
            revealSecrets: isRevealed,
        },
        APP_CONFIGURATION_QUERY_OPTIONS,
    );

    const { mutate: update, isPending } = AppKindSettingsCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("App kind settings updated");
        },
        onError: err => {
            if (isValidationException(err)) {
                formRef.current?.onError(ValidationException.fromHttp(err));
            } else if (err instanceof Error) {
                toast.error(err.message);
            } else {
                toast.error("Failed to update app kind settings");
            }
        },
    });

    function handleSubmit(values: AppConfigKindSettingsFormSchemaOutput) {
        if (!canWrite) return;

        invariant(projectId, "projectId must be defined");
        invariant(env, "env must be defined");
        invariant(appId, "appId must be defined");

        update({
            projectID: projectId,
            env,
            appID: appId,
            payload: mapFormValuesToPayload(values, data?.data.updateVer ?? 0),
        });
    }

    if (isLoading) {
        return <AppLoader />;
    }

    return (
        <div className="flex flex-col gap-4">
            <AppConfigKindSettingsForm
                ref={formRef}
                defaultValues={data?.data}
                onSubmit={handleSubmit}
                readOnly={!canWrite}
                isRevealed={isRevealed}
            >
                <FormActionBar>
                    <KindSettingsRevealSecretsButton
                        canShow={canShowRevealButton}
                        isRevealed={isRevealed}
                        onReveal={() => {
                            setIsRevealDialogOpen(true);
                        }}
                        onHide={() => {
                            setIsRevealed(false);
                        }}
                        isLoading={isFetching && isRevealed}
                        disabled={isPending}
                    />
                    <ProjectPermissionSubmitButton isPending={isPending} />
                </FormActionBar>
            </AppConfigKindSettingsForm>

            <ConfirmRevealSecretsDialog
                open={isRevealDialogOpen}
                onOpenChange={setIsRevealDialogOpen}
                onConfirm={() => {
                    setIsRevealed(true);
                    setIsRevealDialogOpen(false);
                    toast.success("Secrets revealed successfully");
                }}
                isPending={isFetching && isRevealed}
            />
        </div>
    );
}
