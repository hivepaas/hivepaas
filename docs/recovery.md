# Recovering from a bad configuration change

Most of what follows never has to be done by hand. HivePaaS puts the changes that
can lock you out on trial: they are applied immediately and undone automatically
unless somebody comes back through the new configuration and confirms them. This
page is for the cases where that safety net is not available - and for knowing
which case you are in.

## Which mechanism catches what

There are three, and they cover different failures.

| What went wrong | What catches it | How long |
| --- | --- | --- |
| Traefik or the app refuses to start under the new spec | Swarm's own `failure_action: rollback` | up to `start_period + retries × interval`, then the previous spec is restored |
| The service starts and passes its healthcheck, but serves nothing useful | Confirm-or-revert | the confirmation window, 3 minutes by default |
| Neither can run - the database is unreachable, or the node itself is down | Nothing. See [By hand](#by-hand) | - |

The middle row is the one worth understanding. Traefik's healthcheck asks
`/ping`, which reports that the HTTP server is alive and nothing else. A command
line that traefik parses and starts under, but which discovers no routers - a
`--providers.swarm.constraints` that matches nothing, a default middleware that
does not resolve - keeps answering `/ping` with 200 while every application
behind it returns 404. Swarm sees a healthy service. Only somebody failing to
come back tells us otherwise.

That is why confirming has to travel through the new configuration: the request
arriving at all is the proof. The undo does not travel that way - it runs inside
the HivePaaS app process, over the docker socket - so a traefik that is not
serving anything cannot stop itself from being repaired.

## What is on trial

- **HivePaaS routing settings** - domains, TLS, middlewares on the dashboard itself.
- **HivePaaS service settings**, when the proxy fields change - trusted IPs, proxy hops, provider.
- **Traefik config options** - the startup command.

Changing traefik's replica count, or any setting not listed here, is not put on
trial. Those cannot make the dashboard unreachable.

## While a change is on trial

The dashboard shows a countdown with a confirm button. Two things to know:

**Confirming is refused for the first part of the window.** The new configuration
has to be the one answering before a confirmation means anything; before then the
request would have been served by the configuration being replaced, so it proves
nothing. That is 30 seconds for almost everything - replacing traefik's task takes
about four seconds, and a label change is live as soon as traefik's 15-second poll
picks it up. It is two minutes only for a change that restarts HivePaaS itself,
which is a replica or worker setting sent alongside the proxy fields. The
dashboard knows which and shows the countdown.

**Losing the tab does not lose the change.** The countdown is on the server. Any
reload of the settings page picks it back up - which matters here, because
applying traefik options replaces traefik's task, and the HTTP response to the
change itself often never makes it back down a connection that is being cut.

**A change the cluster already undid cannot be confirmed.** Swarm restores the
previous spec when an update never becomes healthy, and it does so without
telling HivePaaS - so a couple of minutes in, the proxy is serving again, the
dashboard loads, and everything looks fine. It is serving the configuration that
was replaced. Confirming there would leave the saved settings describing
something nothing is running, so the confirmation is refused and the dialog says
what happened. Undoing it, now or at the deadline, is what puts the database back
in agreement with the cluster.

If the deadline passes with no confirmation, the previous configuration is
restored automatically. This is the intended outcome for an operator who is
locked out, and for a script that applies a change and exits: there is
deliberately no way to ask for no trial at all.

## By hand

Reach for this only when the automatic paths cannot run: the database is
unreachable, or the node is in a state where the HivePaaS app is not up. You need
SSH to a swarm manager node.

Roll traefik back to its previous spec:

```sh
docker service rollback hivepaas_traefik
```

The same works for the app and the worker:

```sh
docker service rollback hivepaas_app
docker service rollback hivepaas_worker
```

Two limits worth knowing before you rely on it:

- **It restores the immediately previous spec, and only that.** Any service
  update since the bad one - a replica change, anything HivePaaS applied in the
  meantime - has replaced what `rollback` would return to. It is not a history.
- **It does not touch the database.** HivePaaS's record of the configuration will
  disagree with what is running until the next change through the dashboard,
  which reads the live spec and writes the result back. That is by design, but it
  means a manual rollback should be followed by a look at the settings page.

To see what traefik is actually running:

```sh
docker service inspect hivepaas_traefik \
  --format '{{range .Spec.TaskTemplate.ContainerSpec.Args}}{{println .}}{{end}}'
```

And the failure itself is usually in the task list, which shows why a task exited
even when no container is left:

```sh
docker service ps hivepaas_traefik --no-trunc
```
