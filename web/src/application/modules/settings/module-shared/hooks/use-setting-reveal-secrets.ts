import { useState } from "react";

import { toast } from "sonner";

import { CAPABILITY_IDS } from "@application/shared/constants";
import { useCapability } from "@application/shared/permissions/hooks/use-capability";

import { createApiClient } from "@infrastructure/api/client";

export interface UseSettingRevealSecretsOptions<T> {
    settingType: string;
    settingId?: string;
    scope: { type: "project"; projectId: string; env?: string } | { type: "settings" };
    isInherited?: boolean;
    mode: "create" | "edit";
    onSuccess?: (data: T) => void;
}

export interface UseSettingRevealSecretsResult<T> {
    canShowRevealButton: boolean;
    isDialogOpen: boolean;
    setIsDialogOpen: (open: boolean) => void;
    isRevealing: boolean;
    isRevealed: boolean;
    revealedData: T | null;
    revealRevision: number;
    handleConfirmReveal: () => Promise<void>;
}

export function useSettingRevealSecrets<T>({
    settingType,
    settingId,
    scope,
    isInherited = false,
    mode,
    onSuccess,
}: UseSettingRevealSecretsOptions<T>): UseSettingRevealSecretsResult<T> {
    const [isDialogOpen, setIsDialogOpen] = useState(false);
    const [isRevealing, setIsRevealing] = useState(false);
    const [isRevealed, setIsRevealed] = useState(false);
    const [revealedData, setRevealedData] = useState<T | null>(null);
    const [revealRevision, setRevealRevision] = useState(0);

    const { hasCapability: canRevealSecrets } = useCapability(CAPABILITY_IDS.SecretReveal);

    const canShowRevealButton = mode === "edit" && Boolean(settingId) && !isInherited && canRevealSecrets;

    const handleConfirmReveal = async () => {
        if (!settingId || !canShowRevealButton || isRevealing) {
            return;
        }

        setIsRevealing(true);
        try {
            const client = createApiClient();
            const projectPath =
                scope.type === "project" && scope.env && scope.env !== "all"
                    ? `/projects/${scope.projectId}/${encodeURIComponent(scope.env)}/${settingType}/${settingId}`
                    : `/projects/${scope.type === "project" ? scope.projectId : ""}/${settingType}/${settingId}`;
            const basePath = scope.type === "project" ? projectPath : `/settings/${settingType}/${settingId}`;

            const response = await client.v1.get<{ data: T }>(`${basePath}?revealSecrets=true`);
            const { data } = response.data;
            if (data) {
                setRevealedData(data);
                setIsRevealed(true);
                setRevealRevision(r => r + 1);
                setIsDialogOpen(false);
                onSuccess?.(data);
                toast.success("Secrets revealed successfully");
            }
        } catch (error) {
            console.error("Failed to reveal secrets:", error);
            toast.error("Failed to reveal secrets");
        } finally {
            setIsRevealing(false);
        }
    };

    return {
        canShowRevealButton,
        isDialogOpen,
        setIsDialogOpen,
        isRevealing,
        isRevealed,
        revealedData,
        revealRevision,
        handleConfirmReveal,
    };
}
