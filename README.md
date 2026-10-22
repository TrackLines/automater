# automation


TrackLin.es's nightly clean-up, as a small Railway cron service. Once a night it:

1. **Purges:**
   - tickets 30 days in the bin, with their comments, labels, links and attachments;
   - organizations deleted 30 days ago;
   - year-old attachments: it warns the org's admins and the uploader, and deletes the files nobody kept after 13 months;
   - expired Google Chat link codes.
2. **Deletes files** from UploadThing once their day's batch is 30 days old.

Then it exits. Both jobs run every night, even if the other fails, and failures go to [Bugfixes](https://bugfix.es) as the service `automation`.

The privacy policy promises this deletion, and it starts within 30 days of the first real customer (TL-431).

The rules live in the backend's [`retention`](https://github.com/TrackLines/backend/tree/main/retention) package, so this service and the backend's own `/server purge` and `/server sweep-files` run the same code. This service runs no migrations, because the tables are the backend's.

## Railway

A service from this repo, with no domain. It needs no start command: the image's entrypoint runs both jobs and exits. Set these in the service's **Settings** (new services can't use a `railway.json`, since Railway deprecated Config as Code):

- **Build:** Dockerfile (Railway finds `Dockerfile` at the root).
- **Cron Schedule:** `0 0 * * *` (UTC).
- **Restart Policy:** Never, so a failed night isn't run again straight away. The next night retries.
- **Deploys:** only from tagged releases. Publishing the Release Drafter draft creates a `v*` tag, and `.github/workflows/deploy-production.yml` runs `railway up` for it. Turn off the service's GitHub auto-deploy from `main`.

Its variables:

| Variable | What |
|---|---|
| `RDS_URL` | the backend's Postgres, the same value as the backend's. Or `RDS_HOSTNAME`, `RDS_PORT`, `RDS_DB`, `RDS_USERNAME` and `RDS_PASSWORD` |
| `UPLOADTHING_TOKEN` | the same as the backend's, for deleting files |
| `BUGFIXES_AGENT_KEY`, `BUGFIXES_AGENT_SECRET` | an agent for the `automation` service. Without them, or with `BUGFIXES_LOCAL_ONLY=true`, failures are only logged |

It reads these with [go-config](https://github.com/keloran/go-config), like the backend, so the names are the same.

Run it once by hand after setting it up. The log shows each step:

```
automation: starting
automation: connected to postgres at …
purge: starting
purge: 0 ticket(s) from the bin, 0 deleted organization(s), … (8ms)
sweep-files: starting
sweep-files: deleted 0 file(s) from UploadThing (1ms)
automation: done
```

If the database can't be reached, it says so within 15 seconds and exits with 1. A run that takes longer than 10 minutes is stopped and reported, so it never runs into the next night.

## Development

```sh
go test ./...
```

The backend's `retention` tests cover the SQL against a real database. This repo's tests cover running the two jobs.
