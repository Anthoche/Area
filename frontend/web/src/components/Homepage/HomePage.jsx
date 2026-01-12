import React, { useEffect, useMemo, useState } from "react";
import Navbar from "../Navbar.jsx";
import Footer from "../Footer.jsx";
import FilterTag from "./FilterTag.jsx";
import KonectCard from "./KonectCard.jsx";
import "./homepage.css";

const API_BASE =
  import.meta.env.VITE_API_URL ||
  import.meta.env.API_URL ||
  `${window.location.protocol}//${window.location.hostname}:8080`;

export default function HomePage() {
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
  const [activeTypeFilters, setActiveTypeFilters] = useState([]);
  const [activeServiceFilters, setActiveServiceFilters] = useState([]);
  const [form, setForm] = useState({
    name: "Mon Konect",
    triggerType: "",
    triggerValues: {},
    values: {},
  });

  const getUserId = () => Number(localStorage.getItem("user_id") || "");

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
    const githubEmail = params.get("github_email");
    const userIdFromQuery = params.get("user_id");
    if (tokenId && (googleEmail || params.get("google_email"))) {
      localStorage.setItem("google_token_id", tokenId);
    } else if (tokenId && (githubLogin || githubEmail)) {
      localStorage.setItem("github_token_id", tokenId);
    }
    if (googleEmail) {
      localStorage.setItem("google_email", googleEmail);
      localStorage.setItem("user_email", googleEmail);
    }
    if (githubLogin) {
      localStorage.setItem("github_login", githubLogin);
    }
    if (githubEmail) {
      localStorage.setItem("user_email", githubEmail);
    }
    if (userIdFromQuery) {
      localStorage.setItem("user_id", userIdFromQuery);
    }
    if (tokenId || googleEmail || githubLogin || githubEmail) {
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
      const triggerCaps = services.flatMap((s) =>
        (s.triggers || []).map((t) => ({
          id: t.id,
          name: t.name,
          description: t.description,
          fields: t.fields || [],
          service: s.name || s.id,
          serviceId: s.id,
        }))
      );
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
            serviceId: s.id,
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

  const buildPayloadForWorkflow = (wf) => {
    if (!wf) return {};
    const cfg = wf.trigger_config || {};
    const fromCfg = cfg.payload_template || cfg.payload;
    if (fromCfg && typeof fromCfg === "object") {
      return { ...fromCfg };
    }
    return { ...(form.values || {}) };
  };

  const buildIntervalPayload = () => {
    return form.values || {};
  };

  const buildActionUrl = () => {
    const actionUrl = selectedReactionDef?.action_url || "";
    if (actionUrl.startsWith("http")) return actionUrl;
    if (actionUrl.startsWith("/")) return `${API_BASE}${actionUrl}`;
    if (
      (selectedReaction || "").includes("webhook") &&
      form.values?.webhook_url
    ) {
      return form.values.webhook_url;
    }
    if (!actionUrl && form.values?.url) {
      return form.values.url;
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
      let payload = buildPayloadForWorkflow(selectedWorkflow);
      try {
        payload = JSON.parse(payloadPreview || "{}");
      } catch (err) {
        console.warn("invalid payload preview, using defaults", err);
      }
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

  const toggleTypeFilter = (value) => {
    setActiveTypeFilters((prev) =>
      prev.includes(value) ? prev.filter((v) => v !== value) : [...prev, value]
    );
  };

  const toggleServiceFilter = (value) => {
    setActiveServiceFilters((prev) =>
      prev.includes(value) ? prev.filter((v) => v !== value) : [...prev, value]
    );
  };

  const matchesFilters = (wf) => {
    if (!activeTypeFilters.length && !activeServiceFilters.length) return true;
    const trigger = triggers.find((t) => t.id === wf.trigger_type);
    const reaction = reactions.find((r) =>
      matchActionUrl(wf.action_url, r.action_url)
    );
    const serviceIds = [trigger?.serviceId, reaction?.serviceId].filter(Boolean);
    const typeOk =
      !activeTypeFilters.length || activeTypeFilters.includes(wf.trigger_type);
    const serviceOk =
      !activeServiceFilters.length ||
      serviceIds.some((id) => activeServiceFilters.includes(id));
    return typeOk && serviceOk;
  };

  const typeFiltersList = triggers.map((t) => ({
    value: t.id,
    label: t.name,
  }));

  const serviceFiltersList = areas
    .filter((s) => s.enabled !== false)
    .map((s) => ({
      value: s.id,
      label: s.name || s.id,
    }));


  const normalizedWorkflows = workflows
    .filter(matchesFilters)
    .filter((wf) =>
      (wf.name || "")
        .toLowerCase()
        .includes(searchTerm.trim().toLowerCase())
    )
    .map((wf) => {
      const trigger = triggers.find((t) => t.id === wf.trigger_type);
      const reaction = reactions.find((r) =>
        matchActionUrl(wf.action_url, r.action_url)
      );
      const services = [];
      if (trigger?.service) services.push(trigger.service);
      if (reaction?.service && reaction?.service !== trigger?.service) {
        services.push(reaction.service);
      }
      if (!services.length && wf.action_url?.startsWith("http")) {
        services.push("Webhook");
      }
      return {
        ...wf,
        displayType: trigger?.name || wf.trigger_type,
        description: `${trigger?.name || wf.trigger_type} → ${
          reaction?.name || "Reaction"
        }`,
        services: services.length ? services : ["Core"],
      };
    });

  return (
    <div className={`home-page-wrapper page-wrapper ${panelOpen ? "panel-open" : ""}`}>
      <Navbar />
      <div className="home-page-content">
        <div className="home-page-header home-page-section">
          <div className="home-page-header-text">
            <h1>Konects</h1>
            <span>Manage and automate your favorite services seamlessly.</span>
            <span>Create and organize your Konects to boost productivity.</span>
          </div>
          <div className="home-page-create-button">
            <button
              className="create-konect-btn"
              onClick={() => {
                setShowCreate(true);
                setPanelOpen(true);
                setSelectedWorkflow(null);
              }}
            >
              <svg
                xmlns="http://www.w3.org/2000/svg"
                width="20"
                height="20"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
                strokeLinecap="round"
                strokeLinejoin="round"
                className="lucide lucide-plus"
              >
                <path d="M5 12h14"></path>
                <path d="M12 5v14"></path>
              </svg>
              <span>Create a Konect</span>
            </button>
          </div>
        </div>
        <div className="search-section home-page-section">
          <div className="search-wrapper">
            <svg
              xmlns="http://www.w3.org/2000/svg"
              width="20"
              height="20"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
              className="lucide lucide-search search-icon"
            >
              <circle cx="11" cy="11" r="8"></circle>
              <path d="m21 21-4.3-4.3"></path>
            </svg>
            <input
              type="text"
              placeholder="Search konects..."
              className="search-input"
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
            />
          </div>
        </div>
        <div className="konects-filter home-page-section">
          <div className="filter-section">
            <h3 className="filter-title">Type</h3>
            <ul className="filter-buttons">
              {typeFiltersList.map((tag) => (
                <li key={tag.value}>
                  <FilterTag
                    label={tag.label}
                    selected={activeTypeFilters.includes(tag.value)}
                    onClick={() => toggleTypeFilter(tag.value)}
                  />
                </li>
              ))}
            </ul>
          </div>
          <div className="filter-section">
            <h3 className="filter-title">Services</h3>
            <ul className="filter-buttons">
              {serviceFiltersList.map((tag) => (
                <li key={tag.value}>
                  <FilterTag
                    label={tag.label}
                    selected={activeServiceFilters.includes(tag.value)}
                    onClick={() => toggleServiceFilter(tag.value)}
                  />
                </li>
              ))}
            </ul>
          </div>
        </div>
        <div className="konects">
          <h2>My Konects</h2>
          <ul className="konects-list">
            {normalizedWorkflows.map((wf) => (
              <li key={wf.id}>
                <KonectCard
                  title={wf.name}
                  desc={wf.description}
                  type={wf.displayType}
                  services={wf.services}
                  isActive={wf.enabled}
                  onClick={() => {
                    setSelectedWorkflow(wf);
                    setPanelOpen(true);
                    setShowCreate(false);
                  }}
                />
              </li>
            ))}
            {!normalizedWorkflows.length && (
              <li className="muted">No Konect created yet. Create the first one!</li>
            )}
          </ul>
        </div>
      </div>
      <Footer />
      <button
        className="refresh-button"
        title="Refresh konects"
        onClick={() => fetchWorkflows()}
        disabled={loading}
      >
        <svg
          xmlns="http://www.w3.org/2000/svg"
          width="24"
          height="24"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="2"
          strokeLinecap="round"
          strokeLinejoin="round"
          className="lucide lucide-rotate-cw"
        >
          <path d="M21 12a9 9 0 1 1-9-9c2.52 0 4.93 1 6.74 2.74L21 8"></path>
          <path d="M21 3v5h-5"></path>
        </svg>
      </button>

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
  );
}

function matchActionUrl(workflowUrl, reactionUrl) {
  if (!workflowUrl || !reactionUrl) return false;
  if (workflowUrl === reactionUrl) return true;
  if (reactionUrl.startsWith("/") && workflowUrl.endsWith(reactionUrl)) {
    return true;
  }
  return false;
}
