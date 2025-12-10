import React from "react";
import "./servicecard.css";

export default function ServiceCard({ title, color = "#ddd", icons = [], onClick, ghost }) {
  const safeIcons = Array.isArray(icons) ? icons : [];
  return (
    <div
      className={`service-card ${ghost ? "ghost-card" : ""}`}
      style={{ background: color, cursor: "pointer" }}
      onClick={onClick}
    >
      <div className="service-card-header">
        <div className="service-icons">
          {safeIcons.map((ic, i) => (
            <span key={i} className="service-icon">{ic}</span>
          ))}
        </div>
      </div>
      <div className="service-card-body">
        <div className="service-title">{title}</div>
      </div>
    </div>
  );
}
