# Large files into and out of an app's container

Copying a file into an app's container stops after 60 seconds, and out of it
after 180: the CLI's `cp`, and the dashboard's upload and export alike. This
design lifts both limits without loosening the timeouts that protect every app
on the installation.

## Where the limits are

| Layer | Setting | What it cuts |
|---|---|---|
| Traefik v3.7, entrypoint `websecure` | `respondingTimeouts.readTimeout` = 60s, its default (HivePaaS sets none); `writeTimeout` = 0 | the reading of a whole request, its body included: **an upload longer than 60 seconds** - about 75 MB on a 10 Mbit/s uplink |
| The app, `interface/api/server/server.go` | `ReadTimeout` and `WriteTimeout` = 180s | an upload longer than 3 minutes (behind Traefik's 60 s); **a download longer than 3 minutes** |

Go clears a connection's read deadline once a request's body has been read
(`connReader.startBackgroundRead`), so a download is not cut by a read timeout,
Traefik's or the app's - only by the app's write timeout. And Go clears both
deadlines of a connection a handler hijacks (`conn.hijackLocked`), as a
websocket upgrade does, in Traefik and in the app: a websocket is cut by
neither.

Raising Traefik's `readTimeout` would lift the upload limit in one line, for
every app on the installation, and take away what the 60 seconds are for: a
client that sends its request a byte at a time holds a connection for no longer
than that. An admin can still do it, through Traefik's config options; the
platform does not.

## Downloads: the deadline moves with the bytes

`GET .../container/file-download` keeps its shape. While it copies, it pushes
its connection's write deadline 60 seconds past each write
(`http.ResponseController.SetWriteDeadline`): a copy that goes on, however
slow, is never cut; one that stalls for 60 seconds still is.

Two faults of the same handler go with it:

- **A cut is told from an end.** The copy's error was dropped, and a directory -
  a tar, of no announced length - cut short read as a whole one. On an error
  after the first bytes, the handler now closes the connection, so the client
  reads an unexpected EOF rather than a clean end. It takes the connection from
  the server's own writer: gin refuses to hand it over once it has written. An
  error before the first byte is answered as an error still.
- **The file's name** in `Content-Disposition`'s `filename*` was escaped with
  `url.QueryEscape`, which writes a space as `+`; RFC 5987 wants `%20`:
  `url.PathEscape`.

## Uploads: a websocket

`GET .../container/file-upload/stream`, upgraded to a websocket, takes what
`POST .../container/file-upload` takes, as query parameters: `path`, `extract`,
`compressionFormat`, `overwrite` (true when left out, as for the form),
`nodeId`, `containerId`; and `fileName` and `fileSize`, which the form's file
part gave, `fileSize` required for a single file - the tar the server wraps it
in starts with its size. Auth and validation answer as any request does, before
the upgrade: a refusal is an HTTP error.

Then:

1. The client sends the content as binary messages, of at most 4 MiB each, and
   `{"type":"end"}` as a text message when it is done.
2. The server feeds them, through a pipe, to the same use case the form does -
   `appContainerUC.UploadFileToContainer` - so the two go the same way: a file
   wrapped in a tar, an archive extracted, a remote node's agent.
3. It answers one text message, `{"type":"done","data":{"path","message"}}` or
   `{"type":"error","error":<hperrors.ErrorInfo>}`, and closes.

- **A single file is the size it said.** Its content is read through a reader
  held to `fileSize`: a byte more, or fewer, is a 400 naming `fileSize`. Without
  it, `fileSize=0` would have the server buffer whatever came, in memory.
- **A copy that went well is done.** An archive's copy ends at the archive's end
  marker; what the client sends after it - the padding `tar` writes - finds the
  pipe closed, and is no failure.
- **A use case that panics is answered**, as an error, and the stream closed: the
  handler does not wait on it for good.
- **What the client caused is the client's.** A client that leaves before its end
  message, or goes silent, is answered - and kept - as a 400, not a 500.

A client that sends no message for 60 seconds ends the transfer, as an error:
the pipe is closed with it, and the copy into the container fails rather than
waiting. The limit is per message, not per byte: a client sends small ones -
the CLI 256 KiB - so that a slow link still sends one a minute. A use case that fails while the client still sends - no running
container, a refusal from Docker - closes the pipe, and its error is the
answer. The pipe gives the transfer the container's pace: the server reads the
next message only when the last one has been written into the copy.

Upload and stream keep their permissions: `write` for an upload, as the form's.
The demo user, refused every write, is refused this GET too, as `/terminal` is.

## The CLI

`cp` sends its upload over the websocket, and a directory compressed
(`compressionFormat=gzip`, a `.tar.gz` the server extracts). A download stays
uncompressed: a gzipped answer no longer says whether it is a directory's tar or
a file. A server without the stream endpoint (404) gets the form, as now, with
the CLI saying that an upload longer than 60 seconds will be cut. Progress -
bytes and rate - goes to stderr at a terminal.

## Testing

- The download handler's copy under an `httptest` server whose `WriteTimeout`
  is shorter than the copy: it ends whole. A copy that fails half-way: the
  client reads an unexpected EOF.
- The stream's pump under an `httptest` websocket: the content arrives whole and
  in order, the end message ends it, a silent client is cut after the idle time,
  a use case that fails stops the pump.
- Before release, through Traefik in the dind environment: an upload and a
  download that each take longer than 60 seconds.

## Not in this

- Resumable, chunked uploads: a transfer cut by the network starts again. They
  need storage the app's replicas share, and a cleanup of what a client
  abandons.
- The dashboard's switch to the stream endpoint: its own change, after this.
