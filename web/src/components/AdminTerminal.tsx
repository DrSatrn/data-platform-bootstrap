// AdminTerminal renders the built-in platform management console. The goal is
// to give operators a productive command surface inside the UI without exposing
// arbitrary host shell access.
import { FormEvent, useState } from "react";

import { useAdminTerminal } from "../features/system/useAdminTerminal";

const suggestedCommands = ["help", "status", "pipelines", "assets", "quality", "metrics", "logs 5"];

export function AdminTerminal() {
  const [command, setCommand] = useState("status");
  const { entries, pending, error, execute, canExecute } = useAdminTerminal();

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    void execute(command);
    setCommand("");
  }

  return (
    <section className="card wide-card">
      <div className="row-between">
        <div>
          <h2>Admin Terminal</h2>
          <p className="muted">
            This terminal executes built-in platform management commands over the admin API.
          </p>
        </div>
        <span className="badge">{canExecute ? "Admin access" : "Admin token required"}</span>
      </div>
      <div className="command-row">
        {suggestedCommands.map((item) => (
          <button key={item} className="mini-button" onClick={() => setCommand(item)} type="button">
            {item}
          </button>
        ))}
      </div>
      <form className="terminal-form" onSubmit={handleSubmit}>
        <input
          aria-label="Terminal command"
          className="terminal-input"
          onChange={(event) => setCommand(event.target.value)}
          placeholder="Enter a command"
          value={command}
        />
        <button className="nav-button active terminal-submit" disabled={pending || !canExecute} type="submit">
          {pending ? "Running..." : "Run"}
        </button>
      </form>
      {!canExecute ? (
        <p className="muted">Set an admin-capable bearer token in the sidebar to unlock terminal commands.</p>
      ) : null}
      {error ? <p className="muted">Last terminal error: {error}</p> : null}
      <div className="terminal-output">
        {entries.map((entry, index) => (
          <div className="terminal-entry" key={`${entry.command}-${index}`}>
            <p className="terminal-prompt">$ {entry.command}</p>
            {entry.output.map((line, lineIndex) => (
              <p className={entry.success ? "terminal-line" : "terminal-line terminal-error"} key={lineIndex}>
                {line}
              </p>
            ))}
          </div>
        ))}
      </div>
    </section>
  );
}
