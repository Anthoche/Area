import React, { useEffect, useMemo, useState } from "react";
import { Link } from "react-router-dom";
import "./homepage.css";
import SearchBar from "./SearchBar.jsx";
import FilterTag from "./FilterTag.jsx";
import ServiceCard from "./ServiceCard.jsx";
import user from "../../../../lib/assets/user.png";

const API_BASE =
    import.meta.env.VITE_API_URL ||
    import.meta.env.API_URL ||
    `${window.location.protocol}//${window.location.hostname}:8080`;

export default function Homepage() {
    const [workflows, setWorkflows] = useState([]);
    const [areas, setAreas] = useState([]);
    const [triggers, setTriggers] = useState([]);
    const [reactions, setReactions] = useState([]);
    const [selectedReaction, setSelectedReaction] = useState("");
    const [selectedWorkflow, setSelectedWorkflow] = useState(null);
    const [payloadPreview, setPayloadPreview] = useState("{}");
    const [panelOpen, setPanelOpen] = useState(false);
    const [showCreate, setShowCreate] = useState(false);
    const [searchTerm, setSearchTerm] = useState("");
    const [creating, setCreating] = useState(false);
    const [triggering, setTriggering] = useState(false);
    const [togglingTimer, setTogglingTimer] = useState(false);
    const [deleting, setDeleting] = useState(false);
    const [loading, setLoading] = useState(false);
    const [showProfile, setShowProfile] = useState(false);
    const getUserId = () => Number(localStorage.getItem("user_id") || "");
    const userEmail = localStorage.getItem("user_email") || "user@example.com";
    const [form, setForm] = useState({
        name: "Mon Konect",
        triggerType: "",
        triggerValues: {},
        values: {},
    });
    const [activeFilters, setActiveFilters] = useState([]);
    const [existingIds, setExistingIds] = useState([]);

  const cards = workflows.map((wf) => {
    const isNew = !existingIds.includes(wf.id) && !loading;
    return { ...wf, created: isNew };
  });

   useEffect(() => {
    const ids = workflows.map((wf) => wf.id);
    setExistingIds(ids);
  }, [workflows]);


  const selectedReactionDef = useMemo(
    () => reactions.find((r) => r.id === selectedReaction),
    [reactions, selectedReaction]
  );

  const triggerFields = useMemo(() => {
    const trig = triggers.find((t) => t.id === form.triggerType);
    return trig?.fields || [];
  }, [triggers, form.triggerType]);

  const reactionFields = selectedReactionDef?.fields || [];

  const defaultValuesFromFields = (fields, hint) => {
    const googleToken = localStorage.getItem("google_token_id");
    const githubToken = localStorage.getItem("github_token_id");
    return (
      fields?.reduce((acc, f) => {
        if (f.key === "token_id") {
          if (hint?.startsWith("google_") && googleToken) {
            acc[f.key] = Number(googleToken);
          } else if (hint?.startsWith("github") && githubToken) {
            acc[f.key] = Number(githubToken);
          } else if (googleToken) {
            acc[f.key] = Number(googleToken);
          } else if (githubToken) {
            acc[f.key] = Number(githubToken);
          } else {
            acc[f.key] = f.example || "";
          }
        } else if (f.type === "number") {
          acc[f.key] =
            f.example !== undefined && f.example !== null
              ? Number(f.example)
              : 0;
        } else {
          acc[f.key] = f.example || "";
        }
        return acc;
      }, {}) || {}
    );
  };

  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    const tokenId = params.get("token_id");
    const googleEmail = params.get("google_email");
    const githubLogin = params.get("github_login");
    const userIdFromQuery = params.get("user_id");
    if (tokenId && (googleEmail || params.get("google_email"))) {
      localStorage.setItem("google_token_id", tokenId);
    } else if (tokenId && (githubLogin || params.get("github_email"))) {
      localStorage.setItem("github_token_id", tokenId);
    }
    if (googleEmail) {
      localStorage.setItem("google_email", googleEmail);
      localStorage.setItem("user_email", googleEmail);
    }
    if (githubLogin) {
      localStorage.setItem("github_login", githubLogin);
    }
    if (userIdFromQuery) {
      localStorage.setItem("user_id", userIdFromQuery);
    }
    if (tokenId || googleEmail || githubLogin) {
      window.history.replaceState({}, document.title, window.location.pathname);
    }
    fetchAreas().then(() => fetchWorkflows());
  }, []);

  useEffect(() => {
    if (selectedWorkflow) {
      setPayloadPreview(
        JSON.stringify(buildPayloadForWorkflow(selectedWorkflow), null, 2)
      );
    }
  }, [selectedWorkflow, form, selectedReaction]);

  useEffect(() => {
    const colorOptions = [
      "linear-gradient(135deg, #00d0ffc1 0%, #b2f1ffe4 100%)",
      "linear-gradient(135deg, #FF4081 0%, rgba(255, 144, 144, 1) 100%)",
      "linear-gradient(135deg, #00E676 0%, #86ffc4ff 100%)",
      "linear-gradient(135deg, #D500F9 0%, #cfa8d6ff 100%)",
    ];

    setWorkflows((prev) =>
      prev.map((wf, idx) => ({
        ...wf,
        cardColor: wf.cardColor || colorOptions[idx % 4],
      }))
    );
  }, [workflows]);

  useEffect(() => {
    if (!form.triggerType && triggers.length) {
      const first = triggers[0];
      const defaults = defaultValuesFromFields(first.fields || [], first.id);
      setForm((prev) => ({
        ...prev,
        triggerType: first.id,
        triggerValues: defaults,
      }));
    }
  }, [triggers, form.triggerType]);

  const triggerFilterOptions = useMemo(
    () => triggers.map((t) => ({ value: t.id, label: t.name })),
    [triggers]
  );
  const filters = triggerFilterOptions;

  const toggleFilter = (value) => {
    setActiveFilters((prev) =>
      prev.includes(value) ? prev.filter((v) => v !== value) : [...prev, value]
    );
  };

  const fetchWorkflows = async () => {
    try {
      setLoading(true);
      const userId = getUserId();
      if (!userId) {
        throw new Error("missing user id, please login again");
      }
      const res = await fetch(`${API_BASE}/workflows`, {
        headers: { "X-User-ID": String(userId) },
      });
      if (!res.ok) throw new Error("failed to load workflows");
      const data = await res.json();
      const list = Array.isArray(data) ? data : [];
      setWorkflows(list);
      return list;
    } catch (err) {
      console.error(err);
      alert("Impossible de charger les konects");
      return [];
    } finally {
      setLoading(false);
    }
  };

  const fetchAreas = async () => {
    try {
      const res = await fetch(`${API_BASE}/areas`);
      if (!res.ok) throw new Error("failed to load areas");
      const data = await res.json();
      const services = Array.isArray(data.services) ? data.services : [];
      setAreas(services);
      const triggerCaps =
        services
          .find((s) => s.id === "core")
          ?.triggers?.map((t) => ({
            id: t.id,
            name: t.name,
            description: t.description,
            fields: t.fields || [],
          })) || [];
      setTriggers(triggerCaps);
      const reactionCaps = services
        .filter((s) => s.enabled !== false)
        .flatMap((s) =>
          (s.reactions || []).map((r) => ({
            id: r.id,
            name: r.name,
            description: r.description,
            action_url: r.action_url,
            default_payload: r.default_payload,
            service: s.name || s.id,
            fields: r.fields || [],
          }))
        );
      setReactions(reactionCaps);
      if (reactionCaps.length > 0) {
        setSelectedReaction(reactionCaps[0].id);
        const defaults = defaultValuesFromFields(reactionCaps[0].fields || []);
        setForm((prev) => ({ ...prev, values: defaults }));
      }
      if (triggerCaps.length && !form.triggerType) {
        const defaults = defaultValuesFromFields(triggerCaps[0].fields || []);
        setForm((prev) => ({
          ...prev,
          triggerType: triggerCaps[0].id,
          triggerValues: defaults,
        }));
      }
    } catch (err) {
      console.error(err);
      setAreas([]);
      setTriggers([]);
      setReactions([]);
    }
  };

  const matchesFilters = (wf) => {
    if (!activeFilters.length) return true;
    return activeFilters.some((f) => wf.trigger_type === f);
  };

  const buildPayloadForWorkflow = (wf) => {
    if (!wf) return {};
    const payload = { ...(form.values || {}) };
    return payload;
  };

  const buildIntervalPayload = () => {
    return form.values || {};
  };

  const buildActionUrl = () => {
    const actionUrl = selectedReactionDef?.action_url || "";
    if (actionUrl.startsWith("http")) return actionUrl;
    if (actionUrl.startsWith("/")) return `${API_BASE}${actionUrl}`;
    // Fallback for webhook: use provided URL field
    if ((selectedReaction || "").includes("webhook") && form.values?.webhook_url) {
      return form.values.webhook_url;
    }
    return actionUrl;
  };

  const handleCreate = async () => {
    if (!form.name) {
      alert("Nom requis");
      return;
    }
    const userId = getUserId();
    if (!userId) {
      alert("Merci de vous reconnecter (user id manquant)");
      return;
    }
    const requiredFields = reactionFields.filter((f) => f.required);
    for (const f of requiredFields) {
      if (!form.values || form.values[f.key] === undefined || form.values[f.key] === "") {
        alert(`Champ requis: ${f.key}`);
        return;
      }
    }
    setCreating(true);
    try {
      const actionUrl = buildActionUrl();
      const triggerType =
        form.triggerType || (triggers.length ? triggers[0].id : "");
      const triggerConfig = {
        ...form.triggerValues,
        payload: buildIntervalPayload(),
        payload_template: form.values || {},
      };
      const body = {
        name: form.name,
        trigger_type: triggerType,
        action_url: actionUrl,
        trigger_config: triggerConfig,
      };
      const res = await fetch(`${API_BASE}/workflows`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "X-User-ID": String(userId),
        },
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
      alert("Sélectionne un konect");
      return;
    }
    const userId = getUserId();
    if (!userId) {
      alert("Merci de vous reconnecter (user id manquant)");
      return;
    }
    setTriggering(true);
    try {
      const payload = buildPayloadForWorkflow(selectedWorkflow);
      const res = await fetch(
        `${API_BASE}/workflows/${selectedWorkflow.id}/trigger`,
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            "X-User-ID": String(userId),
          },
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

  const handleDelete = async () => {
    if (!selectedWorkflow) return;
    if (!window.confirm("Delete this Konect?")) return;
    const userId = getUserId();
    if (!userId) {
      alert("Merci de vous reconnecter (user id manquant)");
      return;
    }
    setDeleting(true);
    try {
      const res = await fetch(`${API_BASE}/workflows/${selectedWorkflow.id}`, {
        method: "DELETE",
        headers: { "X-User-ID": String(userId) },
      });
      if (!res.ok) {
        const text = await res.text();
        throw new Error(text);
      }
      await fetchWorkflows();
      setSelectedWorkflow(null);
      setPanelOpen(false);
    } catch (err) {
      console.error(err);
      alert("Suppression impossible: " + err.message);
    } finally {
      setDeleting(false);
    }
  };

  const handleToggleTimer = async () => {
    if (!selectedWorkflow) return;
    const action = selectedWorkflow.enabled ? "disable" : "enable";
    const userId = getUserId();
    if (!userId) {
      alert("Merci de vous reconnecter (user id manquant)");
      return;
    }
    setTogglingTimer(true);
    try {
      const res = await fetch(
        `${API_BASE}/workflows/${selectedWorkflow.id}/enabled?action=${action}`,
        { method: "POST", headers: { "X-User-ID": String(userId) } }
      );
      if (!res.ok) {
        const text = await res.text();
        throw new Error(text);
      }
      const list = await fetchWorkflows();
      const refreshed = list.find((w) => w.id === selectedWorkflow.id) || null;
      setSelectedWorkflow(refreshed);
    } catch (err) {
      console.error(err);
      alert("Action timer impossible: " + err.message);
    } finally {
      setTogglingTimer(false);
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
            Create a Konect
          </button>
          <button className="ghost" onClick={fetchWorkflows} disabled={loading}>
            {loading ? "…" : "Refresh"}
          </button>
        </nav>
      </aside>
      <div className={`content-container ${panelOpen ? "panel-open" : ""}`}>
        <main className="main-content">
          <div className="konect-hero">
            <Link className="hero-back" to="/">
              {"<"} Back to Welcome Page
            </Link>
            <button
              className="profile-btn hero-profile"
              onClick={() => setShowProfile((p) => !p)}
              aria-label="Profile"
            >
              <img
                src={user}
                alt="User Profile"
                style={{
                  width: "80%",
                  height: "80%",
                  borderRadius: "50%",
                  objectFit: "cover",
                }}
              />
            </button>

            <div className="konect-hero-content">
              <div className="konect-hero-left">
                <h1 className="konect-title">My Konect</h1>
                <p className="konect-subtitle">
                  Manage and automate your favorite services seamlessly.
                  Create and organize your Konects to boost productivity.
                </p>
              </div>

          <div className="konect-hero-right">
            <button
              className="create-konect-btn"
              onClick={() => {
                setShowCreate(true);
                setPanelOpen(true);
                setSelectedWorkflow(null);
              }}
            >
              + Create a Konect
            </button>
          </div>
        </div>
     </div>
          <div className="section-card">
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
            <SearchBar
              value={searchTerm}
              onChange={setSearchTerm}
              placeholder="🔎  Search a Konect"
            />
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
          </div>

          <div className="section-card">
            <h2 className="section-header centered">My Konects</h2>
            <div className="services-grid">
              {workflows
                .filter(matchesFilters)
                .filter((wf) =>
                  (wf.name || "").toLowerCase().includes(searchTerm.trim().toLowerCase())
                )
                .map((wf) => {
                  const created = !existingIds.includes(wf.id) && !loading;
                  const deleted = deleting && selectedWorkflow && selectedWorkflow.id === wf.id;
                  return (
                    <ServiceCard
                      key={wf.id}
                      title={wf.name}
                      color={wf.cardColor}
                      created={created}
                      deleted={deleted}
                      onClick={() => {
                        setSelectedWorkflow(wf);
                        setPanelOpen(true);
                        setShowCreate(false);
                      }}
                    />
                  );
                })}
              {!workflows.length && (
                <div className="muted">No Konect created yet. Create the first one!</div>
              )}
            </div>
          </div>
        </main>

        {panelOpen && (
          <>
            <div
              className="panel-backdrop"
              onClick={() => {
                setPanelOpen(false);
                setShowCreate(false);
                setSelectedWorkflow(null);
              }}
            />
            <aside className="right-panel">
              <div className="panel-inner">
                {showCreate ? (
                  <>
                    <div className="panel-header">
                      <div>
                        <div className="panel-kicker">Create</div>
                        <h2>Create a Konect</h2>
                        <div className="panel-meta-row">
                          <span className="panel-chip">Trigger + Reaction</span>
                        </div>
                      </div>
                      <button
                        className="panel-close"
                        onClick={() => {
                          setShowCreate(false);
                          setPanelOpen(false);
                        }}
                      >
                        x
                      </button>
                    </div>
                    <div className="panel-body">
                      <div className="panel-section">
                        <div className="panel-section-title">Basics</div>
                        <label className="field">
                          <span>Name</span>
                          <input
                            value={form.name}
                            onChange={(e) =>
                              setForm({ ...form, name: e.target.value })
                            }
                          />
                        </label>
                      </div>
                      <div className="panel-section">
                        <div className="panel-section-title">Trigger</div>
                        <label className="field">
                          <span>Trigger</span>
                          <select
                            value={form.triggerType}
                            onChange={(e) =>
                              setForm((prev) => {
                                const trig = triggers.find(
                                  (t) => t.id === e.target.value
                                );
                                const defaults = defaultValuesFromFields(
                                  trig?.fields || [],
                                  trig?.id
                                );
                                return {
                                  ...prev,
                                  triggerType: e.target.value,
                                  triggerValues: defaults,
                                };
                              })
                            }
                          >
                            {triggers.length ? (
                              triggers.map((t) => (
                                <option key={t.id} value={t.id}>
                                  {t.name}
                                </option>
                              ))
                            ) : (
                              <option value="">Loading triggers…</option>
                            )}
                          </select>
                          <div className="muted">
                            {
                              triggers.find((t) => t.id === form.triggerType)
                                ?.description
                            }
                          </div>
                        </label>
                        {triggerFields
                          .filter((f) => f.key !== "token_id")
                          .map((field) => (
                            <label className="field" key={field.key}>
                              <span>
                                {field.key} {field.required ? "*" : ""}
                              </span>
                              <input
                                type={field.type === "number" ? "number" : "text"}
                                value={form.triggerValues?.[field.key] ?? ""}
                                placeholder={
                                  field.example
                                    ? String(field.example)
                                    : field.description || ""
                                }
                                onChange={(e) =>
                                  setForm((prev) => ({
                                    ...prev,
                                    triggerValues: {
                                      ...prev.triggerValues,
                                      [field.key]:
                                        field.type === "number"
                                          ? Number(e.target.value)
                                          : e.target.value,
                                    },
                                  }))
                                }
                              />
                              <div className="muted">{field.description}</div>
                            </label>
                          ))}
                      </div>
                      <div className="panel-section">
                        <div className="panel-section-title">Reaction</div>
                        <label className="field">
                          <span>Reaction</span>
                          <select
                            value={selectedReaction || form.reaction}
                            onChange={(e) => {
                              const nextId = e.target.value;
                              setSelectedReaction(nextId);
                              const next = reactions.find((r) => r.id === nextId);
                              const defaults = defaultValuesFromFields(
                                next?.fields || [],
                                next?.id
                              );
                              setForm((prev) => ({ ...prev, values: defaults }));
                            }}
                          >
                            {reactions.length ? (
                              reactions.map((r) => (
                                <option key={r.id} value={r.id}>
                                  {r.service} - {r.name}
                                </option>
                              ))
                            ) : (
                              <option value="">Loading reactions…</option>
                            )}
                          </select>
                          <div className="muted">
                            {
                              reactions.find((r) => r.id === selectedReaction)
                                ?.description
                            }
                          </div>
                        </label>
                        {reactionFields
                          .filter((f) => f.key !== "token_id")
                          .map((field) => (
                            <label className="field" key={field.key}>
                              <span>
                                {field.key} {field.required ? "*" : ""}
                              </span>
                              <input
                                type={field.type === "number" ? "number" : "text"}
                                value={form.values?.[field.key] ?? ""}
                                placeholder={
                                  field.example
                                    ? String(field.example)
                                    : field.description || ""
                                }
                                onChange={(e) =>
                                  setForm((prev) => ({
                                    ...prev,
                                    values: {
                                      ...prev.values,
                                      [field.key]:
                                        field.type === "number"
                                          ? Number(e.target.value)
                                          : e.target.value,
                                    },
                                  }))
                                }
                              />
                              <div className="muted">{field.description}</div>
                            </label>
                          ))}
                      </div>
                    </div>
                    <div className="panel-actions">
                      <button
                        className="ghost"
                        onClick={() => {
                          setShowCreate(false);
                          setPanelOpen(false);
                        }}
                      >
                        Cancel
                      </button>
                      <button
                        className="primary-btn"
                        onClick={handleCreate}
                        disabled={creating}
                      >
                        {creating ? "Creating..." : "Create Konect"}
                      </button>
                    </div>
                  </>
                ) : selectedWorkflow ? (
                  <>
                    <div className="panel-header">
                      <div>
                        <div className="panel-kicker">Konect</div>
                        <h2>{selectedWorkflow?.name}</h2>
                        <div className="panel-meta-row">
                          <span className="panel-chip">
                            {selectedWorkflow?.trigger_type}
                          </span>
                          <span
                            className={`panel-chip ${
                              selectedWorkflow?.enabled ? "active" : ""
                            }`}
                          >
                            {selectedWorkflow?.trigger_type === "manual"
                              ? "manual"
                              : selectedWorkflow?.enabled
                              ? "active"
                              : "paused"}
                          </span>
                        </div>
                      </div>
                      <button
                        className="panel-close"
                        onClick={() => {
                          setSelectedWorkflow(null);
                          setPanelOpen(false);
                        }}
                      >
                        x
                      </button>
                    </div>
                    <div className="panel-body">
                      <div className="panel-section">
                        <div className="panel-section-title">Payload preview</div>
                        <label className="field">
                          <span>Payload</span>
                          <textarea
                            className="payload-area"
                            value={payloadPreview}
                            onChange={(e) => setPayloadPreview(e.target.value)}
                            rows={8}
                          />
                        </label>
                      </div>
                      <div className="muted">
                        Edit the payload before triggering or saving updates.
                      </div>
                    </div>
                    <div className="panel-actions">
                      {selectedWorkflow?.trigger_type !== "manual" ? (
                        <button
                          className="primary-btn"
                          onClick={handleToggleTimer}
                          disabled={togglingTimer}
                        >
                          {togglingTimer
                            ? "Applying..."
                            : selectedWorkflow?.enabled
                            ? "Pause Konect"
                            : "Start Konect"}
                        </button>
                      ) : (
                        <button
                          className="primary-btn"
                          onClick={handleTrigger}
                          disabled={triggering}
                        >
                          {triggering ? "Triggering..." : "Trigger now"}
                        </button>
                      )}
                      <button
                        className="danger-btn"
                        onClick={handleDelete}
                        disabled={deleting}
                      >
                        {deleting ? "Deleting..." : "Delete Konect"}
                      </button>
                    </div>
                  </>
                ) : (
                  <>
                    <div className="panel-header">
                      <div>
                        <div className="panel-kicker">Konect</div>
                        <h2>No selection</h2>
                        <div className="panel-meta-row">
                          <span className="panel-chip">Choose a Konect</span>
                        </div>
                      </div>
                      <button
                        className="panel-close"
                        onClick={() => {
                          setSelectedWorkflow(null);
                          setPanelOpen(false);
                        }}
                      >
                        x
                      </button>
                    </div>
                    <div className="panel-body">
                      <div className="muted">
                        Select a Konect from the list or create a new one.
                      </div>
                    </div>
                  </>
                )}
              </div>
            </aside>
          </>
        )}
      </div>
    </div>
  );
}
