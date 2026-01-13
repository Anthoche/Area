/**
 * @file FilterTag.jsx
 * @description
 * Reusable UI component representing a selectable filter tag.
 * Commonly used in filtering interfaces such as:
 *  - Search filters
 *  - Category selectors
 *  - Tag-based navigation
 *
 * The component visually reflects its selected state and
 * delegates interaction handling to its parent.
 */

import React from "react";
import "./filtertag.css";

export default function FilterTag({label, selected, onClick}) {
    return (
        <button
            className={`filter-tag ${selected ? "active" : ""}`}
            type="button"
            onClick={onClick}
        >
            {label}
        </button>
    );
}
