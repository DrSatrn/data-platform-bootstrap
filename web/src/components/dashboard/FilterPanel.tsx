import type { DashboardDefinition, DashboardPreset } from "../../features/dashboard/useDashboardData";

export function FilterPanel({
  activeDashboard,
  selectedPresetID,
  selectPreset,
  isEditing,
  draft,
  canEdit,
  updateDashboardFilter,
  addPreset,
  removePreset,
  updatePreset,
  updatePresetFilter
}: {
  activeDashboard: DashboardDefinition | null;
  selectedPresetID: string | null;
  selectPreset: (presetID: string) => void;
  isEditing: boolean;
  draft: DashboardDefinition | null;
  canEdit: boolean;
  updateDashboardFilter: (field: "from_month" | "to_month" | "category", value: string) => void;
  addPreset: () => void;
  removePreset: (presetID: string) => void;
  updatePreset: (presetID: string, field: "name" | "description", value: string) => void;
  updatePresetFilter: (presetID: string, field: "from_month" | "to_month" | "category", value: string) => void;
}) {
  return (
    <>
      <article className="card wide-card">
        <div className="row-between">
          <h3>Report Context</h3>
          <div className="inline-actions">
            <select className="terminal-input compact-input" onChange={(event) => selectPreset(event.target.value)} value={selectedPresetID ?? ""}>
              <option value="">No preset</option>
              {(activeDashboard?.presets ?? []).map((preset) => (
                <option key={preset.id} value={preset.id}>
                  {preset.name}
                </option>
              ))}
            </select>
          </div>
        </div>
        <p className="muted">
          Dashboard-wide filters apply before widget-specific filters so teams can reuse one saved layout across multiple reporting contexts.
        </p>
        <div className="form-grid">
          <label className="stack">
            <span className="muted">Default from month</span>
            <input className="terminal-input" disabled={!isEditing} onChange={(event) => updateDashboardFilter("from_month", event.target.value)} placeholder="YYYY-MM" value={activeDashboard?.default_filters?.from_month ?? ""} />
          </label>
          <label className="stack">
            <span className="muted">Default to month</span>
            <input className="terminal-input" disabled={!isEditing} onChange={(event) => updateDashboardFilter("to_month", event.target.value)} placeholder="YYYY-MM" value={activeDashboard?.default_filters?.to_month ?? ""} />
          </label>
          <label className="stack">
            <span className="muted">Default category</span>
            <input className="terminal-input" disabled={!isEditing} onChange={(event) => updateDashboardFilter("category", event.target.value)} placeholder="Food" value={activeDashboard?.default_filters?.category ?? ""} />
          </label>
        </div>
      </article>

      {isEditing && draft ? (
        <article className="card wide-card">
          <div className="row-between">
            <h4>Preset Library</h4>
            <button className="mini-button" disabled={!canEdit} onClick={addPreset} type="button">
              Add preset
            </button>
          </div>
          <div className="stack">
            {(draft.presets ?? []).map((preset) => (
              <PresetEditor
                canEdit={canEdit}
                key={preset.id}
                preset={preset}
                removePreset={removePreset}
                updatePreset={updatePreset}
                updatePresetFilter={updatePresetFilter}
              />
            ))}
          </div>
        </article>
      ) : null}
    </>
  );
}

function PresetEditor({
  canEdit,
  preset,
  removePreset,
  updatePreset,
  updatePresetFilter
}: {
  canEdit: boolean;
  preset: DashboardPreset;
  removePreset: (presetID: string) => void;
  updatePreset: (presetID: string, field: "name" | "description", value: string) => void;
  updatePresetFilter: (presetID: string, field: "from_month" | "to_month" | "category", value: string) => void;
}) {
  return (
    <div className="subcard">
      <div className="row-between">
        <strong>{preset.name}</strong>
        <button className="mini-button" disabled={!canEdit} onClick={() => removePreset(preset.id)} type="button">
          Remove preset
        </button>
      </div>
      <div className="form-grid">
        <label className="stack">
          <span className="muted">Preset name</span>
          <input className="terminal-input" onChange={(event) => updatePreset(preset.id, "name", event.target.value)} value={preset.name} />
        </label>
        <label className="stack wide-field">
          <span className="muted">Description</span>
          <input className="terminal-input" onChange={(event) => updatePreset(preset.id, "description", event.target.value)} value={preset.description ?? ""} />
        </label>
        <label className="stack">
          <span className="muted">From month</span>
          <input className="terminal-input" onChange={(event) => updatePresetFilter(preset.id, "from_month", event.target.value)} placeholder="YYYY-MM" value={preset.filters?.from_month ?? ""} />
        </label>
        <label className="stack">
          <span className="muted">To month</span>
          <input className="terminal-input" onChange={(event) => updatePresetFilter(preset.id, "to_month", event.target.value)} placeholder="YYYY-MM" value={preset.filters?.to_month ?? ""} />
        </label>
        <label className="stack">
          <span className="muted">Category</span>
          <input className="terminal-input" onChange={(event) => updatePresetFilter(preset.id, "category", event.target.value)} placeholder="Food" value={preset.filters?.category ?? ""} />
        </label>
      </div>
    </div>
  );
}
