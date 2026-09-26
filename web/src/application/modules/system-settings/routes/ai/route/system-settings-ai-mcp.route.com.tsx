import { useEffect } from "react";

import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { Controller, useForm } from "react-hook-form";
import { toast } from "sonner";
import { McpSettingsCommands, McpSettingsQueries } from "~/system-settings/data";
import { SectionHeader } from "~/system-settings/module-shared";

import { EnvConfig } from "@config";

import { AppLoader, FormActionBar, InfoBlock, LabelWithInfo } from "@application/shared/components";
import { MODULE_IDS } from "@application/shared/constants";
import { PermissionTooltipAction, useConditionalModule } from "@application/shared/permissions";

import { Button, Checkbox } from "@/components/ui";

import { CopyField, McpConnectSection, McpRecentCallsSection } from "../building-blocks";

interface FormValues {
    enabled: boolean;
}

/** Where the server answers: the API's own address, and /mcp under it. */
function mcpEndpoint(): string {
    return `${EnvConfig.API_URL.replace(/\/+$/, "")}/mcp`;
}

export function SystemSettingsAiMcpRoute() {
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.System });
    const { data, isLoading } = McpSettingsQueries.useFindOne();
    const settings = data?.data;

    const { control, handleSubmit, reset } = useForm<FormValues>({
        defaultValues: { enabled: settings?.enabled ?? false },
    });

    useEffect(() => {
        reset({ enabled: settings?.enabled ?? false });
    }, [settings, reset]);

    const { mutate: update, isPending } = McpSettingsCommands.useUpdateOne({
        onSuccess: (_response, request) => {
            toast.success(request.payload.enabled ? "MCP server enabled" : "MCP server disabled");
        },
    });

    function onSubmit(values: FormValues) {
        if (!canWrite) {
            return;
        }
        update({
            payload: {
                enabled: values.enabled,
                // Kept as it is: nothing reads it until the tools that change things exist.
                allowWrite: settings?.allowWrite ?? false,
                updateVer: settings?.updateVer ?? 0,
            },
        });
    }

    if (isLoading) {
        return <AppLoader />;
    }

    const endpoint = mcpEndpoint();

    return (
        <div className="pt-2">
            <form
                onSubmit={event => {
                    event.preventDefault();
                    void handleSubmit(onSubmit)(event);
                }}
                className="flex flex-col gap-6"
            >
                <div className={cn(dashedBorderBox)}>
                    <span className="font-semibold text-orange-500">Note:</span> The MCP server lets an AI assistant -
                    Claude Code, Claude Desktop, an editor - read HivePaaS for you: the status and logs of apps, tasks,
                    nodes, the app store and scheduled jobs. It uses an API key, sees only what the key&apos;s user can
                    see, and changes nothing. Every call is recorded in the audit log.
                </div>

                <SectionHeader>General</SectionHeader>
                <fieldset
                    disabled={!canWrite}
                    className="m-0 flex min-w-0 flex-col gap-6 border-0 px-3 py-0"
                >
                    <InfoBlock
                        titleWidth={220}
                        title={
                            <LabelWithInfo
                                label="Enabled"
                                content="Serve the Model Context Protocol. While it is off, the endpoint answers 404, whatever key is sent."
                            />
                        }
                    >
                        <Controller
                            control={control}
                            name="enabled"
                            render={({ field }) => (
                                <Checkbox
                                    checked={field.value}
                                    onCheckedChange={checked => {
                                        field.onChange(checked === true);
                                    }}
                                />
                            )}
                        />
                    </InfoBlock>
                    <InfoBlock
                        titleWidth={220}
                        title={
                            <LabelWithInfo
                                label="Endpoint"
                                content="The address a client connects to, over streamable HTTP."
                            />
                        }
                    >
                        <CopyField
                            what="Endpoint"
                            value={endpoint}
                        />
                    </InfoBlock>
                </fieldset>

                {/* Inside the form, before the sticky action bar, so Save stays last.
                    Neither is part of what Save sends. */}
                <McpConnectSection
                    endpoint={endpoint}
                    enabled={settings?.enabled ?? false}
                />
                <McpRecentCallsSection />

                <FormActionBar>
                    <PermissionTooltipAction
                        id={MODULE_IDS.System}
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
            </form>
        </div>
    );
}
