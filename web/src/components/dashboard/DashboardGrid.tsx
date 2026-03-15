import type { DashboardWidget } from "../../features/dashboard/useDashboardData";
import { WidgetRenderer } from "./WidgetRenderer";

export function DashboardGrid({
  widgets,
  widgetData
}: {
  widgets: DashboardWidget[];
  widgetData: Record<string, { series: Array<Record<string, string | number>> }>;
}) {
  return (
    <div className="dashboard-grid wide-card">
      {widgets.map((widget) => (
        <WidgetRenderer key={widget.id} widget={widget} series={widgetData[widget.id]?.series ?? []} />
      ))}
    </div>
  );
}
