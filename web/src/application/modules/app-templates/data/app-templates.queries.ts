import {
    type InfiniteData,
    type UseInfiniteQueryResult,
    type UseQueryResult,
    useInfiniteQuery,
    useQuery,
} from "@tanstack/react-query";

import {
    type AppTemplateCatalog,
    type AppTemplateDetail,
    type ListAppTemplatesFilter,
    type ListAppTemplatesResponse,
    appTemplatesApi,
} from "../api";

export const APP_TEMPLATES_QUERY_KEYS = {
    catalog: () => ["app-templates", "catalog"] as const,
    list: (filter?: ListAppTemplatesFilter) => ["app-templates", "list", filter] as const,
    detail: (name: string) => ["app-templates", "detail", name] as const,
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
            const nextOffset = lastPage.meta.pageOffset + lastPage.meta.count;
            if (nextOffset < lastPage.meta.total) {
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
