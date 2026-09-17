# CWMS Data API source for Table Rock

The `scrape tablerock` command reads Table Rock level/flow/generation from the
USACE **CWMS Data API** (CDA), a structured JSON API, and falls back to the
`tab7d.htm` HTML scraper (`lib/tablerock`) when CDA is disabled, errors, or has
no complete row yet. Water temperature is unaffected — it still comes from the
separate White River Sky source (`lib/tablerock/temperature.go`).

This began as a spike; the evaluation that justified adopting it is kept below.

## Usage

- Production path: `scrape tablerock` uses `Client.GetLatestCompleteRecord`,
  which returns the most recent timestamp where **every stored field** has a
  value, so a slow computed series (e.g. total outflow) lagging the raw series
  never causes an ambiguous `0` to be stored.
- `CDA_DISABLED=true` forces the HTML scraper (an escape hatch if CDA changes a
  timeseries ID or has an outage).
- `go run . cda-spike` fetches the last 48h from CDA and prints the rows (a
  read-only diagnostic; `Client.GetMeasurements` is a display view that may show
  `0` for not-yet-published fields — do not store it).

Requires outbound HTTPS to `cwms-data.usace.army.mil`. The client honours
`HTTPS_PROXY` (`http.ProxyFromEnvironment`) and uses normal TLS verification —
no `InsecureSkipVerify`, which the HTML scraper currently needs.

## Field mapping (tab7d column → CWMS timeseries ID, office `SWL`)

| tab7d column        | CWMS timeseries ID                                             | units | verified |
|---------------------|---------------------------------------------------------------|-------|----------|
| Elevation           | `Table_Rock_Dam-Headwater.Elev.Inst.1Hour.0.Decodes-rev`      | ft    | ✅ 912.30 == HTML |
| Generation (MWh)    | `Table_Rock_Dam.Energy-Gen_Plant.Total.1Hour.1Hour.CCP-Comp`  | MWh   | ✅ |
| Turbine release     | `Table_Rock_Dam.Flow-Plant.Ave.1Hour.1Hour.CCP-Comp`          | cfs   | ✅ 20 == HTML |
| Spillway release    | `Table_Rock_Dam.Flow-Tainter Total.Ave.1Hour.1Hour.Regi-Comp` | cfs   | ✅ 0 == HTML |
| Total release       | `Table_Rock_Dam.Flow-Res Out.Ave.1Hour.1Hour.Regi-Comp`       | cfs   | ✅ 20 == turbine+spillway |
| _(bonus)_ Inflow    | `Table_Rock_Dam.Flow-Res In.Ave...`                           | cfs   | available, not stored today |
| Tailwater elevation | `Table_Rock_Dam-Tailwater.Elev...`                            | ft    | available, not stored today |

Internal consistency check from a live pull (high-generation period):

```
measuredAt (UTC)          level      gen    turbine   spillway      total
2026-09-16 23:00         912.32    122.8     8152.0        0.0     8152.0   # turbine == total, spillway 0
2026-09-17 02:00         912.30     67.3     4642.0        0.0     4642.0
2026-09-17 13:00         912.30      0.0       20.0        0.0       20.0   # minimum flow, no generation
```

`turbine + spillway == total` and generation MWh tracks turbine flow, as
expected hydrologically.

## Gotchas found

1. **Version suffix varies per series.** Elevation/plant-flow/generation are
   published under `CCP-Comp`, but `Flow-Res Out.CCP-Comp` is **empty** — the
   populated total-outflow series is `Regi-Comp`. The mapping can't assume one
   uniform version; each ID was chosen by checking which variant actually has
   data.

2. **Per-series latency / alignment.** Each field is a separate timeseries with
   its own publication cadence. In a live pull the newest raw series (turbine,
   elevation) had a point the computed series (`Flow-Res Out`) did not yet, so
   the most recent aligned row had `total` missing:

   ```
   2026-09-17 14:00   level=912.31  turbine=20  spillway=0  total=0(missing)
   ```

   The spike anchors rows on the elevation timestamps and leaves absent fields
   at their zero value, so a **missing value is indistinguishable from a real
   zero**. A production version should either (a) emit the latest row only when
   all required fields are present, or (b) carry optionality (pointers) so
   "missing" and "0 cfs" are distinct. For the hourly cron this means: take the
   latest _complete_ row, not simply the latest elevation timestamp.

3. **Temperature is not replaced.** CDA only exposes **tailwater** water
   temperature (`Table_Rock_Dam-Tailwater.Temp-Water...`), i.e. below the dam —
   not the lake-surface temperature the app displays. That still comes from the
   separate White River Sky source (see `lib/tablerock/temperature.go`).

## CDA vs. HTML scraping

| | `tab7d.htm` (fallback) | CWMS Data API (primary) |
|---|---|---|
| Format | fixed-width text in HTML | JSON, versioned (`?version=2`) |
| Parsing | split on `<hr>`, whitespace tokens, regex date rows | `json.Unmarshal` |
| Timezone | page-local (CST/CDT), needs `America/Chicago` parse | epoch-millis UTC, unambiguous |
| TLS | needs `InsecureSkipVerify: true` | valid cert, verification on |
| Proxy | custom transport ignores proxy env | honours `HTTPS_PROXY` |
| Row assembly | one request → full row | N requests → align by timestamp |
| Latency handling | source pre-aligns the row | caller must handle per-series lag |
| Units | implicit / assumed | explicit in the response |

## What was adopted

CDA is the primary source for level/flow/generation, with the `tab7d` HTML
scraper kept as an automatic fallback and temperature left on its separate
source. The spike's two open questions were resolved as:

- **"Latest complete row" rule** — `GetLatestCompleteRecord` returns the newest
  timestamp where all stored fields are present; per-field presence is tracked
  with pointers so a not-yet-published field is never stored as `0`.
- **Observed vs. forecast** — the mapping uses observed/computed series only and
  avoids the `National-CWMS-Forecast` series (predictions, not observations).

The `tab7d` scraper (`lib/tablerock`) remains as the fallback while CDA proves
itself in production; `CDA_DISABLED=true` switches back to it on demand. The
mapping notes above are the reference if a timeseries ID ever changes.

### Backfill

`scrape tablerock --all` still uses the `tab7d` 7-day HTML table, which is what
that flag was built for. Only the hourly latest path uses CDA.
