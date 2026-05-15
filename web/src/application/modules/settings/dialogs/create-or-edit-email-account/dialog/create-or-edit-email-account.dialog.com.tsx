import { useEffect, useState } from "react";

import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { toast } from "sonner";
import { ProjectEmailCommands } from "~/projects/data/commands";
import { ProjectEmailQueries } from "~/projects/data/queries";
import { EmailCommands } from "~/settings/data/commands";
import { EmailQueries } from "~/settings/data/queries";
import type { SettingEmail } from "~/settings/domain";

import { EEmailKind } from "@application/shared/enums";

import { CreateOrEditEmailAccountForm } from "../form";
import { useCreateOrEditEmailAccountDialogState } from "../hooks";
import type { CreateOrEditEmailAccountFormOutput } from "../schemas";

const FIELD_MAPPING_NAMES = [
    "fromAddress",
    "fromName",
    "toAddress",
    "toAddresses",
    "subject",
    "content",
    "password",
] as const;

export function CreateOrEditEmailAccountDialog() {
    const {
        state,
        props: dialogOptions,
        close: closeDialog,
        clear: clearDialog,
    } = useCreateOrEditEmailAccountDialogState();
    const [hasChanges, setHasChanges] = useState(false);
    const [testStatus, setTestStatus] = useState<"idle" | "succeeded" | "failed">("idle");

    const { mutate: createSettingEmailAccount, isPending: isCreatingSetting } = EmailCommands.useCreateOne({
        onSuccess: () => {
            toast.success("Email account created successfully");
            closeDialog();
            dialogOptions?.onSuccess?.();
        },
        onError: dialogOptions?.onError,
    });
    const { mutate: updateSettingEmailAccount, isPending: isUpdatingSetting } = EmailCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Email account updated successfully");
            closeDialog();
            dialogOptions?.onSuccess?.();
        },
        onError: dialogOptions?.onError,
    });
    const { mutate: createProjectEmailAccount, isPending: isCreatingProject } = ProjectEmailCommands.useCreateOne({
        onSuccess: () => {
            toast.success("Project Email account created successfully");
            closeDialog();
            dialogOptions?.onSuccess?.();
        },
        onError: dialogOptions?.onError,
    });
    const { mutate: updateProjectEmailAccount, isPending: isUpdatingProject } = ProjectEmailCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Project Email account updated successfully");
            closeDialog();
            dialogOptions?.onSuccess?.();
        },
        onError: dialogOptions?.onError,
    });
    const { mutate: testSendMail, isPending: isTesting } = EmailCommands.useTestSendMail({
        onSuccess: () => {
            setTestStatus("succeeded");
        },
        onError: () => {
            setTestStatus("failed");
        },
    });

    useEffect(() => {
        if (state.mode === "closed") {
            setHasChanges(false);
            setTestStatus("idle");
            clearDialog();
        }
    }, [clearDialog, state.mode]);

    const detailId = state.mode === "edit" ? state.id : "";
    const settingDetailQuery = EmailQueries.useFindOneById(
        { id: detailId },
        {
            enabled: state.mode === "edit" && state.scope.type === "settings",
        },
    );
    const projectDetailQuery = ProjectEmailQueries.useFindOneById(
        {
            projectID: state.mode === "edit" && state.scope.type === "project" ? state.scope.projectId : "",
            id: detailId,
        },
        {
            enabled: state.mode === "edit" && state.scope.type === "project",
        },
    );
    const detailQuery =
        state.mode === "edit" && state.scope.type === "project" ? projectDetailQuery : settingDetailQuery;
    const emailAccount = detailQuery.data?.data;

    function createPayload(values: CreateOrEditEmailAccountFormOutput) {
        const isSmtp = values.kind === EEmailKind.SMTP;
        const headers = toRecord(values.headers);

        return {
            availableInProjects:
                state.mode !== "closed" && state.scope.type === "project" ? false : values.availableInProjects,
            default: values.default,
            name: values.name,
            kind: values.kind,
            smtp: isSmtp
                ? {
                      host: values.smtpHost,
                      port: values.smtpPort,
                      username: values.smtpUsername,
                      displayName: values.smtpDisplayName,
                      password: values.smtpPassword,
                      ssl: false,
                  }
                : null,
            http: isSmtp
                ? null
                : {
                      endpoint: values.httpEndpoint,
                      method: values.httpMethod,
                      username: values.httpUsername,
                      displayName: values.httpDisplayName,
                      password: values.httpPassword,
                      contentType: values.httpContentType,
                      headers,
                      fieldMapping: toFieldMapping(values.fieldMapping),
                  },
        };
    }

    function onSubmit(values: CreateOrEditEmailAccountFormOutput) {
        if (state.mode === "closed") {
            return;
        }

        const payload = createPayload(values);

        if (state.mode === "edit" && emailAccount) {
            const updatePayload = {
                ...payload,
                updateVer: emailAccount.updateVer,
            };

            if (state.scope.type === "project") {
                updateProjectEmailAccount({
                    projectID: state.scope.projectId,
                    id: emailAccount.id,
                    payload: updatePayload,
                });
                return;
            }

            updateSettingEmailAccount({
                id: emailAccount.id,
                payload: updatePayload,
            });
            return;
        }

        if (state.scope.type === "project") {
            createProjectEmailAccount({
                projectID: state.scope.projectId,
                payload,
            });
            return;
        }

        createSettingEmailAccount({ payload });
    }

    function onTestSendMail(values: CreateOrEditEmailAccountFormOutput) {
        setTestStatus("idle");
        const { availableInProjects: _availableInProjects, default: _default, ...payload } = createPayload(values);
        testSendMail({
            payload: {
                ...payload,
                testRecipient: "",
                testSubject: "test subject",
                testContent: "test content",
            },
        });
    }

    function handleClose() {
        if (isPending) {
            return;
        }

        if (hasChanges && !window.confirm("Are you sure you want to close without saving changes?")) {
            return;
        }

        closeDialog();
        dialogOptions?.onClose?.();
    }

    const open = state.mode !== "closed";
    const isPending = isCreatingSetting || isUpdatingSetting || isCreatingProject || isUpdatingProject;
    const showAvailableInProjects = state.mode !== "closed" && state.scope.type === "settings";
    const initialValues = emailAccount ? toInitialValues(emailAccount) : undefined;
    const isDetailLoading = state.mode === "edit" && detailQuery.isFetching;

    return (
        <Dialog
            open={open}
            onOpenChange={handleClose}
        >
            <DialogContent className="min-w-[390px] w-[760px] max-h-[90vh] overflow-y-auto">
                <DialogHeader>
                    <DialogTitle>Create or update an e-mail account</DialogTitle>
                </DialogHeader>
                {isDetailLoading && <div className="py-8 text-sm text-muted-foreground">Loading email account...</div>}
                {state.mode !== "closed" && !isDetailLoading && (state.mode === "open" || initialValues) && (
                    <CreateOrEditEmailAccountForm
                        isPending={isPending}
                        isTesting={isTesting}
                        testStatus={testStatus}
                        onSubmit={onSubmit}
                        onTestSendMail={onTestSendMail}
                        onHasChanges={setHasChanges}
                        initialValues={initialValues}
                        showAvailableInProjects={showAvailableInProjects}
                    />
                )}
            </DialogContent>
        </Dialog>
    );
}

function toKeyValueList(map?: object | null) {
    return Object.entries(map ?? {}).map(([key, value]) => ({ key, value: String(value ?? "") }));
}

function toRecord(rows: { key: string; value: string }[]) {
    return rows.reduce<Record<string, string>>((acc, item) => {
        const key = item.key.trim();
        if (!key) {
            return acc;
        }

        return {
            ...acc,
            [key]: item.value.trim(),
        };
    }, {});
}

function toFieldMapping(rows: { key: string; value: string }[]) {
    const mapping: Record<(typeof FIELD_MAPPING_NAMES)[number], string> = {
        fromAddress: "",
        fromName: "",
        toAddress: "",
        toAddresses: "",
        subject: "",
        content: "",
        password: "",
    };

    return rows.reduce((acc, item) => {
        const key = item.key.trim();
        if (!FIELD_MAPPING_NAMES.includes(key as (typeof FIELD_MAPPING_NAMES)[number])) {
            return acc;
        }

        return {
            ...acc,
            [key]: item.value.trim(),
        };
    }, mapping);
}

function toInitialValues(emailAccount: SettingEmail): Partial<CreateOrEditEmailAccountFormOutput> {
    return {
        name: emailAccount.name,
        kind: emailAccount.kind,
        smtpHost: emailAccount.smtp?.host ?? "",
        smtpPort: emailAccount.smtp?.port ?? 587,
        smtpUsername: emailAccount.smtp?.username ?? "",
        smtpDisplayName: emailAccount.smtp?.displayName ?? "",
        smtpPassword: emailAccount.smtp?.password ?? "",
        httpEndpoint: emailAccount.http?.endpoint ?? "",
        httpMethod:
            emailAccount.http?.method === "GET" || emailAccount.http?.method === "PUT"
                ? emailAccount.http.method
                : "POST",
        httpUsername: emailAccount.http?.username ?? "",
        httpDisplayName: emailAccount.http?.displayName ?? "",
        httpPassword: emailAccount.http?.password ?? "",
        httpContentType: emailAccount.http?.contentType ?? "application/json",
        headers: toKeyValueList(emailAccount.http?.headers),
        fieldMapping: toKeyValueList(emailAccount.http?.fieldMapping),
        availableInProjects: emailAccount.availableInProjects ?? false,
        default: emailAccount.default ?? false,
    };
}
