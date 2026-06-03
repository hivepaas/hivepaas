import { Button } from "@components/ui";
import type { ProjectImageBuildRepoCacheInfo } from "~/projects/domain";

import { InfoBlock } from "@application/shared/components";

interface RepoCacheInfoFieldsProps {
    hasQueried: boolean;
    cacheInfo?: ProjectImageBuildRepoCacheInfo;
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
        <InfoBlock title="Source Cache Info">
            <div className="flex min-h-9 flex-wrap items-center gap-10">
                {hasQueried && (
                    <>
                        <span>Total Files: {totalFiles}</span>
                        <span>Total Size: {cacheInfo?.totalSizeHR ?? "-"}</span>
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
