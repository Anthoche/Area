import React from "react";
import "./konectcard.css";

export default function KonectCard({title, desc, services, type, isActive}) {
    const servicesList = services.map((s) => <li className="konect-card-inf konect-card-service">{s}</li>);
    const konectType = type === "timer" ? "Timer" : "Manual";
    const konectActive = isActive ? <div className='konect-card-inf konect-status active'>Active</div> :
        <div className='konect-card-inf konect-status paused'>Paused</div>;

    return (
        <div className="konect-card">
            <div className="konect-card-header">
                <h3 className="konect-card-title">{title}</h3>
                {konectActive}
            </div>
            <div className="konect-card-body">
                <p>{desc}</p>
            </div>
            <div className="konect-card-footer">
                <ul className="konect-card-services">
                    {servicesList}
                </ul>
                <span className="konect-card-inf konect-type">{konectType}</span>
            </div>
        </div>
    );
}
