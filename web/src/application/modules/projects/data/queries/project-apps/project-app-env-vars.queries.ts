import { type UseQueryOptions, useQuery } from "@tanstack/react-query";
import { useProjectAppEnvVarsApi } from "~/projects/api/hooks/project-apps";
import type {
    ProjectAppEnvVars_FindLinkSuggestions_Req,
    ProjectAppEnvVars_FindLinkSuggestions_Res,
    ProjectAppEnvVars_FindLinkTargets_Req,
    ProjectAppEnvVars_FindLinkTargets_Res,
    ProjectAppEnvVars_FindOne_Req,
    ProjectAppEnvVars_FindOne_Res,
} from "~/projects/api/services";
import { QK } from "~/projects/data/constants";

/**
 * Find one project app env vars query
 */
type FindOneReq = ProjectAppEnvVars_FindOne_Req["data"];
type FindOneRes = ProjectAppEnvVars_FindOne_Res;

type FindOneOptions = Omit<UseQueryOptions<FindOneRes>, "queryKey" | "queryFn">;

function useFindOne(request: FindOneReq, options: FindOneOptions = {}) {
    const { queries } = useProjectAppEnvVarsApi();

    return useQuery({
        queryKey: [QK["projects.apps.env-vars.$.find-one"], request],
        queryFn: ({ signal }) => queries.findOne(request, signal),
        ...options,
    });
}

type FindLinkTargetsReq = ProjectAppEnvVars_FindLinkTargets_Req["data"];
type FindLinkTargetsOptions = Omit<UseQueryOptions<ProjectAppEnvVars_FindLinkTargets_Res>, "queryKey" | "queryFn">;

function useFindLinkTargets(request: FindLinkTargetsReq, options: FindLinkTargetsOptions = {}) {
    const { queries } = useProjectAppEnvVarsApi();

    return useQuery({
        queryKey: [QK["projects.apps.env-vars.$.find-link-targets"], request],
        queryFn: ({ signal }) => queries.findLinkTargets(request, signal),
        ...options,
    });
}

type FindLinkSuggestionsReq = ProjectAppEnvVars_FindLinkSuggestions_Req["data"];
type FindLinkSuggestionsOptions = Omit<
    UseQueryOptions<ProjectAppEnvVars_FindLinkSuggestions_Res>,
    "queryKey" | "queryFn"
>;

function useFindLinkSuggestions(request: FindLinkSuggestionsReq, options: FindLinkSuggestionsOptions = {}) {
    const { queries } = useProjectAppEnvVarsApi();

    return useQuery({
        queryKey: [QK["projects.apps.env-vars.$.find-link-suggestions"], request],
        queryFn: ({ signal }) => queries.findLinkSuggestions(request, signal),
        ...options,
    });
}

export const ProjectAppEnvVarsQueries = Object.freeze({
    useFindOne,
    useFindLinkTargets,
    useFindLinkSuggestions,
});
