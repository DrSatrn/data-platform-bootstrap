import type { DashboardWidget } from "../../features/dashboard/useDashboardData";
import { axisOptions, datasetOptions, layoutOf, metricOptions } from "./widgetUtils";

export function WidgetEditor({
  canEdit,
  index,
  widget,
  moveWidget,
  nudgeWidget,
  resizeWidget,
  removeWidget,
  updateWidget,
  updateWidgetFilter
}: {
  canEdit: boolean;
  index: number;
  widget: DashboardWidget;
  moveWidget: (widgetID: string, direction: -1 | 1) => void;
  nudgeWidget: (widgetID: string, axis: "x" | "y", delta: -1 | 1) => void;
  resizeWidget: (widgetID: string, axis: "w" | "h", delta: -1 | 1) => void;
  removeWidget: (widgetID: string) => void;
  updateWidget: (widgetID: string, field: keyof DashboardWidget, value: string | number | string[]) => void;
  updateWidgetFilter: (widgetID: string, field: "from_month" | "to_month" | "category", value: string) => void;
}) {
  return (
    <div className="subcard">
      <div className="row-between">
        <strong>{widget.name || `Widget ${index + 1}`}</strong>
        <div className="inline-actions">
          <button className="mini-button" disabled={!canEdit} onClick={() => moveWidget(widget.id, -1)} type="button">
            Up
          </button>
          <button className="mini-button" disabled={!canEdit} onClick={() => moveWidget(widget.id, 1)} type="button">
            Down
          </button>
          <button className="mini-button" disabled={!canEdit} onClick={() => nudgeWidget(widget.id, "x", -1)} type="button">
            Left
          </button>
          <button className="mini-button" disabled={!canEdit} onClick={() => nudgeWidget(widget.id, "x", 1)} type="button">
            Right
          </button>
          <button className="mini-button" disabled={!canEdit} onClick={() => resizeWidget(widget.id, "w", -1)} type="button">
            Narrower
          </button>
          <button className="mini-button" disabled={!canEdit} onClick={() => resizeWidget(widget.id, "w", 1)} type="button">
            Wider
          </button>
          <button className="mini-button" disabled={!canEdit} onClick={() => resizeWidget(widget.id, "h", -1)} type="button">
            Shorter
          </button>
          <button className="mini-button" disabled={!canEdit} onClick={() => resizeWidget(widget.id, "h", 1)} type="button">
            Taller
          </button>
          <button className="mini-button" disabled={!canEdit} onClick={() => removeWidget(widget.id)} type="button">
            Remove
          </button>
        </div>
      </div>
      <div className="widget-editor-grid">
        <label className="stack">
          <span className="muted">Widget name</span>
          <input className="terminal-input" onChange={(event) => updateWidget(widget.id, "name", event.target.value)} value={widget.name} />
        </label>
        <label className="stack">
          <span className="muted">Type</span>
          <select className="terminal-input" onChange={(event) => updateWidget(widget.id, "type", event.target.value)} value={widget.type}>
            <option value="kpi">KPI</option>
            <option value="table">Table</option>
            <option value="line">Line</option>
            <option value="bar">Bar</option>
          </select>
        </label>
        <label className="stack">
          <span className="muted">Dataset</span>
          <select className="terminal-input" onChange={(event) => updateWidget(widget.id, "dataset_ref", event.target.value)} value={widget.dataset_ref ?? ""}>
            <option value="">None</option>
            {datasetOptions.map((option) => (
              <option key={option} value={option}>
                {option}
              </option>
            ))}
          </select>
        </label>
        <label className="stack">
          <span className="muted">Metric</span>
          <select className="terminal-input" onChange={(event) => updateWidget(widget.id, "metric_ref", event.target.value)} value={widget.metric_ref ?? ""}>
            <option value="">None</option>
            {metricOptions.map((option) => (
              <option key={option} value={option}>
                {option}
              </option>
            ))}
          </select>
        </label>
        <label className="stack">
          <span className="muted">Value field</span>
          <input className="terminal-input" onChange={(event) => updateWidget(widget.id, "value_field", event.target.value)} value={widget.value_field ?? ""} />
        </label>
        <label className="stack">
          <span className="muted">X axis</span>
          <select className="terminal-input" onChange={(event) => updateWidget(widget.id, "x_axis", event.target.value)} value={widget.x_axis ?? "month"}>
            {axisOptions.map((option) => (
              <option key={option} value={option}>
                {option}
              </option>
            ))}
          </select>
        </label>
        <label className="stack">
          <span className="muted">Y axis</span>
          <select className="terminal-input" onChange={(event) => updateWidget(widget.id, "y_axis", event.target.value)} value={widget.y_axis ?? "net_cashflow"}>
            {axisOptions.map((option) => (
              <option key={option} value={option}>
                {option}
              </option>
            ))}
          </select>
        </label>
        <label className="stack">
          <span className="muted">Limit</span>
          <input
            className="terminal-input"
            min={1}
            onChange={(event) => updateWidget(widget.id, "limit", Number(event.target.value) || 0)}
            type="number"
            value={widget.limit ?? 12}
          />
        </label>
        <label className="stack">
          <span className="muted">Group by</span>
          <input
            className="terminal-input"
            onChange={(event) =>
              updateWidget(
                widget.id,
                "group_by",
                event.target.value
                  .split(",")
                  .map((item) => item.trim())
                  .filter(Boolean)
              )
            }
            placeholder="month, category"
            value={(widget.group_by ?? []).join(", ")}
          />
        </label>
        <label className="stack wide-field">
          <span className="muted">Description</span>
          <input className="terminal-input" onChange={(event) => updateWidget(widget.id, "description", event.target.value)} value={widget.description ?? ""} />
        </label>
        <label className="stack">
          <span className="muted">From month</span>
          <input className="terminal-input" onChange={(event) => updateWidgetFilter(widget.id, "from_month", event.target.value)} placeholder="YYYY-MM" value={widget.filters?.from_month ?? ""} />
        </label>
        <label className="stack">
          <span className="muted">To month</span>
          <input className="terminal-input" onChange={(event) => updateWidgetFilter(widget.id, "to_month", event.target.value)} placeholder="YYYY-MM" value={widget.filters?.to_month ?? ""} />
        </label>
        <label className="stack">
          <span className="muted">Category filter</span>
          <input className="terminal-input" onChange={(event) => updateWidgetFilter(widget.id, "category", event.target.value)} placeholder="Category" value={widget.filters?.category ?? ""} />
        </label>
      </div>
      <p className="muted">
        Grid position col {layoutOf(widget).x + 1}, row {layoutOf(widget).y + 1}, span {layoutOf(widget).w} x {layoutOf(widget).h}
      </p>
    </div>
  );
}
