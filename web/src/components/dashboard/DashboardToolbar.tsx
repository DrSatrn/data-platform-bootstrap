import type { DashboardDefinition } from "../../features/dashboard/useDashboardData";

export function DashboardToolbar({
  activeDashboard,
  dashboards,
  selectedDashboardID,
  canEdit,
  isEditing,
  isSaving,
  saveError,
  selectDashboard,
  createDashboard,
  duplicateDashboard,
  deleteDashboard,
  startEditing,
  cancelEditing,
  saveDashboard
}: {
  activeDashboard: DashboardDefinition | null;
  dashboards: DashboardDefinition[];
  selectedDashboardID: string | null;
  canEdit: boolean;
  isEditing: boolean;
  isSaving: boolean;
  saveError: string | null;
  selectDashboard: (dashboardID: string) => void;
  createDashboard: () => void;
  duplicateDashboard: () => void;
  deleteDashboard: () => Promise<void>;
  startEditing: () => void;
  cancelEditing: () => void;
  saveDashboard: () => Promise<void>;
}) {
  return (
    <div className="hero card wide-card">
      <p className="eyebrow">Saved Reporting Surface</p>
      <div className="row-between">
        <div>
          <h2>{activeDashboard?.name ?? "Dashboard Workspace"}</h2>
          <p className="lede">
            {activeDashboard?.description ??
              "Saved dashboards now drive the reporting experience and can be edited directly from the browser."}
          </p>
        </div>
        <div className="inline-actions">
          <select className="terminal-input compact-input" onChange={(event) => selectDashboard(event.target.value)} value={selectedDashboardID ?? ""}>
            {dashboards.map((item) => (
              <option key={item.id} value={item.id}>
                {item.name}
              </option>
            ))}
          </select>
          <button className="mini-button" disabled={!canEdit} onClick={createDashboard} type="button">
            New dashboard
          </button>
          <button className="mini-button" disabled={!canEdit} onClick={duplicateDashboard} type="button">
            Duplicate
          </button>
          <button className="mini-button" disabled={isSaving || !activeDashboard || !canEdit} onClick={() => void deleteDashboard()} type="button">
            Delete
          </button>
          {!isEditing ? (
            <button className="mini-button" disabled={!canEdit} onClick={startEditing} type="button">
              Edit dashboard
            </button>
          ) : (
            <>
              <button className="mini-button" onClick={cancelEditing} type="button">
                Cancel
              </button>
              <button className="mini-button" disabled={isSaving || !canEdit} onClick={() => void saveDashboard()} type="button">
                {isSaving ? "Saving..." : "Save dashboard"}
              </button>
            </>
          )}
        </div>
      </div>
      {!canEdit ? <p className="muted">Editor token required to create or modify saved dashboards.</p> : null}
      <div className="inline-actions">
        {activeDashboard?.shared_role ? <span className="badge">shared with {activeDashboard.shared_role}+</span> : null}
        {activeDashboard?.owner ? <span className="badge">owner {activeDashboard.owner}</span> : null}
        {(activeDashboard?.tags ?? []).map((tag) => (
          <span className="badge" key={tag}>
            {tag}
          </span>
        ))}
      </div>
      {saveError ? <p className="muted">Save error: {saveError}</p> : null}
    </div>
  );
}
