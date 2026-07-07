import { useState } from "react";

import { toast } from "sonner";
import { ProjectEmailCommands } from "~/projects/data/commands";
import { ProjectEmailQueries } from "~/projects/data/queries";
import { EmailCommands } from "~/settings/data/commands";
import { EmailQueries } from "~/settings/data/queries";
import type { SettingEmail } from "~/settings/domain";
import {
    CreateOrEditEmailAccountForm,
    TestSendMailDialog,
} from "~/settings/module-shared/components/email-account-form";
import type {
    CreateOrEditEmailAccountFormInput,
    CreateOrEditEmailAccountFormOutput,
    TestSendMailFormOutput,
} from "~/settings/module-shared/components/email-account-form";
import { useSettingsScopePermissions } from "~/settings/module-shared/hooks";
import { SettingsFormRouteHeader } from "~/settings/module-shared/components/settings-form-route-header";

import { AppLoader } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";
import { EEmailKind } from "@application/shared/enums";
import { useAppNavigate } from "@application/shared/hooks/router";

import type { EmailAccountTableScope } from "../email-account-table";

type EmailAccountFormRouteMode = "create" | "edit";

const FIELD_MAPPING_NAMES = [
    "fromAddress",
    "fromName",
    "toAddress",
    "toAddresses",
    "subject",
    "content",
    "password",
] as const;

export function EmailAccountFormRoute({ mode, scope, emailAccountId }: Props) {
    const [hasChanges, setHasChanges] = useState(false);
    const [saveRevision, setSaveRevision] = useState(0);
    const [testStatus, setTestStatus] = useState<"idle" | "succeeded" | "failed">("idle");
    const [testDialogOpen, setTestDialogOpen] = useState(false);
    const [testAccountValues, setTestAccountValues] = useState<CreateOrEditEmailAccountFormOutput | null>(null);
    const { canWrite } = useSettingsScopePermissions(scope);
    const { navigate } = useAppNavigate();

    const listRoute = getEmailAccountListRoute(scope);
    const isEditMode = mode === "edit";
    const detailId = isEditMode ? (emailAccountId ?? "") : "";

    function navigateToList() {
        navigate.modules(listRoute, { ignorePrevPath: true });
    }

    function markSaved() {
        setHasChanges(false);
        setSaveRevision(revision => revision + 1);
    }

    const { mutate: createSettingEmailAccount, isPending: isCreatingSetting } = EmailCommands.useCreateOne({
        onSuccess: () => {
            toast.success("Email account created successfully");
            markSaved();
        },
    });
    const { mutate: updateSettingEmailAccount, isPending: isUpdatingSetting } = EmailCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Email account updated successfully");
            markSaved();
        },
    });
    const { mutate: createProjectEmailAccount, isPending: isCreatingProject } = ProjectEmailCommands.useCreateOne({
        onSuccess: () => {
            toast.success("Project Email account created successfully");
            markSaved();
        },
    });
    const { mutate: updateProjectEmailAccount, isPending: isUpdatingProject } = ProjectEmailCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Project Email account updated successfully");
            markSaved();
        },
    });
    const { mutate: testSendMail, isPending: isTesting } = EmailCommands.useTestSendMail({
        onSuccess: () => {
            setTestStatus("succeeded");
        },
        onError: () => {
            setTestStatus("failed");
        },
    });

    const settingDetailQuery = EmailQueries.useFindOneById(
        { id: detailId },
        { enabled: isEditMode && scope.type === "settings" },
    );
    const projectDetailQuery = ProjectEmailQueries.useFindOneById(
        {
            projectID: scope.type === "project" ? scope.projectId : "",
            id: detailId,
        },
        { enabled: isEditMode && scope.type === "project" },
    );
    const detailQuery = scope.type === "project" ? projectDetailQuery : settingDetailQuery;
    const emailAccount = detailQuery.data?.data;
    const readOnlyInherited = scope.type === "project" && emailAccount?.inherited === true;

    function createPayload(values: CreateOrEditEmailAccountFormOutput) {
        const isSmtp = values.kind === EEmailKind.SMTP;
        const headers = toRecord(values.headers);

        return {
            availableInProjects: scope.type === "project" ? false : values.availableInProjects,
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
        const payload = createPayload(values);

        if (isEditMode && emailAccount) {
            const updatePayload = { ...payload, updateVer: emailAccount.updateVer };

            if (scope.type === "project") {
                updateProjectEmailAccount({
                    projectID: scope.projectId,
                    id: emailAccount.id,
                    payload: updatePayload,
                });
                return;
            }

            updateSettingEmailAccount({ id: emailAccount.id, payload: updatePayload });
            return;
        }

        if (scope.type === "project") {
            createProjectEmailAccount({ projectID: scope.projectId, payload });
            return;
        }

        createSettingEmailAccount({ payload });
    }

    function openTestSendMailDialog(values: CreateOrEditEmailAccountFormOutput) {
        setTestStatus("idle");
        setTestAccountValues(values);
        setTestDialogOpen(true);
    }

    function onTestSendMail(values: TestSendMailFormOutput) {
        if (!testAccountValues) {
            return;
        }

        setTestStatus("idle");
        const {
            availableInProjects: _availableInProjects,
            default: _default,
            ...payload
        } = createPayload(testAccountValues);

        testSendMail({
            payload: {
                ...payload,
                testRecipient: values.testRecipient,
                testSubject: values.testSubject,
                testContent: values.testContent,
            },
        });
    }

    function handleClose() {
        if (isPending || isTesting) return;
        if (
            !readOnlyInherited &&
            canWrite &&
            hasChanges &&
            !window.confirm("Are you sure you want to close without saving changes?")
        )
            return;

        navigateToList();
    }

    function handleTestDialogOpenChange(nextOpen: boolean) {
        if (isTesting) {
            return;
        }

        setTestDialogOpen(nextOpen);

        if (!nextOpen) {
            setTestStatus("idle");
            setTestAccountValues(null);
        }
    }

    const isPending = isCreatingSetting || isUpdatingSetting || isCreatingProject || isUpdatingProject;
    const isDetailLoading = isEditMode && detailQuery.isFetching;
    const initialValues: Partial<CreateOrEditEmailAccountFormInput> | undefined = emailAccount
        ? toInitialValues(emailAccount)
        : undefined;
    const shouldRenderForm = mode === "create" || initialValues;
    const title = mode === "create" ? "Create Email Account" : "Edit Email Account";

    return (
        <>
            <div className="flex w-full flex-col overflow-hidden">
                <SettingsFormRouteHeader title={title} />

                {isDetailLoading && (
                    <div className="flex min-h-[220px] items-center justify-center">
                        <AppLoader />
                    </div>
                )}

                {!isDetailLoading && shouldRenderForm && (
                    <CreateOrEditEmailAccountForm
                        isPending={isPending}
                        onSubmit={onSubmit}
                        onOpenTestSendMail={openTestSendMailDialog}
                        onHasChanges={setHasChanges}
                        savedVersion={saveRevision}
                        initialValues={initialValues}
                        showAvailableInProjects={scope.type === "settings"}
                        readOnlyInherited={readOnlyInherited}
                        readOnly={!canWrite}
                        onClose={handleClose}
                    />
                )}
            </div>
            <TestSendMailDialog
                open={testDialogOpen}
                isPending={isTesting}
                testStatus={testStatus}
                onOpenChange={handleTestDialogOpenChange}
                onSubmit={onTestSendMail}
            />
        </>
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

function getEmailAccountListRoute(scope: EmailAccountTableScope) {
    if (scope.type === "project") {
        return ROUTE.projects.single.providerConfiguration.emailAccounts.$route(scope.projectId);
    }

    return ROUTE.settings.emailAccounts.$route;
}

interface Props {
    mode: EmailAccountFormRouteMode;
    scope: EmailAccountTableScope;
    emailAccountId?: string;
}
