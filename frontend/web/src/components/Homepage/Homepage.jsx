import React, { useEffect, useState } from "react";
import "./homepage.css";
import SearchBar from "./SearchBar";
import FilterTag from "./FilterTag";
import ServiceCard from "./ServiceCard";

const API_BASE =
  import.meta.env.VITE_API_URL ||
  import.meta.env.API_URL ||
  `${window.location.protocol}//${window.location.hostname}:8080`;

export default function Homepage() {
  const [workflows, setWorkflows] = useState([]);
  const [selectedWorkflow, setSelectedWorkflow] = useState(null);
  const [payloadPreview, setPayloadPreview] = useState("{}");
  const [panelOpen, setPanelOpen] = useState(false);
  const [showCreate, setShowCreate] = useState(false);
  const [creating, setCreating] = useState(false);
  const [triggering, setTriggering] = useState(false);
  const [loading, setLoading] = useState(false);
  const [showProfile, setShowProfile] = useState(false);
  const userEmail = localStorage.getItem("user_email") || "user@example.com";
  const [form, setForm] = useState({
    name: "Mon Konnect",
    triggerType: "manual",
    intervalMinutes: 5,
    reaction: "discord",
    discordUrl: "",
    emailTo: "",
    emailSubject: "Hello",
    emailBody: "Envoyé depuis Area",
    calSummary: "Nouvel événement",
    calStart: "",
    calEnd: "",
    calAttendees: "",
  });
  const [activeFilters, setActiveFilters] = useState([]);

  useEffect(() => {
    fetchWorkflows();
  }, []);

  useEffect(() => {
    if (selectedWorkflow) {
      setPayloadPreview(
        JSON.stringify(buildPayloadForWorkflow(selectedWorkflow), null, 2)
      );
    }
  }, [selectedWorkflow, form]);

  const filters = [
    { value: "manual", label: "Manual" },
    { value: "interval", label: "Timer" },
    { value: "google", label: "Google" },
    { value: "discord", label: "Discord" },
    { value: "webhook", label: "Webhook" },
  ];

  const toggleFilter = (value) => {
    setActiveFilters((prev) =>
      prev.includes(value) ? prev.filter((v) => v !== value) : [...prev, value]
    );
  };

  const fetchWorkflows = async () => {
    try {
      setLoading(true);
      const res = await fetch(`${API_BASE}/workflows`);
      if (!res.ok) throw new Error("failed to load workflows");
      const data = await res.json();
      setWorkflows(Array.isArray(data) ? data : []);
    } catch (err) {
      console.error(err);
      alert("Impossible de charger les konnects");
    } finally {
      setLoading(false);
    }
  };

  const matchesFilters = (wf) => {
    if (!activeFilters.length) return true;
    return activeFilters.some((f) => {
      const url = (wf.action_url || "").toLowerCase();
      switch (f) {
        case "manual":
          return wf.trigger_type === "manual";
        case "interval":
          return wf.trigger_type === "interval";
        case "google":
          return url.includes("google");
        case "discord":
          return url.includes("discord");
        case "webhook":
          return url.startsWith("http") && !url.includes("google");
        default:
          return true;
      }
    });
  };

  const buildPayloadForWorkflow = (wf) => {
    if (!wf) return {};
    const url = wf.action_url || "";
    if (url.includes("google/email")) {
      return {
        token_id: Number(localStorage.getItem("google_token_id")) || 1,
        to: form.emailTo || "dest@example.com",
        subject: form.emailSubject || "Hello",
        body: form.emailBody || "From Area",
      };
    }
    if (url.includes("google/calendar")) {
      return {
        token_id: Number(localStorage.getItem("google_token_id")) || 1,
        summary: form.calSummary || "Area Event",
        start: form.calStart || new Date().toISOString(),
        end:
          form.calEnd || new Date(Date.now() + 60 * 60 * 1000).toISOString(),
        attendees: form.calAttendees
          ? form.calAttendees.split(",").map((v) => v.trim())
          : [],
      };
    }
    return { content: "Hello from Area" };
  };

  const buildActionUrl = () => {
    switch (form.reaction) {
      case "discord":
        return form.discordUrl || "https://discord.com/api/webhooks/...";
      case "gmail":
        return `${API_BASE}/actions/google/email`;
      case "calendar":
        return `${API_BASE}/actions/google/calendar`;
      default:
        return "";
    }
  };

  const handleCreate = async () => {
    if (!form.name) {
      alert("Nom requis");
      return;
    }
    if (form.reaction === "discord" && !form.discordUrl) {
      alert("URL webhook Discord requise");
      return;
    }
    setCreating(true);
    try {
      const actionUrl = buildActionUrl();
      const body = {
        name: form.name,
        trigger_type: form.triggerType,
        action_url: actionUrl,
        trigger_config:
          form.triggerType === "interval"
            ? { interval_minutes: Number(form.intervalMinutes) || 1 }
            : {},
      };
      const res = await fetch(`${API_BASE}/workflows`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
      });
      if (!res.ok) {
        const text = await res.text();
        throw new Error(text);
      }
      await fetchWorkflows();
      setShowCreate(false);
      setSelectedWorkflow(null);
      setPanelOpen(false);
    } catch (err) {
      console.error(err);
      alert("Création impossible: " + err.message);
    } finally {
      setCreating(false);
    }
  };

  const handleTrigger = async () => {
    if (!selectedWorkflow) {
      alert("Sélectionne un konnect");
      return;
    }
    setTriggering(true);
    try {
      const payload = buildPayloadForWorkflow(selectedWorkflow);
      const res = await fetch(
        `${API_BASE}/workflows/${selectedWorkflow.id}/trigger`,
        {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(payload),
        }
      );
      if (!res.ok) {
        const text = await res.text();
        throw new Error(text);
      }
      alert("Déclenché !");
    } catch (err) {
      console.error(err);
      alert("Echec du déclenchement: " + err.message);
    } finally {
      setTriggering(false);
    }
  };

  return (
    <div className={`layout-root ${panelOpen ? "panel-open" : ""}`}>
      <aside className="sidebar">
        <div className="sidebar-title">Discovery</div>
        <nav className="sidebar-nav">
          <button
            className="primary-btn"
            onClick={() => {
              setShowCreate(true);
              setPanelOpen(true);
              setSelectedWorkflow(null);
            }}
          >
            Create Konnect
          </button>
          <button className="ghost" onClick={fetchWorkflows} disabled={loading}>
            {loading ? "…" : "Refresh"}
          </button>
        </nav>
      </aside>

      <div className={`content-container ${panelOpen ? "panel-open" : ""}`}>
        <main className="main-content">
          <div className="top-row">
            <h1 className="main-title">My Konect</h1>
            <button
              className="profile-btn"
              onClick={() => setShowProfile((p) => !p)}
              aria-label="Profile"
            >
              👤
            </button>
          </div>
          {showProfile && (
            <div className="profile-card">
              <div className="profile-email">{userEmail}</div>
              <button
                className="ghost"
                onClick={() => {
                  localStorage.clear();
                  window.location.href = "/";
                }}
              >
                Logout
              </button>
            </div>
          )}
          <SearchBar />

          <div className="tags-row">
            {filters.map((f) => (
              <FilterTag
                key={f.value}
                label={f.label}
                selected={activeFilters.includes(f.value)}
                onClick={() => toggleFilter(f.value)}
              />
            ))}
          </div>
          <h2 className="section-header">My Konects</h2>
          <div className="services-grid">
            <ServiceCard
              title="Create Konnect"
              color="rgba(0,0,0,0.05)"
              icons={["＋"]}
              ghost
              onClick={() => {
                setShowCreate(true);
                setPanelOpen(true);
                setSelectedWorkflow(null);
              }}
            />
            {workflows.filter(matchesFilters).map((wf, idx) => (
              <ServiceCard
                key={wf.id}
                title={wf.name}
                color={["#00D2FF", "#FF4081", "#00E676", "#D500F9"][idx % 4]}
                onClick={() => {
                  setSelectedWorkflow(wf);
                  setPanelOpen(true);
                  setShowCreate(false);
                }}
              />
            ))}
            {!workflows.length && (
              <div className="muted">No Konnect yet. Create the first one!</div>
            )}
          </div>
        </main>

        {panelOpen && (
          <aside className="right-panel">
            {showCreate ? (
              <>
                <button
                  className="close-btn"
                  onClick={() => {
                    setShowCreate(false);
                    setPanelOpen(false);
                  }}
                >
                  ✖
                </button>
                <h2>Create Konnect</h2>
                <label className="field">
                  <span>Name</span>
                  <input
                    value={form.name}
                    onChange={(e) => setForm({ ...form, name: e.target.value })}
                  />
                </label>
                <label className="field">
                  <span>Trigger</span>
                  <select
                    value={form.triggerType}
                    onChange={(e) =>
                      setForm({ ...form, triggerType: e.target.value })
                    }
                  >
                    <option value="manual">Manual</option>
                    <option value="interval">Timer (minutes)</option>
                  </select>
                </label>
                {form.triggerType === "interval" && (
                  <label className="field">
                    <span>Every (min)</span>
                    <input
                      type="number"
                      min={1}
                      value={form.intervalMinutes}
                      onChange={(e) =>
                        setForm({ ...form, intervalMinutes: e.target.value })
                      }
                    />
                  </label>
                )}
                <label className="field">
                  <span>Reaction</span>
                  <select
                    value={form.reaction}
                    onChange={(e) => setForm({ ...form, reaction: e.target.value })}
                  >
                    <option value="discord">Discord Webhook</option>
                    <option value="gmail">Google Email</option>
                    <option value="calendar">Google Calendar</option>
                  </select>
                </label>
                {form.reaction === "discord" && (
                  <label className="field">
                    <span>Discord webhook URL</span>
                    <input
                      value={form.discordUrl}
                      onChange={(e) =>
                        setForm({ ...form, discordUrl: e.target.value })
                      }
                      placeholder="https://discord.com/api/webhooks/..."
                    />
                  </label>
                )}
                {form.reaction === "gmail" && (
                  <>
                    <label className="field">
                      <span>To</span>
                      <input
                        value={form.emailTo}
                        onChange={(e) =>
                          setForm({ ...form, emailTo: e.target.value })
                        }
                        placeholder="dest@example.com"
                      />
                    </label>
                    <label className="field">
                      <span>Subject</span>
                      <input
                        value={form.emailSubject}
                        onChange={(e) =>
                          setForm({ ...form, emailSubject: e.target.value })
                        }
                      />
                    </label>
                    <label className="field">
                      <span>Body</span>
                      <textarea
                        value={form.emailBody}
                        onChange={(e) =>
                          setForm({ ...form, emailBody: e.target.value })
                        }
                      />
                    </label>
                  </>
                )}
                {form.reaction === "calendar" && (
                  <>
                    <label className="field">
                      <span>Summary</span>
                      <input
                        value={form.calSummary}
                        onChange={(e) =>
                          setForm({ ...form, calSummary: e.target.value })
                        }
                      />
                    </label>
                    <label className="field">
                      <span>Start (ISO)</span>
                      <input
                        value={form.calStart}
                        onChange={(e) =>
                          setForm({ ...form, calStart: e.target.value })
                        }
                        placeholder="2025-12-09T14:00:00Z"
                      />
                    </label>
                    <label className="field">
                      <span>End (ISO)</span>
                      <input
                        value={form.calEnd}
                        onChange={(e) =>
                          setForm({ ...form, calEnd: e.target.value })
                        }
                        placeholder="2025-12-09T15:00:00Z"
                      />
                    </label>
                    <label className="field">
                      <span>Attendees (comma-separated)</span>
                      <input
                        value={form.calAttendees}
                        onChange={(e) =>
                          setForm({ ...form, calAttendees: e.target.value })
                        }
                        placeholder="person@example.com, other@example.com"
                      />
                    </label>
                  </>
                )}
                <button
                  className="primary-btn"
                  onClick={handleCreate}
                  disabled={creating}
                >
                  {creating ? "Creating..." : "Create workflow"}
                </button>
              </>
            ) : selectedWorkflow ? (
              <>
                <button
                  className="close-btn"
                  onClick={() => {
                    setSelectedWorkflow(null);
                    setPanelOpen(false);
                  }}
                >
                  ✖
                </button>
                <h2>{selectedWorkflow?.name}</h2>
                <p className="muted">{selectedWorkflow?.trigger_type}</p>
                <label className="field">
                  <span>Payload</span>
                  <textarea
                    value={payloadPreview}
                    onChange={(e) => setPayloadPreview(e.target.value)}
                    rows={8}
                  />
                </label>
                <button
                  className="primary-btn"
                  onClick={handleTrigger}
                  disabled={triggering}
                >
                  {triggering ? "Triggering…" : "Trigger now"}
                </button>
              </>
            ) : (
              <div className="muted">Sélectionnez un konnect ou créez-en un.</div>
            )}
          </aside>
        )}
      </div>
    </div>
  );
}

function getColor(i) {
  const colors = ["#00D2FF", "#FF4081", "#FF4081", "#00E676", "#D500F9"];
  return colors[i % colors.length];
}
