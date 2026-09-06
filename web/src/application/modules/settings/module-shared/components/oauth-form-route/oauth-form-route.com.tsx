import { useState } from "react";

import { toast } from "sonner";
import { OAuthCommands } from "~/settings/data/commands";
import { OAuthQueries } from "~/settings/data/queries";
import type { SettingOAuth } from "~/settings/domain";
import { CreateOrEditOAuthForm } from "~/settings/module-shared/components/oauth-form";
import type {
    CreateOrEditOAuthFormInput,
    CreateOrEditOAuthFormOutput,
} from "~/settings/module-shared/components/oauth-form";
import { SettingsFormRouteHeader } from "~/settings/module-shared/components/settings-form-route-header";
import { useSettingRevealSecrets } from "~/settings/module-shared/hooks";

import { AppLoader } from "@application/shared/components";
import { MODULE_IDS, ROUTE } from "@application/shared/constants";
import { EOAuthKind } from "@application/shared/enums";
import { useAppNavigate } from "@application/shared/hooks/router";
import { useConditionalModule } from "@application/shared/permissions";

import { RevealSecretsProvider } from "@/components/ui/input-password";

import { ConfirmRevealSecretsDialog } from "../confirm-reveal-secrets-dialog";
import { RevealSecretsButton } from "../reveal-secrets-button";

type OAuthFormRouteMode = "create" | "edit";

export function OAuthFormRoute({ mode, oauthId }: Props) {
    const [hasChanges, setHasChanges] = useState(false);
    const [saveRevision, setSaveRevision] = useState(0);
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.Settings });
    const { navigate } = useAppNavigate();

    const isEditMode = mode === "edit";
    const detailId = isEditMode ? (oauthId ?? "") : "";

    function navigateToList() {
        navigate.modules(ROUTE.settings.oauth.$route, { ignorePrevPath: true });
    }

    function markSaved() {
        setHasChanges(false);
        setSaveRevision(revision => revision + 1);
    }

    const { mutate: createOAuth, isPending: isCreating } = OAuthCommands.useCreateOne({
        onSuccess: () => {
            toast.success("OAuth created successfully");
            markSaved();
        },
    });
    const { mutate: updateOAuth, isPending: isUpdating } = OAuthCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("OAuth updated successfully");
            markSaved();
        },
    });

    const detailQuery = OAuthQueries.useFindOneById({ id: detailId }, { enabled: isEditMode });
    const oauth = detailQuery.data?.data;

    function createPayload(values: CreateOrEditOAuthFormOutput) {
        return {
            default: values.default,
            kind: values.kind,
            name: values.name,
            organization: values.organization,
            clientId: values.clientId,
            clientSecret: values.clientSecret,
            authURL: values.authURL,
            tokenURL: values.tokenURL,
            profileURL: values.profileURL,
            autoDiscoveryURL: values.kind === EOAuthKind.OpenIDConnect ? values.autoDiscoveryURL : "",
            scopes: values.scopes
                .split(",")
                .map(item => item.trim())
                .filter(Boolean),
        };
    }

    function onSubmit(values: CreateOrEditOAuthFormOutput) {
        if (!canWrite) {
            return;
        }
        const payload = createPayload(values);

        if (isEditMode && oauth) {
            updateOAuth({ id: oauth.id, payload: { ...payload, updateVer: oauth.updateVer } });
            return;
        }

        createOAuth({ payload });
    }

    function handleClose() {
        if (isPending) {
            return;
        }
        if (canWrite && hasChanges && !window.confirm("Are you sure you want to close without saving changes?")) {
            return;
        }

        navigateToList();
    }

    const {
        canShowRevealButton,
        isDialogOpen,
        setIsDialogOpen,
        isRevealing,
        isRevealed,
        revealedData,
        revealRevision,
        handleConfirmReveal,
    } = useSettingRevealSecrets<SettingOAuth>({
        settingType: "oauth",
        settingId: detailId,
        scope: { type: "settings" },
        mode,
    });

    const activeOAuth = revealedData ?? oauth;

    const isPending = isCreating || isUpdating;
    const isDetailLoading = isEditMode && detailQuery.isFetching;
    const initialValues: Partial<CreateOrEditOAuthFormInput> | undefined = activeOAuth
        ? {
              name: activeOAuth.name,
              kind: (activeOAuth.kind ?? EOAuthKind.Github) as EOAuthKind,
              organization: activeOAuth.organization,
              clientId: activeOAuth.clientId,
              clientSecret: activeOAuth.clientSecret,
              authURL: activeOAuth.authURL ?? "",
              tokenURL: activeOAuth.tokenURL ?? "",
              profileURL: activeOAuth.profileURL ?? "",
              autoDiscoveryURL: activeOAuth.autoDiscoveryURL ?? "",
              scopes: activeOAuth.scopes?.join(", ") ?? "",
              default: activeOAuth.default ?? false,
          }
        : undefined;
    const shouldRenderForm = mode === "create" || Boolean(initialValues);
    const title = mode === "create" ? "Create OAuth" : "Edit OAuth";

    return (
        <div className="flex w-full flex-col">
            <SettingsFormRouteHeader
                title={title}
                actions={
                    canShowRevealButton ? (
                        <RevealSecretsButton
                            onClick={() => {
                                setIsDialogOpen(true);
                            }}
                            isLoading={isRevealing}
                        />
                    ) : undefined
                }
            />

            <ConfirmRevealSecretsDialog
                open={isDialogOpen}
                onOpenChange={setIsDialogOpen}
                onConfirm={handleConfirmReveal}
                isPending={isRevealing}
            />

            {isDetailLoading && (
                <div className="flex min-h-[220px] items-center justify-center">
                    <AppLoader />
                </div>
            )}

            {!isDetailLoading && shouldRenderForm && (
                <RevealSecretsProvider value={{ isRevealed }}>
                    <CreateOrEditOAuthForm
                        key={`${detailId}-${revealRevision}`}
                        isPending={isPending}
                        onSubmit={onSubmit}
                        onHasChanges={setHasChanges}
                        savedVersion={saveRevision}
                        initialValues={initialValues}
                        disableProvider={isEditMode}
                        readOnly={!canWrite}
                        onClose={handleClose}
                    />
                </RevealSecretsProvider>
            )}
        </div>
    );
}

interface Props {
    mode: OAuthFormRouteMode;
    oauthId?: string;
}
