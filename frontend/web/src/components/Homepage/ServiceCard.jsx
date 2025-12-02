import React from "react";
import "./homepage.css";

export default function ServiceCard({ title, color = "#ddd", icons = [] }) {
  return (
    <div className="service-card" style={{ background: color }}>
      <div className="service-card-header">
        <div className="service-icons">
          {icons.map((ic, i) => (
            <span key={i} className="service-icon">
              {ic}
            </span>
          ))}
        </div>
      </div>
      <div className="service-card-body">
        <div className="service-title">{title}</div>
      </div>
    </div>
  );
}
