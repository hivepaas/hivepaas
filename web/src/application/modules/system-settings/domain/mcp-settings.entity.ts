/**
 * Whether HivePaaS serves the Model Context Protocol, at <API base path>/mcp.
 * Off, the path answers 404.
 */
export interface McpSettings {
    enabled: boolean;
    /** What the tools that change things will be gated on. Read by the server from phase 2 on. */
    allowWrite: boolean;
    updateVer: number;
}
