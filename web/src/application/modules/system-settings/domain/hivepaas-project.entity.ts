/**
 * The project HivePaaS runs in: its own apps - the backend, the worker, the
 * proxy, the agent - and the system apps it provisions. The projects list leaves
 * it out; an admin reaches it from System › HivePaaS › Actions.
 */
export type HivePaaSProject = {
    id: string;
    name: string;
    key: string;
};
