# Spike: CWMS Data API as an alternative to `tab7d.htm` scraping

**Status:** spike / proof-of-concept. Read-only, nothing is wired into the
scrape/store path. Goal: determine whether the USACE **CWMS Data API** (CDA)
can replace scraping the human-facing `tab7d.htm` page, and at what cost.

**Verdict:** Yes — every field the current scraper stores is available from CDA
as structured JSON, verified against the live API and cross-checked with the
HTML page. The main costs are (1) one HTTP request per field instead of one for
the whole row, and (2) per-series publication latency that the alignment step
has to handle. Water temperature still comes from a separate source, exactly as
today.

## How to run

```bash
go run . cda-spike        # fetches the last 48h from CDA and prints the rows
```

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

| | `tab7d.htm` (current) | CWMS Data API (this spike) |
|---|---|---|
| Format | fixed-width text in HTML | JSON, versioned (`?version=2`) |
| Parsing | split on `<hr>`, whitespace tokens, regex date rows | `json.Unmarshal` |
| Timezone | page-local (CST/CDT), needs `America/Chicago` parse | epoch-millis UTC, unambiguous |
| TLS | needs `InsecureSkipVerify: true` | valid cert, verification on |
| Proxy | custom transport ignores proxy env | honours `HTTPS_PROXY` |
| Row assembly | one request → full row | N requests → align by timestamp |
| Latency handling | source pre-aligns the row | caller must handle per-series lag |
| Units | implicit / assumed | explicit in the response |

## Recommendation

Adopt CDA for level/flow/generation and keep temperature on its current
separate source. It removes the brittle HTML parsing, the timezone guesswork,
and the TLS-verification bypass, in exchange for a small amount of
timestamp-alignment logic (gotcha #2). Before switching the production path:

- Decide the "latest complete row" rule and represent missing values
  explicitly (don't store an ambiguous `0`).
- Pick observed vs. computed/forecast series deliberately (this spike avoids
  the `National-CWMS-Forecast` series, which are predictions, not observations).
- Keep `tab7d` mapping notes around until CDA has run in parallel long enough
  to trust the row alignment.

This spike does not change any existing behaviour; it only adds `lib/cda` and
the `cda-spike` command.
