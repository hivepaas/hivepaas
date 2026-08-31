import { Button } from "@components/ui";
import type { ImageBuildRepoCacheInfo } from "~/settings/domain";

import { InfoBlock } from "@application/shared/components";
import { getFriendlyDataSize } from "@application/shared/utils/data-size";

interface RepoCacheInfoFieldsProps {
    hasQueried: boolean;
    cacheInfo?: ImageBuildRepoCacheInfo;
    isQuerying: boolean;
    isClearing: boolean;
    readOnly?: boolean;
    onQuery: () => void;
    onClear: () => void;
}

export function RepoCacheInfoFields({
    hasQueried,
    cacheInfo,
    isQuerying,
    isClearing,
    readOnly = false,
    onQuery,
    onClear,
}: RepoCacheInfoFieldsProps) {
    const totalFiles = cacheInfo?.totalFiles ?? 0;
    const canClear = hasQueried && totalFiles > 0;

    return (
        <InfoBlock
            titleWidth={220}
            title="Source Cache Info"
        >
            <div className="flex min-h-9 flex-wrap items-center gap-10">
                {hasQueried && (
                    <>
                        <span>Total Files: {totalFiles}</span>
                        <span>Total Size: {getFriendlyDataSize(cacheInfo?.totalSizeBytes) || "-"}</span>
                    </>
                )}

                {canClear && (
                    <Button
                        type="button"
                        onClick={onClear}
                        disabled={readOnly}
                        isLoading={isClearing}
                    >
                        Clear Cache
                    </Button>
                )}

                {!hasQueried && (
                    <Button
                        type="button"
                        variant="link"
                        onClick={onQuery}
                        isLoading={isQuerying}
                        className="text-primary underline-offset-4 hover:underline px-0"
                    >
                        Query
                    </Button>
                )}
            </div>
        </InfoBlock>
    );
}
