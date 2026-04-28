import React, { createContext, useCallback, useContext, useMemo, useState } from "react";

import type { AppStorageMount } from "~/projects/domain";

type StorageMountWithId = AppStorageMount & { _id: string };

interface StorageMountsContextValue {
    mounts: StorageMountWithId[];
    addMount: (mount: AppStorageMount) => void;
    updateMount: (id: string, mount: AppStorageMount) => void;
    removeMount: (id: string) => void;
    getMountById: (id: string) => StorageMountWithId | undefined;
}

const StorageMountsContext = createContext<StorageMountsContextValue | undefined>(undefined);

// eslint-disable-next-line react-refresh/only-export-components
export function useStorageMounts() {
    const context = useContext(StorageMountsContext);
    if (!context) {
        throw new Error("useStorageMounts must be used within StorageMountsProvider");
    }
    return context;
}

interface StorageMountsProviderProps {
    initialMounts?: AppStorageMount[];
    children: React.ReactNode;
}

export function StorageMountsProvider({ initialMounts = [], children }: StorageMountsProviderProps) {
    const [mounts, setMounts] = useState<StorageMountWithId[]>(() =>
        initialMounts.map((mount, index) => ({
            ...mount,
            _id: `mount-${Date.now()}-${index}`,
        })),
    );

    const addMount = useCallback((mount: AppStorageMount) => {
        const newMount: StorageMountWithId = {
            ...mount,
            _id: `mount-${Date.now()}-${Math.random()}`,
        };
        setMounts(prev => [...prev, newMount]);
    }, []);

    const updateMount = useCallback((id: string, mount: AppStorageMount) => {
        setMounts(prev => prev.map(m => (m._id === id ? { ...mount, _id: id } : m)));
    }, []);

    const removeMount = useCallback((id: string) => {
        setMounts(prev => prev.filter(m => m._id !== id));
    }, []);

    const getMountById = useCallback(
        (id: string) => {
            return mounts.find(m => m._id === id);
        },
        [mounts],
    );

    const value = useMemo(
        () => ({
            mounts,
            addMount,
            updateMount,
            removeMount,
            getMountById,
        }),
        [mounts, addMount, updateMount, removeMount, getMountById],
    );

    return <StorageMountsContext.Provider value={value}>{children}</StorageMountsContext.Provider>;
}
