/**
 * How a request actually reached HivePaaS.
 *
 * The number of proxies in front cannot be read off a deployment diagram:
 * whether a proxy adds itself to X-Forwarded-For is that proxy's own decision, so
 * the length of the chain is the only thing that settles it. This is the
 * measurement, taken from the request that asked for it - which, when the browser
 * asks, is the same path real traffic takes.
 */
export type HivePaaSRequestInfo = {
    /** The peer HivePaaS is talking to - the proxy, or Traefik. */
    remoteAddr: string;
    /** What HivePaaS currently believes the caller's address is. */
    clientIp: string;
    /** The X-Forwarded-For chain as it arrived, in order. The caller is first. */
    forwardedFor: string[];
    /** The proxy-related headers that arrived, verbatim. */
    headers: Record<string, string>;
    /** What to put in Proxy Hops, derived from the chain. */
    suggestedProxyHops: number;
    /** What the numbers mean, and how to tell the reading was taken wrongly. */
    explanation: string;
};
