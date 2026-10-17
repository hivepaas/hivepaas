import { useId } from "react";

import { Textarea } from "@components/ui/textarea";
import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { SectionHeader } from "~/system-settings/module-shared";

import { licenseKeyRefusalMessage } from "@application/shared/api";
import { InfoBlock, LabelWithInfo } from "@application/shared/components";
import { MODULE_IDS } from "@application/shared/constants";
import { LicenseCommands } from "@application/shared/data/commands";
import { useGlobalAlertDialogState } from "@application/shared/dialogs";
import type { License } from "@application/shared/entities";
import { PermissionTooltipAction } from "@application/shared/permissions";
import { licensePlanName } from "@application/shared/utils";

import { Button } from "@/components/ui";
import { FieldError } from "@/components/ui/field";

import { keyInForce, licenseInForceSummary } from "../building-blocks";
import { type LicenseKeyFormInput, type LicenseKeyFormOutput, LicenseKeyFormSchema } from "../schemas";

interface Props {
    /** The license in force: replacing a key that holds is confirmed first. */
    license: License;
    canWrite: boolean;
}

/**
 * Where a key is pasted: the one the portal shows or emails, for a purchase, a
 * trial or a renewal. A key the server refuses - one that does not verify, is
 * for another installation, or is past its grace - is answered beside the box,
 * and the key stored is kept.
 */
export function LicenseKeyForm({ license, canWrite }: Props) {
    const headingId = useId();
    const keyFieldId = useId();
    const keyErrorId = useId();
    const openAlert = useGlobalAlertDialogState(state => state.open);

    const {
        register,
        handleSubmit,
        reset,
        setError,
        formState: { errors },
    } = useForm<LicenseKeyFormInput, unknown, LicenseKeyFormOutput>({
        defaultValues: { key: "" },
        resolver: zodResolver(LicenseKeyFormSchema),
        mode: "onSubmit",
    });

    const { mutate: installKey, isPending } = LicenseCommands.useInstallKey({
        onSuccess: response => {
            reset({ key: "" });
            toast.success(`License installed: ${licenseInForceSummary(response.data)}`);
        },
        onError: error => {
            // Any other failure was toasted by the API hook.
            const message = licenseKeyRefusalMessage(error);
            if (message !== undefined) {
                setError("key", { type: "server", message }, { shouldFocus: true });
            }
        },
    });

    function onValid({ key }: LicenseKeyFormOutput) {
        if (!canWrite) {
            return;
        }

        // The server takes any key that can be in force here, a lower plan
        // included, and nothing can be undone but by pasting the old key again.
        if (!keyInForce(license)) {
            installKey({ key });
            return;
        }

        openAlert({
            props: {
                type: "warning",
                title: "Replace the license key?",
                description: `This key replaces the ${licensePlanName(license.plan)} key in force, whatever its plan.`,
                actionText: "Replace key",
                onAction: () => {
                    installKey({ key });
                },
            },
        });
    }

    return (
        <section
            aria-labelledby={headingId}
            className="flex flex-col gap-4"
        >
            <SectionHeader>
                <h2 id={headingId}>Install a key</h2>
            </SectionHeader>
            <form
                onSubmit={event => {
                    event.preventDefault();
                    void handleSubmit(onValid)(event);
                }}
                className="px-3"
            >
                <fieldset
                    disabled={!canWrite}
                    className="m-0 flex min-w-0 flex-col gap-6 border-0 p-0"
                >
                    <InfoBlock
                        titleWidth={220}
                        title={
                            <LabelWithInfo
                                label="License key"
                                htmlFor={keyFieldId}
                                content="Checked here, offline: its signatures, the installation it is for, and its dates. A key that cannot be in force on this installation is refused, and the key stored is kept."
                            />
                        }
                    >
                        <div className="flex max-w-[640px] flex-col items-start gap-2">
                            <Textarea
                                id={keyFieldId}
                                rows={4}
                                className="font-mono text-xs break-all"
                                placeholder="Paste the key the portal shows, or the one it emailed you"
                                autoComplete="off"
                                autoCapitalize="off"
                                autoCorrect="off"
                                spellCheck={false}
                                aria-invalid={errors.key ? true : undefined}
                                aria-describedby={errors.key ? keyErrorId : undefined}
                                {...register("key")}
                            />
                            <FieldError
                                id={keyErrorId}
                                // The server's reasons can run to two lines: what was refused, then why.
                                className="whitespace-pre-wrap"
                                errors={[errors.key]}
                            />
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
                                        Install key
                                    </Button>
                                )}
                            </PermissionTooltipAction>
                        </div>
                    </InfoBlock>
                </fieldset>
            </form>
        </section>
    );
}
