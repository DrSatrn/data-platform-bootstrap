// This hook drives the built-in admin terminal used in the management portal.
// The terminal talks to the platform's own command API rather than opening a
// shell on the host, which keeps the surface safer and easier to audit.
import { useState } from "react";

import { useAuth } from "../auth/useAuth";
import { postJSON } from "../../lib/api";

type TerminalResponse = {
  command: string;
  success: boolean;
  output: string[];
};

type TerminalEntry = {
  command: string;
  success: boolean;
  output: string[];
};

export function useAdminTerminal() {
  const { token, session } = useAuth();
  const [entries, setEntries] = useState<TerminalEntry[]>([
    {
      command: "help",
      success: true,
      output: ["Built-in platform terminal ready. Try: help, status, pipelines, assets, quality, metrics, logs 5"]
    }
  ]);
  const [pending, setPending] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function execute(command: string) {
    setPending(true);
    setError(null);
    try {
      if (!session?.capabilities.run_admin_terminal) {
        throw new Error("Admin role required to run terminal commands.");
      }
      const result = await postJSON<TerminalResponse, { command: string }>(
        "/api/v1/admin/terminal/execute",
        { command },
        token.trim() || undefined
      );
      setEntries((current) => current.concat([{ command: result.command, success: result.success, output: result.output }]));
    } catch (nextError) {
      const message = nextError instanceof Error ? nextError.message : "Unknown terminal error";
      setError(message);
      setEntries((current) => current.concat([{ command, success: false, output: [message] }]));
    } finally {
      setPending(false);
    }
  }

  return { entries, pending, error, execute, canExecute: session?.capabilities.run_admin_terminal ?? false };
}
