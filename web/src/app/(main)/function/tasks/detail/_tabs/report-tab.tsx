"use client";

import * as React from "react";

import { CheckIcon, CopyIcon, DownloadIcon, FileTextIcon, RefreshCwIcon } from "lucide-react";
import { toast } from "sonner";

import { Markdown } from "@/components/markdown";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { useI18n } from "@/i18n";
import { translate as swt } from "@/i18n/runtime";
import { api } from "@/lib/api";
import { copyText } from "@/lib/utils";

export function ReportTab({ taskId }: { taskId: string }) {
  "use no memo";
  const { t: swt, locale: swLocale } = useI18n();

  const [report, setReport] = React.useState<string>("");
  const [loading, setLoading] = React.useState(true);
  const [copied, setCopied] = React.useState(false);
  const [error, setError] = React.useState(false);
  const [revision, refresh] = React.useReducer((n: number) => n + 1, 0);

  React.useEffect(() => {
    let active = true;
    setLoading(true);
    setReport("");
    setError(false);
    api
      .report(taskId)
      .then((text) => {
        if (active) setReport(text);
      })
      .catch(() => {
        if (active) setError(true);
      })
      .finally(() => {
        if (active) setLoading(false);
      });
    return () => {
      active = false;
    };
  }, [taskId, revision]);

  function download() {
    const url = URL.createObjectURL(new Blob([report], { type: "text/markdown;charset=utf-8" }));
    const link = document.createElement("a");
    link.href = url;
    link.download = `skopos-task-${taskId.replace(/[^a-zA-Z0-9_-]/g, "_")}-report.md`;
    link.click();
    setTimeout(() => URL.revokeObjectURL(url), 1000);
  }

  async function copy() {
    if (!report) return;
    const ok = await copyText(report);
    if (ok) {
      setCopied(true);
      toast.success(swt("interface.m0676"));
      setTimeout(() => setCopied(false), 1500);
    } else {
      toast.error(swt("interface.m0677"));
    }
  }

  let content: React.ReactNode;
  if (loading) {
    content = (
      <div className="flex flex-col items-center justify-center gap-2 rounded-md border border-dashed py-16 text-muted-foreground text-sm">
        <FileTextIcon className="size-8 opacity-40" />
        {swt("interface.m0260")}
      </div>
    );
  } else if (error) {
    content = (
      <div
        role="alert"
        className="flex items-center justify-between gap-3 rounded-md border border-destructive/30 p-4 text-sm"
      >
        <span>The report could not be generated. No empty report has been substituted.</span>
        <Button variant="outline" size="sm" onClick={refresh}>
          Retry
        </Button>
      </div>
    );
  } else if (report) {
    content = (
      <div className="overflow-x-auto rounded-md border bg-card p-5 lg:p-8">
        <Markdown text={report} />
      </div>
    );
  } else {
    content = (
      <div className="flex flex-col items-center justify-center gap-2 rounded-md border border-dashed py-16 text-muted-foreground text-sm">
        <FileTextIcon className="size-8 opacity-40" />
        {swt("interface.m0678")}
      </div>
    );
  }

  return (
    <Card>
      <CardHeader className="sticky top-0 z-10 flex-row flex-wrap items-center justify-between gap-3 border-b bg-card">
        <CardTitle className="flex items-center gap-2 text-sm">
          <FileTextIcon className="size-4" /> {swt("interface.m0679")}
        </CardTitle>
        <div className="flex gap-2">
          <Button size="sm" variant="outline" onClick={refresh} disabled={loading}>
            <RefreshCwIcon /> Refresh
          </Button>
          {report && (
            <Button size="sm" onClick={download}>
              <DownloadIcon /> Download Markdown
            </Button>
          )}
          {report && (
            <Button size="sm" variant="outline" onClick={copy}>
              {copied ? <CheckIcon /> : <CopyIcon />} {swt("interface.m0680")}
            </Button>
          )}
        </div>
      </CardHeader>
      <CardContent>
        <p className="mb-4 max-w-3xl text-xs leading-relaxed text-muted-foreground">
          Generated from stored task records. Review findings, recorded conclusions, unfinished attempts and coverage
          limitations before sharing. Raw sessions, traffic and approval history remain available in their workspaces.
        </p>
        {content}
      </CardContent>
    </Card>
  );
}
