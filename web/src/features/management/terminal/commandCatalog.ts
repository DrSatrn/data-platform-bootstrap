import type { CommandArea, OperatorRole, TerminalCommandTemplate } from "../types";

export function groupCommandsByArea(commands: TerminalCommandTemplate[]) {
  const grouped = new Map<CommandArea, TerminalCommandTemplate[]>();
  for (const command of commands) {
    const existing = grouped.get(command.area) ?? [];
    existing.push(command);
    grouped.set(command.area, existing);
  }
  return Array.from(grouped.entries()).map(([area, entries]) => ({
    area,
    entries: [...entries].sort((left, right) => left.label.localeCompare(right.label))
  }));
}

export function filterCommands(commands: TerminalCommandTemplate[], query: string) {
  const normalized = query.trim().toLowerCase();
  if (!normalized) {
    return [...commands];
  }
  return commands.filter((command) =>
    [command.label, command.command, command.summary, command.area]
      .join(" ")
      .toLowerCase()
      .includes(normalized)
  );
}

export function roleCanExecute(role: OperatorRole | null | undefined, minimumRole: OperatorRole) {
  return roleRank(role) >= roleRank(minimumRole);
}

export function topCommandSuggestions(
  commands: TerminalCommandTemplate[],
  recentCommands: string[],
  role: OperatorRole | null | undefined,
  limit = 5
) {
  const recentOrder = new Map<string, number>();
  recentCommands.forEach((command, index) => {
    recentOrder.set(command, index);
  });

  return [...commands]
    .filter((command) => roleCanExecute(role, command.minimumRole))
    .sort((left, right) => {
      const leftRecent = recentOrder.get(left.command);
      const rightRecent = recentOrder.get(right.command);
      if (leftRecent !== undefined && rightRecent !== undefined) {
        return leftRecent - rightRecent;
      }
      if (leftRecent !== undefined) {
        return -1;
      }
      if (rightRecent !== undefined) {
        return 1;
      }
      return left.label.localeCompare(right.label);
    })
    .slice(0, limit);
}

function roleRank(role: OperatorRole | null | undefined) {
  switch (role) {
    case "viewer":
      return 1;
    case "editor":
      return 2;
    case "admin":
      return 3;
    default:
      return 0;
  }
}
