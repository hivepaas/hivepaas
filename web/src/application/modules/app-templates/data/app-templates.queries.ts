import {
    type InfiniteData,
    type UseInfiniteQueryResult,
    type UseMutationOptions,
    type UseMutationResult,
    type UseQueryResult,
    useInfiniteQuery,
    useMutation,
    useQuery,
    useQueryClient,
} from "@tanstack/react-query";
import { QK } from "~/projects/data/constants";

import {
    type AppTemplateCatalog,
    type AppTemplateDetail,
    type AppTemplateImageTagsResponse,
    type CreateAppFromTemplateReq,
    type CreateAppFromTemplateResp,
    type EnvApp,
    type GetAppTemplateImageTagsParams,
    type ListAppTemplatesFilter,
    type ListAppTemplatesResponse,
    type ListEnvAppsParams,
    type PreflightResult,
    appTemplatesApi,
} from "../api";

export const APP_TEMPLATES_QUERY_KEYS = {
    catalog: () => ["app-templates", "catalog"] as const,
    list: (filter?: ListAppTemplatesFilter) => ["app-templates", "list", filter] as const,
    detail: (name: string) => ["app-templates", "detail", name] as const,
    imageTags: (params: GetAppTemplateImageTagsParams) =>
        ["app-templates", "image-tags", params.templateName, params.version, params.variant] as const,
    envApps: (params: ListEnvAppsParams) =>
        ["app-templates", "env-apps", params.projectID, params.projectEnv, params.search ?? ""] as const,
};

export const PAGE_LIMIT_DEFAULT = 50;

/**
 * Query hook to fetch the app template catalog (categories tree, tags, source, revision).
 */
export function useGetAppTemplateCatalog(): UseQueryResult<AppTemplateCatalog> {
    return useQuery<AppTemplateCatalog>({
        queryKey: APP_TEMPLATES_QUERY_KEYS.catalog(),
        queryFn: ({ signal }) => appTemplatesApi.getCatalog(signal),
        staleTime: 5 * 60 * 1000,
    });
}

/**
 * Infinite query hook to fetch templates 50 items at a time with "Load More" pagination.
 */
export function useListAppTemplatesInfinite(
    filter: Omit<ListAppTemplatesFilter, "pageOffset" | "pageLimit">,
): UseInfiniteQueryResult<InfiniteData<ListAppTemplatesResponse>> {
    return useInfiniteQuery<ListAppTemplatesResponse>({
        queryKey: APP_TEMPLATES_QUERY_KEYS.list(filter),
        queryFn: ({ pageParam, signal }) => {
            const offset = typeof pageParam === "number" ? pageParam : 0;
            return appTemplatesApi.listTemplates(
                {
                    ...filter,
                    pageOffset: offset,
                    pageLimit: PAGE_LIMIT_DEFAULT,
                },
                signal,
            );
        },
        initialPageParam: 0,
        getNextPageParam: (lastPage: ListAppTemplatesResponse) => {
            const pageOffset = lastPage.meta.page?.offset ?? lastPage.meta.pageOffset ?? 0;
            const count = lastPage.data.length;
            const total = lastPage.meta.page?.total ?? lastPage.meta.total ?? 0;

            const nextOffset = pageOffset + count;
            if (nextOffset < total) {
                return nextOffset;
            }
            return undefined;
        },
    });
}

/**
 * Query hook to fetch detailed information for a specific template.
 */
export function useGetAppTemplate(name?: string): UseQueryResult<AppTemplateDetail> {
    return useQuery<AppTemplateDetail>({
        queryKey: APP_TEMPLATES_QUERY_KEYS.detail(name ?? ""),
        queryFn: ({ signal }) => {
            if (!name) {
                throw new Error("Template name is required");
            }
            return appTemplatesApi.getTemplate(name, signal);
        },
        enabled: Boolean(name),
        staleTime: 5 * 60 * 1000,
    });
}

/**
 * Query hook to fetch scanned registry image tags for a template version/variant.
 */
export function useGetAppTemplateImageTags(
    params: GetAppTemplateImageTagsParams,
    options?: { enabled?: boolean },
): UseQueryResult<AppTemplateImageTagsResponse> {
    return useQuery<AppTemplateImageTagsResponse>({
        queryKey: APP_TEMPLATES_QUERY_KEYS.imageTags(params),
        queryFn: ({ signal }) => appTemplatesApi.getImageTags(params, signal),
        enabled: options?.enabled ?? Boolean(params.templateName),
        staleTime: 60 * 1000,
    });
}

/**
 * Mutation hook asking what creating a request would run into. It writes nothing,
 * and its findings are what the deploy dialog warns about before it creates.
 */
export function usePreflightAppFromTemplate(
    options?: Omit<UseMutationOptions<PreflightResult, Error, CreateAppFromTemplateReq>, "mutationFn">,
): UseMutationResult<PreflightResult, Error, CreateAppFromTemplateReq> {
    return useMutation({
        mutationFn: (req: CreateAppFromTemplateReq) => appTemplatesApi.preflightAppFromTemplate(req),
        ...options,
    });
}

/**
 * Mutation hook to create a new app from a template.
 */
export function useCreateAppFromTemplate(
    options?: Omit<UseMutationOptions<CreateAppFromTemplateResp, Error, CreateAppFromTemplateReq>, "mutationFn">,
): UseMutationResult<CreateAppFromTemplateResp, Error, CreateAppFromTemplateReq> {
    const queryClient = useQueryClient();
    const { onSuccess, ...restOptions } = options ?? {};

    return useMutation({
        mutationFn: (req: CreateAppFromTemplateReq) => appTemplatesApi.createAppFromTemplate(req),
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.apps.$.find-many-paginated"]],
            });
            if (onSuccess) {
                onSuccess(response, ...rest);
            }
        },
        ...restOptions,
    });
}

/**
 * Query hook for the environment's apps, for a parameter that names one. Held
 * briefly: the list is a picker's, and a deploy dialog is open for a minute.
 */
export function useListEnvApps(params: ListEnvAppsParams, enabled: boolean): UseQueryResult<EnvApp[]> {
    return useQuery<EnvApp[]>({
        queryKey: APP_TEMPLATES_QUERY_KEYS.envApps(params),
        queryFn: ({ signal }) => appTemplatesApi.listEnvApps(params, signal),
        enabled: enabled && Boolean(params.projectID) && Boolean(params.projectEnv),
        staleTime: 30 * 1000,
    });
}
