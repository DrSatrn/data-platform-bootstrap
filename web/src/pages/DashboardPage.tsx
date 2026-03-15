// DashboardPage renders the reporting-oriented landing view and now includes a
// lightweight dashboard editor. The goal is to make the reporting layer feel
// like a real internal tool where operators can shape the UI, not just consume
// a fixed dashboard.
import { DashboardGrid } from "../components/dashboard/DashboardGrid";
import { DashboardToolbar } from "../components/dashboard/DashboardToolbar";
import { FilterPanel } from "../components/dashboard/FilterPanel";
import { WidgetEditor } from "../components/dashboard/WidgetEditor";
import { ErrorMessage } from "../components/ErrorMessage";
import { LoadingSpinner } from "../components/LoadingSpinner";
import { useAuth } from "../features/auth/useAuth";
import { useDashboardData } from "../features/dashboard/useDashboardData";
import { sortWidgets } from "../components/dashboard/widgetUtils";

export function DashboardPage() {
  const { session } = useAuth();
  const {
    dashboard,
    dashboards,
    draft,
    widgetData,
    viewFilters,
    isEditing,
    isSaving,
    error,
    saveError,
    selectedDashboardID,
    selectedPresetID,
    selectDashboard,
    selectPreset,
    startEditing,
    cancelEditing,
    updateDraft,
    updateDashboardFilter,
    updateViewFilter,
    addPreset,
    removePreset,
    updatePreset,
    updatePresetFilter,
    updateWidget,
    updateWidgetFilter,
    addWidget,
    removeWidget,
    moveWidget,
    nudgeWidget,
    resizeWidget,
    createDashboard,
    duplicateDashboard,
    deleteDashboard,
    exportWidgetCSV,
    saveDashboard
  } = useDashboardData();

  if (error) {
    return <ErrorMessage message={error} title="Dashboard error" />;
  }

  const activeDashboard = isEditing && draft ? draft : dashboard;
  if (!activeDashboard && dashboards.length === 0) {
    return <LoadingSpinner label="Loading dashboards..." />;
  }

  return (
    <section className="page-grid">
      <DashboardToolbar
        activeDashboard={activeDashboard}
        canEdit={Boolean(session?.capabilities.edit_dashboards)}
        cancelEditing={cancelEditing}
        createDashboard={createDashboard}
        dashboards={dashboards}
        deleteDashboard={deleteDashboard}
        duplicateDashboard={duplicateDashboard}
        isEditing={isEditing}
        isSaving={isSaving}
        saveDashboard={saveDashboard}
        saveError={saveError}
        selectDashboard={selectDashboard}
        selectedDashboardID={selectedDashboardID}
        startEditing={startEditing}
      />

      <FilterPanel
        activeDashboard={activeDashboard}
        addPreset={addPreset}
        canEdit={Boolean(session?.capabilities.edit_dashboards)}
        draft={draft}
        isEditing={isEditing}
        removePreset={removePreset}
        selectPreset={selectPreset}
        selectedPresetID={selectedPresetID}
        updateDashboardFilter={updateDashboardFilter}
        updateViewFilter={updateViewFilter}
        updatePreset={updatePreset}
        updatePresetFilter={updatePresetFilter}
        viewFilters={viewFilters}
      />

      {isEditing && draft ? (
        <article className="card wide-card">
          <div className="row-between">
            <h3>Dashboard Editor</h3>
            <button
              className="mini-button"
              disabled={!session?.capabilities.edit_dashboards}
              onClick={addWidget}
              type="button"
            >
              Add widget
            </button>
          </div>
          <div className="form-grid">
            <label className="stack">
              <span className="muted">Dashboard name</span>
              <input className="terminal-input" onChange={(event) => updateDraft("name", event.target.value)} value={draft.name} />
            </label>
            <label className="stack wide-field">
              <span className="muted">Description</span>
              <textarea className="terminal-input" onChange={(event) => updateDraft("description", event.target.value)} rows={3} value={draft.description} />
            </label>
            <label className="stack">
              <span className="muted">Owner</span>
              <input className="terminal-input" onChange={(event) => updateDraft("owner", event.target.value)} value={draft.owner ?? ""} />
            </label>
            <label className="stack">
              <span className="muted">Shared role</span>
              <select className="terminal-input" onChange={(event) => updateDraft("shared_role", event.target.value)} value={draft.shared_role ?? "viewer"}>
                <option value="viewer">viewer</option>
                <option value="editor">editor</option>
                <option value="admin">admin</option>
              </select>
            </label>
            <label className="stack wide-field">
              <span className="muted">Tags</span>
              <input className="terminal-input" onChange={(event) => updateDraft("tags", event.target.value)} value={(draft.tags ?? []).join(", ")} />
            </label>
          </div>
          <div className="stack">
            {draft.widgets.map((widget, index) => (
              <WidgetEditor
                canEdit={Boolean(session?.capabilities.edit_dashboards)}
                index={index}
                key={widget.id}
                moveWidget={moveWidget}
                nudgeWidget={nudgeWidget}
                removeWidget={removeWidget}
                resizeWidget={resizeWidget}
                updateWidget={updateWidget}
                updateWidgetFilter={updateWidgetFilter}
                widget={widget}
              />
            ))}
          </div>
        </article>
      ) : null}

      <DashboardGrid exportWidgetCSV={exportWidgetCSV} widgetData={widgetData} widgets={sortWidgets(activeDashboard?.widgets ?? [])} />
    </section>
  );
}
