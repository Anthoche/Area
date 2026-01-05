import React, { useEffect, useRef, useState } from "react";

export default function ServiceCard({ title, color = "#ddd", icons = [], onClick, ghost, created, deleted, onAnimationEnd }) {
    const safeIcons = Array.isArray(icons) ? icons : [];
    const [animate, setAnimate] = useState(false);
    const [exit, setExit] = useState(false);
    const cardRef = useRef(null);

    useEffect(() => {
        if (created) {
            setAnimate(true);
            const timeout = setTimeout(() => setAnimate(false), 700);
            return () => clearTimeout(timeout);
        }
    }, [created]);

    return (
        <div
            ref={cardRef}
            className={`service-card ${ghost ? "ghost-card" : ""} ${animate ? "service-card-appear" : ""} ${exit ? "service-card-exit" : ""}`}
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
