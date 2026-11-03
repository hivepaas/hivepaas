import { use, useMemo } from "react";

import { type Result, match } from "oxide.ts";

import { useApiErrorNotifications } from "@infrastructure/api";

import type {
    AppManaged_AddReplica_Req,
    AppManaged_CreateRestorePoint_Req,
    AppManaged_DeleteRestorePoint_Req,
    AppManaged_Promote_Req,
    AppManaged_RefreshCredentials_Req,
    AppManaged_RemoveReplica_Req,
    AppManaged_Restore_Req,
    AppManaged_Switch_Req,
    AppManaged_Unmanage_Req,
    AppManaged_UpdateRestorePoint_Req,
    AppManaged_UpdateSettings_Req,
    AppManaged_UpgradeScripts_Req,
} from "../../../api/services";
import { ProjectsApiContext } from "../../api-context/projects.api.context";

type AppScope = { projectID: string; env: string; appID: string };

function createHook() {
    return function useAppManagedApi() {
        const { api } = use(ProjectsApiContext);
        const { notifyError } = useApiErrorNotifications();
        const managed = api.projects.apps.managed.$;

        const queries = useMemo(() => {
            const read = <T>(result: Result<T, Error>): T =>
                match(result, {
                    Ok: _ => _,
                    Err: error => {
                        throw error;
                    },
                });
            return {
                findSettings: async (request: AppScope, signal?: AbortSignal) =>
                    read(await managed.findSettings({ data: request }, signal)),
                findStatus: async (request: AppScope, signal?: AbortSignal) =>
                    read(await managed.findStatus({ data: request }, signal)),
                findSwitch: async (request: AppScope, signal?: AbortSignal) =>
                    read(await managed.findSwitch({ data: request }, signal)),
                findScripts: async (request: AppScope, signal?: AbortSignal) =>
                    read(await managed.findScripts({ data: request }, signal)),
            };
        }, [managed]);

        const mutations = useMemo(() => {
            const write =
                (message: string) =>
                <T>(result: Result<T, Error>): T =>
                    match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message, error });
                            throw error;
                        },
                    });
            return {
                updateSettings: async (request: AppManaged_UpdateSettings_Req["data"]) =>
                    write("Failed to update the database's settings")(await managed.updateSettings({ data: request })),
                refreshCredentials: async (request: AppManaged_RefreshCredentials_Req["data"]) =>
                    write("Failed to refresh the archive's credentials")(
                        await managed.refreshCredentials({ data: request }),
                    ),
                unmanage: async (request: AppManaged_Unmanage_Req["data"]) =>
                    write("Failed to unmanage the database")(await managed.unmanage({ data: request })),
                restore: async (request: AppManaged_Restore_Req["data"]) =>
                    write("Failed to restore the database")(await managed.restore({ data: request })),
                addReplica: async (request: AppManaged_AddReplica_Req["data"]) =>
                    write("Failed to add the replica")(await managed.addReplica({ data: request })),
                promote: async (request: AppManaged_Promote_Req["data"]) =>
                    write("Failed to promote the replica")(await managed.promote({ data: request })),
                removeReplica: async (request: AppManaged_RemoveReplica_Req["data"]) =>
                    write("Failed to remove the replica")(await managed.removeReplica({ data: request })),
                switchConsumers: async (request: AppManaged_Switch_Req["data"]) =>
                    write("Failed to switch the consumers")(await managed.switchConsumers({ data: request })),
                createRestorePoint: async (request: AppManaged_CreateRestorePoint_Req["data"]) =>
                    write("Failed to make the restore point")(await managed.createRestorePoint({ data: request })),
                updateRestorePoint: async (request: AppManaged_UpdateRestorePoint_Req["data"]) =>
                    write("Failed to update the restore point")(await managed.updateRestorePoint({ data: request })),
                upgradeScripts: async (request: AppManaged_UpgradeScripts_Req["data"]) =>
                    write("Failed to update the database's scripts")(await managed.upgradeScripts({ data: request })),
                deleteRestorePoint: async (request: AppManaged_DeleteRestorePoint_Req["data"]) =>
                    write("Failed to delete the restore point")(await managed.deleteRestorePoint({ data: request })),
            };
        }, [managed, notifyError]);

        return { queries, mutations };
    };
}

export const useAppManagedApi = createHook();
