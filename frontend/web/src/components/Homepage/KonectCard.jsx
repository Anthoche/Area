import React from "react";
import "./konectcard.css";

export default function KonectCard({ title, desc, services, type, isActive, onClick }) {
    const servicesList = services.map((s) => (
        <li className="konect-card-inf konect-card-service" key={"t-" + title + "-s-" + s}>
            {s}
        </li>
    ));
    const typeLabel =
        type === "timer" ? "Timer" : type === "manual" ? "Manual" : type || "Manual";
    const konectActive = isActive ? (
        <div className="konect-card-inf konect-status active">Active</div>
    ) : (
        <div className="konect-card-inf konect-status paused">Paused</div>
    );

    return (
        <button className="konect-card" type="button" onClick={onClick}>
            <div className="konect-card-header">
                <h3 className="konect-card-title">{title}</h3>
                {konectActive}
            </div>
            <div className="konect-card-body">
                <p>{desc}</p>
            </div>
            <div className="konect-card-footer">
                <ul className="konect-card-services">{servicesList}</ul>
                <span className="konect-card-inf konect-type">{typeLabel}</span>
            </div>
        </button>
    );
}
