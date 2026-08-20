"use client";

import dynamic from "next/dynamic";
import { useMemo, useRef, useState } from "react";
import type { OnMount } from "@monaco-editor/react";

const Monaco = dynamic(() => import("@monaco-editor/react"), {
  ssr: false,
  loading: () => (
    <div className="flex h-full items-center px-4 text-sm text-[#1e7a3a]">
      loading editor_
    </div>
  ),
});

type Language = "javascript" | "cpp";

const TEMPLATES: Record<Language, string> = {
  javascript: `console.log(1 + 1);
`,
  cpp: `#include <iostream>
using namespace std;

int main() {
    cout << 1 + 1 << endl;
    return 0;
}
`,
};

type RunResult = {
  id?: string;
  status?: string;
  output?: string;
  error?: string;
};

async function pollResult(id: string): Promise<RunResult> {
  const deadline = Date.now() + 30_000;

  while (Date.now() < deadline) {
    const res = await fetch(`/api/result?id=${encodeURIComponent(id)}`);
    const data = (await res.json()) as RunResult;

    if (res.status === 202 || data.status === "pending") {
      await new Promise((r) => setTimeout(r, 400));
      continue;
    }

    return data;
  }

  throw new Error("timed out waiting for worker");
}

export default function Editor() {
  const [language, setLanguage] = useState<Language>("javascript");
  const [code, setCode] = useState(TEMPLATES.javascript);
  const [running, setRunning] = useState(false);
  const [result, setResult] = useState<RunResult | null>(null);
  const [error, setError] = useState("");
  const runRef = useRef<() => Promise<void>>(async () => {});

  const monacoLanguage = useMemo(
    () => (language === "cpp" ? "cpp" : "javascript"),
    [language],
  );

  const prompt = language === "cpp" ? "g++" : "node";

  function onLanguageChange(next: Language) {
    setLanguage(next);
    setCode(TEMPLATES[next]);
    setResult(null);
    setError("");
  }

  async function run() {
    setRunning(true);
    setError("");
    setResult(null);

    try {
      const res = await fetch(`/api/execute`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ language, code }),
      });

      if (!res.ok) {
        throw new Error(await res.text());
      }

      const queued = (await res.json()) as { id: string; status: string };
      const done = await pollResult(queued.id);
      setResult({ ...done, id: queued.id });
    } catch (e) {
      setError(e instanceof Error ? e.message : "run failed");
    } finally {
      setRunning(false);
    }
  }

  runRef.current = run;

  const onMount: OnMount = (editor, monaco) => {
    monaco.editor.defineTheme("terminal", {
      base: "vs-dark",
      inherit: true,
      rules: [],
      colors: {
        "editor.background": "#050805",
        "editor.foreground": "#3dff7a",
        "editor.lineHighlightBackground": "#0a140a",
        "editorCursor.foreground": "#3dff7a",
        "editorLineNumber.foreground": "#1e7a3a",
        "editorLineNumber.activeForeground": "#3dff7a",
        "editor.selectionBackground": "#1e7a3a55",
        "editorGutter.background": "#050805",
      },
    });
    monaco.editor.setTheme("terminal");
    editor.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyCode.Enter, () => {
      void runRef.current();
    });
  };

  let outputText = "press run or ctrl+enter";
  if (error) {
    outputText = error;
  } else if (running) {
    outputText = "queued… waiting for worker";
  } else if (result) {
    outputText =
      [result.output, result.error].filter(Boolean).join("\n") || "(no output)";
  }

  return (
    <div className="flex h-screen flex-col p-3">
      <div className="relative flex min-h-0 flex-1 flex-col overflow-hidden rounded-md border border-[#1e7a3a] bg-[#050805] shadow-[0_0_40px_#3dff7a22]">
        <header className="flex items-center justify-between border-b border-[#1e7a3a] px-3 py-2 text-xs">
          <div className="flex items-center gap-2">
            <span className="h-2.5 w-2.5 rounded-full bg-[#ff5f56]" />
            <span className="h-2.5 w-2.5 rounded-full bg-[#ffbd2e]" />
            <span className="h-2.5 w-2.5 rounded-full bg-[#27c93f]" />
            <img
              src="/logo.svg"
              alt="Code Run"
              width={22}
              height={22}
              className="ml-3"
            />
            <span className="ml-1 font-semibold tracking-widest text-[#3dff7a]">
              CODE RUN
            </span>
            <span className="hidden text-[#1e7a3a] sm:inline">
              — {prompt}
            </span>
          </div>
          <div className="flex items-center gap-3">
            <span className="text-[#1e7a3a]">$</span>
            <select
              value={language}
              onChange={(e) => onLanguageChange(e.target.value as Language)}
              className="border-0 bg-transparent font-mono text-xs text-[#3dff7a] outline-none"
            >
              <option value="javascript">javascript</option>
              <option value="cpp">c++</option>
            </select>
            <button
              type="button"
              onClick={run}
              disabled={running}
              className="font-mono text-xs text-[#3dff7a] hover:underline disabled:opacity-50"
            >
              {running ? "./run &" : "./run"}
            </button>
          </div>
        </header>

        <div className="grid min-h-0 flex-1 grid-rows-2 lg:grid-cols-2 lg:grid-rows-1">
          <div className="relative min-h-0 border-b border-[#1e7a3a] lg:border-b-0 lg:border-r">
            <div className="absolute left-3 top-2 z-10 text-[10px] text-[#1e7a3a]">
              $ cat main.{language === "cpp" ? "cpp" : "js"}
            </div>
            <div className="h-full pt-7">
              <Monaco
                height="100%"
                theme="terminal"
                language={monacoLanguage}
                value={code}
                onChange={(value) => setCode(value ?? "")}
                onMount={onMount}
                options={{
                  minimap: { enabled: false },
                  fontSize: 13,
                  fontFamily: "var(--font-geist-mono), ui-monospace, monospace",
                  automaticLayout: true,
                  tabSize: 2,
                  lineNumbers: "on",
                  renderLineHighlight: "line",
                  cursorBlinking: "blink",
                  cursorStyle: "block",
                  padding: { top: 8 },
                  scrollbar: { verticalScrollbarSize: 8 },
                }}
              />
            </div>
          </div>

          <section className="relative flex min-h-0 flex-col p-4">
            <p className="mb-2 text-[10px] text-[#1e7a3a]">
              $ {prompt}{" "}
              {result?.id ? `# job ${result.id}` : language === "cpp" ? "&& ./a.out" : ""}
            </p>
            <pre
              className={`flex-1 overflow-auto whitespace-pre-wrap text-sm leading-6 ${
                error || result?.error ? "text-[#ff5f56]" : "text-[#3dff7a]"
              }`}
            >
              {outputText}
              {!running && !error ? (
                <span className="cursor-blink ml-0.5 inline-block w-2 bg-[#3dff7a]">
                  &nbsp;
                </span>
              ) : null}
            </pre>
          </section>
        </div>
        <div className="scanlines absolute inset-0" />
      </div>
    </div>
  );
}
