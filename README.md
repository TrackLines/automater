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

A service from this repo, with no domain. `railway.json` sets the Dockerfile build, the schedule `0 0 * * *` (UTC) and no restarts. Its variables:

| Variable | What |
|---|---|
| `DATABASE_URL` | the backend's Postgres (Railway's reference variable). Or the backend's own `RDS_HOSTNAME`, `RDS_PORT`, `RDS_DB`, `RDS_USERNAME` and `RDS_PASSWORD` |
| `UPLOADTHING_TOKEN` | the same as the backend's, for deleting files |
| `BUGFIXES_AGENT_KEY`, `BUGFIXES_AGENT_SECRET` | an agent for the `automation` service. Without them, failures are only logged |

Run it once by hand after setting it up. The log should show `purge: …` and `sweep-files: …`.

## Development

```sh
go test ./...
```

The backend's `retention` tests cover the SQL against a real database. This repo's tests cover running the two jobs and reading the database settings.
