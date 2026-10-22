"use client";

import * as React from "react";

import Link from "next/link";

import { ArrowLeftIcon, ArrowRightIcon, RefreshCwIcon, SearchIcon, XIcon } from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { api } from "@/lib/api";
import { nodeSummary, taskWorkspaceHref } from "@/lib/task-workspace";
import type { ExplorationNodePage, ExploreKind, TaskNode } from "@/lib/types";
import { cn } from "@/lib/utils";

function recordKind(node: TaskNode) {
  if (node.type === "finding") return "Finding";
  if (node.type === "fact") return node.state === "origin" ? "Starting context" : "Recorded conclusion";
  return "Exploration attempt";
}

export function ResultsTab({ taskId }: { taskId: string }) {
  const [query, setQuery] = React.useState("");
  const [search, setSearch] = React.useState("");
  const [kind, setKind] = React.useState("all");
  const [page, setPage] = React.useState(1);
  const [data, setData] = React.useState<ExplorationNodePage | null>(null);
  const [selectedId, setSelectedId] = React.useState<string | null>(null);
  const [loading, setLoading] = React.useState(true);
  const [error, setError] = React.useState(false);
  const [revision, refresh] = React.useReducer((n: number) => n + 1, 0);
  const size = 25;

  React.useEffect(() => {
    // An unchanged initial query must not clear a row selected during mount.
    if (query.trim() === search) return;
    const timer = setTimeout(() => {
      setSearch(query.trim());
      setPage(1);
      setSelectedId(null);
    }, 250);
    return () => clearTimeout(timer);
  }, [query, search]);

  React.useEffect(() => {
    let live = true;
    let timer: ReturnType<typeof setTimeout>;
    setLoading(true);
    setData(null);
    setError(false);
    async function load() {
      try {
        const result = await api.explorationNodes(taskId, {
          page,
          size,
          kinds: kind === "all" ? ["fact", "intent", "finding"] : [kind as ExploreKind],
          q: search,
          order: "desc",
        });
        if (!live) return;
        setData(result);
        setError(false);
        setSelectedId((current) => (result.items.some((node) => node.id === current) ? current : null));
      } catch {
        if (live) setError(true);
      } finally {
        if (live) {
          setLoading(false);
          timer = setTimeout(load, 15000);
        }
      }
    }
    void load();
    return () => {
      live = false;
      clearTimeout(timer);
    };
  }, [taskId, page, kind, search, revision]);

  const selected = data?.items.find((item) => item.id === selectedId);
  const assets = selected ? (data?.assets[selected.id] ?? []) : [];
  let payload = selected?.payload ?? "";
  try {
    payload = JSON.stringify(JSON.parse(payload), null, 2);
  } catch {
    /* Plain-text records remain readable. */
  }

  return (
    <div className="space-y-4">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <h2 className="text-lg font-semibold tracking-tight">Results ledger</h2>
          <p className="mt-1 max-w-3xl text-sm leading-relaxed text-muted-foreground">
            Review exploration findings, recorded conclusions and unfinished attempts together. Negative observations
            stay in their original context.
          </p>
        </div>
        <Button variant="outline" size="sm" onClick={refresh}>
          <RefreshCwIcon /> Refresh
        </Button>
      </div>
      <div className="rounded-md border bg-muted/30 p-3 text-xs leading-relaxed text-muted-foreground">
        This ledger reflects the exploration graph; the Findings workspace also includes standalone finding records. A
        completed attempt is not a security verdict. Conclusions may describe no vulnerability observed, partial
        evidence or an inconclusive result. Blocked and budget-exhausted attempts need separate review; unrecorded
        checks cannot be inferred from this ledger.
      </div>
      <div className="flex flex-wrap items-center gap-2">
        <div className="relative min-w-48 max-w-sm flex-1">
          <SearchIcon className="pointer-events-none absolute top-2.5 left-3 size-4 text-muted-foreground" />
          <Input
            aria-label="Search recorded results"
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            placeholder="Search conclusions and evidence…"
            className="pl-9"
          />
        </div>
        <Select
          value={kind}
          onValueChange={(value) => {
            setKind(value);
            setPage(1);
            setSelectedId(null);
          }}
        >
          <SelectTrigger aria-label="Filter result type" className="w-48">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All recorded results</SelectItem>
            <SelectItem value="fact">Recorded conclusions</SelectItem>
            <SelectItem value="intent">Exploration attempts</SelectItem>
            <SelectItem value="finding">Findings</SelectItem>
          </SelectContent>
        </Select>
        <span className="ml-auto text-xs tabular-nums text-muted-foreground">
          {data ? `${data.total} records` : ""}
        </span>
      </div>
      {error && (
        <div
          role="alert"
          className="flex items-center justify-between gap-3 rounded-md border border-destructive/30 p-3 text-sm"
        >
          <span>
            {data ? "Results could not refresh. The displayed records may be stale." : "Results could not be loaded."}
          </span>
          <Button size="sm" variant="outline" onClick={refresh}>
            Retry
          </Button>
        </div>
      )}
      <div className={cn("grid min-w-0 gap-4", selected && "xl:grid-cols-[minmax(0,1fr)_20rem]")}>
        <div className="min-w-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Record</TableHead>
                <TableHead>Summary</TableHead>
                <TableHead>Recorded state</TableHead>
                <TableHead className={cn(selected && "hidden 2xl:table-cell")}>When</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {!data && loading && (
                <TableRow>
                  <TableCell colSpan={4} className="h-36 text-center text-muted-foreground" role="status">
                    Loading recorded results…
                  </TableCell>
                </TableRow>
              )}
              {data?.items.map((node) => (
                <TableRow key={node.id} data-state={selectedId === node.id ? "selected" : undefined}>
                  <TableCell className="align-top">
                    <Badge variant="outline" className="whitespace-nowrap">
                      {recordKind(node)}
                    </Badge>
                    <code className="mt-1 block text-[11px] text-muted-foreground">#{node.id}</code>
                  </TableCell>
                  <TableCell className="max-w-lg whitespace-normal align-top">
                    <button
                      type="button"
                      className="w-full rounded text-left text-sm leading-relaxed underline-offset-4 hover:underline focus-visible:outline-2 focus-visible:outline-ring"
                      aria-expanded={selectedId === node.id}
                      aria-controls="result-inspector"
                      onClick={() => setSelectedId((current) => (current === node.id ? null : node.id))}
                    >
                      <span className="line-clamp-3 break-words">{nodeSummary(node.payload)}</span>
                    </button>
                    {node.inherited && (
                      <span className="mt-1 block text-xs text-muted-foreground">
                        Read-only source task {node.source_task_id}
                      </span>
                    )}
                  </TableCell>
                  <TableCell className="align-top">
                    <code className="rounded bg-muted px-1.5 py-1 text-[11px]">{node.state || "unspecified"}</code>
                    {node.type === "intent" && ["blocked", "exhausted", "stopped", "deleted"].includes(node.state) && (
                      <span className="mt-2 block text-xs text-muted-foreground">Execution incomplete or stopped</span>
                    )}
                  </TableCell>
                  <TableCell
                    className={cn(
                      "whitespace-nowrap align-top text-xs text-muted-foreground",
                      selected && "hidden 2xl:table-cell",
                    )}
                  >
                    {node.ts ? new Date(node.ts).toLocaleString() : "Not recorded"}
                  </TableCell>
                </TableRow>
              ))}
              {data?.items.length === 0 && (
                <TableRow>
                  <TableCell colSpan={4} className="h-36 text-center text-muted-foreground">
                    {search || kind !== "all"
                      ? "No records match these filters."
                      : "No exploration results recorded yet. This does not establish that checks passed."}
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
          <div className="mt-3 flex items-center justify-between text-xs text-muted-foreground">
            <span>
              Page {page}
              {data ? ` of ${Math.max(1, Math.ceil(data.total / size))}` : ""} · Refreshes every 15 seconds
            </span>
            <div className="flex gap-2">
              <Button
                variant="outline"
                size="sm"
                disabled={page === 1 || loading}
                onClick={() => {
                  setPage((n) => n - 1);
                  setSelectedId(null);
                }}
              >
                <ArrowLeftIcon /> Previous
              </Button>
              <Button
                variant="outline"
                size="sm"
                disabled={!data || page * size >= data.total || loading}
                onClick={() => {
                  setPage((n) => n + 1);
                  setSelectedId(null);
                }}
              >
                Next <ArrowRightIcon />
              </Button>
            </div>
          </div>
        </div>
        {selected && (
          <aside
            id="result-inspector"
            aria-label="Recorded result details"
            className="sticky top-0 max-h-[calc(100dvh-13rem)] min-w-0 self-start overflow-y-auto rounded-lg border bg-card"
          >
            <div className="flex items-center justify-between border-b p-3">
              <h3 className="text-sm font-semibold">Record #{selected.id}</h3>
              <Button
                variant="ghost"
                size="icon-sm"
                aria-label="Close result details"
                onClick={() => setSelectedId(null)}
              >
                <XIcon />
              </Button>
            </div>
            <div className="space-y-4 p-4">
              <p className="whitespace-pre-wrap break-words text-sm leading-relaxed">{nodeSummary(selected.payload)}</p>
              <dl className="grid grid-cols-[auto_1fr] gap-x-4 gap-y-2 text-xs">
                <dt className="text-muted-foreground">Type</dt>
                <dd>{recordKind(selected)}</dd>
                <dt className="text-muted-foreground">State</dt>
                <dd>{selected.state}</dd>
                <dt className="text-muted-foreground">Origin</dt>
                <dd className="break-all">{selected.origin || "Not recorded"}</dd>
                <dt className="text-muted-foreground">Recorded at</dt>
                <dd>{selected.ts ? new Date(selected.ts).toLocaleString() : "Not recorded"}</dd>
              </dl>
              {selected.delete_reason && <p className="text-xs">Removal reason: {selected.delete_reason}</p>}
              {assets.length > 0 && (
                <div>
                  <h4 className="mb-2 text-xs font-semibold">Linked assets</h4>
                  {assets.map((asset) => (
                    <p key={asset.id} className="break-all font-mono text-xs">
                      {asset.label}
                    </p>
                  ))}
                </div>
              )}
              <details>
                <summary className="cursor-pointer text-xs font-medium">Original recorded payload</summary>
                <pre className="mt-2 max-h-96 overflow-auto whitespace-pre-wrap break-all rounded-md bg-muted p-3 text-xs">
                  {payload || "No payload recorded."}
                </pre>
              </details>
              <div className="flex flex-wrap gap-3 text-xs">
                <Link
                  className="text-primary underline underline-offset-4"
                  href={taskWorkspaceHref(taskId, "sessions")}
                >
                  Open session evidence
                </Link>
                <Link className="text-primary underline underline-offset-4" href={taskWorkspaceHref(taskId, "graph")}>
                  Open exploration paths
                </Link>
              </div>
            </div>
          </aside>
        )}
      </div>
    </div>
  );
}
