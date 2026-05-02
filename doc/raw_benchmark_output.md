# Raw Benchmark Output - Quark ORM

```text
=== RUN   TestBenchmarkEngines
=== RUN   TestBenchmarkEngines/SQLite
2026/05/02 13:47:14 INFO quark client initialized dialect=sqlite max_results=10000
[SQLite] Inserting 10,000 records...
  [VISUAL LOG] INSERT: processed 1000 records (last duration: 7.375µs)
  [VISUAL LOG] INSERT: processed 2000 records (last duration: 6.333µs)
  [VISUAL LOG] INSERT: processed 3000 records (last duration: 5.625µs)
  [VISUAL LOG] INSERT: processed 4000 records (last duration: 6.083µs)
  [VISUAL LOG] INSERT: processed 5000 records (last duration: 5.25µs)
  [VISUAL LOG] INSERT: processed 6000 records (last duration: 5.25µs)
  [VISUAL LOG] INSERT: processed 7000 records (last duration: 5.166µs)
  [VISUAL LOG] INSERT: processed 8000 records (last duration: 5.292µs)
  [VISUAL LOG] INSERT: processed 9000 records (last duration: 5.417µs)
  [VISUAL LOG] INSERT: processed 10000 records (last duration: 5.209µs)
[SQLite] Selecting all records (List)...
[SQLite] Streaming all records (Iter)...
[SQLite] Updating 5,000 records...
  [VISUAL LOG] UPDATE: processed 1000 records (last duration: 4.083µs)
  [VISUAL LOG] UPDATE: processed 2000 records (last duration: 3.167µs)
  [VISUAL LOG] UPDATE: processed 3000 records (last duration: 2.75µs)
  [VISUAL LOG] UPDATE: processed 4000 records (last duration: 2.875µs)
  [VISUAL LOG] UPDATE: processed 5000 records (last duration: 2.75µs)
[SQLite] Deleting 5,000 records...
  [VISUAL LOG] DELETE (hard): processed 1000 records (last duration: 2.458µs)
  [VISUAL LOG] DELETE (hard): processed 2000 records (last duration: 1.833µs)
  [VISUAL LOG] DELETE (hard): processed 3000 records (last duration: 1.625µs)
  [VISUAL LOG] DELETE (hard): processed 4000 records (last duration: 1.5µs)
  [VISUAL LOG] DELETE (hard): processed 5000 records (last duration: 1.542µs)

--- BENCHMARK METRICS SUMMARY ---
[INSERT] Total Ops: 10000, Avg Time: 6.118µs, Total Time: 61.186522ms
[SELECT] Total Ops: 1, Avg Time: 22.5µs, Total Time: 22.5µs
[SELECT (stream)] Total Ops: 1, Avg Time: 19.334µs, Total Time: 19.334µs
[UPDATE] Total Ops: 5000, Avg Time: 3.235µs, Total Time: 16.175375ms
[DELETE (hard)] Total Ops: 5000, Avg Time: 1.917µs, Total Time: 9.588279ms
=== RUN   TestBenchmarkEngines/Postgres
2026/05/02 13:47:15 INFO quark client initialized dialect=postgres max_results=10000
[Postgres] Inserting 10,000 records...
  [VISUAL LOG] INSERT: processed 1000 records (last duration: 227.167µs)
  [VISUAL LOG] INSERT: processed 2000 records (last duration: 181.75µs)
  [VISUAL LOG] INSERT: processed 3000 records (last duration: 196.209µs)
  [VISUAL LOG] INSERT: processed 4000 records (last duration: 195.459µs)
  [VISUAL LOG] INSERT: processed 5000 records (last duration: 194.5µs)
  [VISUAL LOG] INSERT: processed 6000 records (last duration: 201.208µs)
  [VISUAL LOG] INSERT: processed 7000 records (last duration: 189.584µs)
  [VISUAL LOG] INSERT: processed 8000 records (last duration: 193.375µs)
  [VISUAL LOG] INSERT: processed 9000 records (last duration: 219.125µs)
  [VISUAL LOG] INSERT: processed 10000 records (last duration: 196.541µs)
[Postgres] Selecting all records (List)...
[Postgres] Streaming all records (Iter)...
[Postgres] Updating 5,000 records...
  [VISUAL LOG] UPDATE: processed 1000 records (last duration: 111.458µs)
  [VISUAL LOG] UPDATE: processed 2000 records (last duration: 104µs)
  [VISUAL LOG] UPDATE: processed 3000 records (last duration: 118.834µs)
  [VISUAL LOG] UPDATE: processed 4000 records (last duration: 153.709µs)
  [VISUAL LOG] UPDATE: processed 5000 records (last duration: 131.541µs)
[Postgres] Deleting 5,000 records...
  [VISUAL LOG] DELETE (hard): processed 1000 records (last duration: 158.417µs)
  [VISUAL LOG] DELETE (hard): processed 2000 records (last duration: 122.791µs)
  [VISUAL LOG] DELETE (hard): processed 3000 records (last duration: 99.458µs)
  [VISUAL LOG] DELETE (hard): processed 4000 records (last duration: 113.083µs)
  [VISUAL LOG] DELETE (hard): processed 5000 records (last duration: 112.667µs)

--- BENCHMARK METRICS SUMMARY ---
[DELETE (hard)] Total Ops: 5000, Avg Time: 128.526µs, Total Time: 642.632468ms
[INSERT] Total Ops: 10000, Avg Time: 198.549µs, Total Time: 1.985498553s
[SELECT] Total Ops: 1, Avg Time: 502µs, Total Time: 502µs
[SELECT (stream)] Total Ops: 1, Avg Time: 411.375µs, Total Time: 411.375µs
[UPDATE] Total Ops: 5000, Avg Time: 129.763µs, Total Time: 648.817527ms
=== RUN   TestBenchmarkEngines/MySQL
2026/05/02 13:47:18 INFO quark client initialized dialect=mysql max_results=10000
[MySQL] Inserting 10,000 records...
  [VISUAL LOG] INSERT: processed 1000 records (last duration: 1.001375ms)
  [VISUAL LOG] INSERT: processed 5000 records (last duration: 751.834µs)
  [VISUAL LOG] INSERT: processed 10000 records (last duration: 787.25µs)
[MySQL] Selecting all records (List)...
[MySQL] Streaming all records (Iter)...
[MySQL] Updating 5,000 records...
  [VISUAL LOG] UPDATE: processed 5000 records (last duration: 277.958µs)
[MySQL] Deleting 5,000 records...
  [VISUAL LOG] DELETE (hard): processed 5000 records (last duration: 325.583µs)

--- BENCHMARK METRICS SUMMARY ---
[INSERT] Total Ops: 10000, Avg Time: 979.431µs, Total Time: 9.794319035s
[SELECT] Total Ops: 1, Avg Time: 402.208µs, Total Time: 402.208µs
[SELECT (stream)] Total Ops: 1, Avg Time: 434.458µs, Total Time: 434.458µs
[UPDATE] Total Ops: 5000, Avg Time: 275.566µs, Total Time: 1.377834816s
[DELETE (hard)] Total Ops: 5000, Avg Time: 266.179µs, Total Time: 1.330897214s
=== RUN   TestBenchmarkEngines/MSSQL
2026/05/02 13:47:32 INFO quark client initialized dialect=mssql max_results=10000
[MSSQL] Inserting 10,000 records...
  [VISUAL LOG] INSERT: processed 10000 records (last duration: 771.875µs)
[MSSQL] Updating 5,000 records...
  [VISUAL LOG] UPDATE: processed 5000 records (last duration: 297.416µs)
[MSSQL] Deleting 5,000 records...
  [VISUAL LOG] DELETE (hard): processed 5000 records (last duration: 235.291µs)

--- BENCHMARK METRICS SUMMARY ---
[DELETE (hard)] Total Ops: 5000, Avg Time: 265.405µs, Total Time: 1.327027224s
[INSERT] Total Ops: 10000, Avg Time: 651.5µs, Total Time: 6.515007703s
[SELECT] Total Ops: 1, Avg Time: 1.863292ms, Total Time: 1.863292ms
[SELECT (stream)] Total Ops: 1, Avg Time: 1.842833ms, Total Time: 1.842833ms
[UPDATE] Total Ops: 5000, Avg Time: 266.882µs, Total Time: 1.334411798s
=== RUN   TestBenchmarkEngines/Oracle
2026/05/02 13:47:41 INFO quark client initialized dialect=oracle max_results=10000
[Oracle] Inserting 10,000 records...
  [VISUAL LOG] INSERT: processed 10000 records (last duration: 440.625µs)
[Oracle] Updating 5,000 records...
  [VISUAL LOG] UPDATE: processed 5000 records (last duration: 254.667µs)
[Oracle] Deleting 5,000 records...
  [VISUAL LOG] DELETE (hard): processed 5000 records (last duration: 263.667µs)

--- BENCHMARK METRICS SUMMARY ---
[DELETE (hard)] Total Ops: 5000, Avg Time: 269.337µs, Total Time: 1.346685046s
[INSERT] Total Ops: 10000, Avg Time: 431.733µs, Total Time: 4.317334732s
[SELECT] Total Ops: 1, Avg Time: 3.652292ms, Total Time: 3.652292ms
[SELECT (stream)] Total Ops: 1, Avg Time: 1.691917ms, Total Time: 1.691917ms
[UPDATE] Total Ops: 5000, Avg Time: 271.937µs, Total Time: 1.359689576s
--- PASS: TestBenchmarkEngines (33.99s)
    --- PASS: TestBenchmarkEngines/SQLite (0.14s)
    --- PASS: TestBenchmarkEngines/Postgres (3.37s)
    --- PASS: TestBenchmarkEngines/MySQL (14.03s)
    --- PASS: TestBenchmarkEngines/MSSQL (9.30s)
    --- PASS: TestBenchmarkEngines/Oracle (7.15s)
PASS
ok  	github.com/jcsvwinston/GoFrame/pkg/quark	34.354s
```
