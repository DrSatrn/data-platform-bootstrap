import type { DashboardWidget } from "../../features/dashboard/useDashboardData";
import { WidgetRenderer } from "./WidgetRenderer";

export function DashboardGrid({
  widgets,
  widgetData,
  exportWidgetCSV
}: {
  widgets: DashboardWidget[];
  widgetData: Record<string, { series: Array<Record<string, string | number>> }>;
  exportWidgetCSV: (widgetID: string) => Promise<void>;
}) {
  return (
    <div className="dashboard-grid wide-card">
      {widgets.map((widget) => (
        <WidgetRenderer key={widget.id} widget={widget} onExportCSV={exportWidgetCSV} series={widgetData[widget.id]?.series ?? []} />
      ))}
    </div>
  );
}
